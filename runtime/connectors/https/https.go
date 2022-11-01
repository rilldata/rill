package https

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
)

func init() {
	connectors.Register("https", connector{})
}

var spec = connectors.Spec{
	DisplayName: "http(s)",
	Description: "Connect to a remote file.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Path",
			Description: "Path to the remote file.",
			Placeholder: "https://example.com/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
	},
}

type Config struct {
	Path string `mapstructure:"path"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFile(ctx context.Context, source *connectors.Source, callback func(filename string) error) error {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	extension, err := getUrlExtension(conf.Path)
	if err != nil {
		return fmt.Errorf("failed to parse path %s, %v", conf.Path, err)
	}

	resp, err := http.Get(conf.Path)
	if err != nil {
		return fmt.Errorf("failed to fetch url %s:  %v", conf.Path, err)
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf("%s*.%s", source.Name, extension),
	)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	if err != nil {
		return err
	}
	io.Copy(f, resp.Body)
	callback(f.Name())

	return nil
}

func getUrlExtension(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	p := strings.Split(u.Path, ".")

	return p[len(p)-1], nil
}
