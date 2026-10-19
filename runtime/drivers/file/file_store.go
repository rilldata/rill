package file

import (
	"context"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

// FilePaths implements drivers.FileStore
func (c *connection) FilePaths(ctx context.Context, src map[string]any) ([]string, func(), error) {
	conf, err := parseSourceProperties(src)
	if err != nil {
		return nil, nil, err
	}

	path, err := c.resolveLocalPath(conf.Path)
	if err != nil {
		return nil, nil, err
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return nil, nil, err
	}
	if len(localPaths) == 0 {
		return nil, nil, fmt.Errorf("file does not exist at %q", conf.Path)
	}

	return localPaths, func() {}, nil
}

func (c *connection) resolveLocalPath(path string) (string, error) {
	return fileutil.ResolveLocalPath(path, c.root, c.driverConfig.AllowHostAccess)
}
