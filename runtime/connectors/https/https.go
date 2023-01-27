package https

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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

func (c connector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source) (connectors.FileIterator, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	extension, err := urlExtension(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, conf.Path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url %s:  %w", conf.Path, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url %s:  %w", conf.Path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch url %s: %s", conf.Path, resp.Status)
	}

	file, err := fileutil.CopyToTempFile(resp.Body, source.Name, extension)
	if err != nil {
		return nil, err
	}
	return &iterator{ctx: ctx, files: []string{file}}, nil
}

// implements connector.FileIterator
type iterator struct {
	ctx   context.Context
	files []string
	index int
}

func (i *iterator) Close() error {
	fileutil.ForceRemoveFiles(i.files)
	return nil
}

func (i *iterator) NextBatch(n int) ([]string, error) {
	if !i.HasNext() {
		return nil, io.EOF
	}

	start := i.index
	end := i.index + n
	if end > len(i.files) {
		end = len(i.files)
	}
	return i.files[start:end], nil
}

func (i *iterator) HasNext() bool {
	return i.index < len(i.files)
}

func urlExtension(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return fileutil.FullExt(u.Path), nil
}
