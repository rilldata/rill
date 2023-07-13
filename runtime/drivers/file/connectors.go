package file

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type config struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}

func parseConfig(props map[string]any) (*config, error) {
	conf := &config{}
	err := mapstructure.Decode(props, &conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func init() {
	drivers.RegisterConnector("local_file", &connection{})
}

var spec = drivers.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path or URL to file",
			Placeholder: "/path/to/file",
		},
		{
			Key:         "format",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Format",
			Description: "Either CSV or Parquet. Inferred if not set.",
			Placeholder: "csv",
		},
	},
}

// ConnectorSpec implements drivers.Connection.
func (c *connection) Spec() drivers.Spec {
	return spec
}

func (c *connection) HasAnonymousAccess(ctx context.Context, props map[string]any) (bool, error) {
	return true, nil
}

// FilePaths implements drivers.FileStore
func (c *connection) FilePaths(ctx context.Context, src *drivers.FileSource) ([]string, error) {
	conf, err := parseConfig(src.Properties)
	if err != nil {
		return nil, err
	}

	path, err := c.resolveLocalPath(conf.Path)
	if err != nil {
		return nil, err
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("file does not exist at %s", conf.Path)
	}

	return localPaths, nil
}

func (c *connection) resolveLocalPath(path string) (string, error) {
	path, err := fileutil.ExpandHome(path)
	if err != nil {
		return "", err
	}

	finalPath := path
	if !filepath.IsAbs(path) {
		finalPath = filepath.Join(c.root, path)
	}

	allowHostAccess := false
	if val, ok := c.config["allow_host_access"].(bool); ok {
		allowHostAccess = val
	}
	if !allowHostAccess && !strings.HasPrefix(finalPath, c.root) {
		// path is outside the repo root
		return "", fmt.Errorf("file connector cannot ingest source: path is outside repo root")
	}
	return finalPath, nil
}
