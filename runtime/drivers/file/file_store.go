package file

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

// FilePaths implements drivers.FileStore
func (c *connection) FilePaths(ctx context.Context, src *drivers.FileSource) ([]string, error) {
	conf, err := parseSourceProperties(src.Properties)
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
	if val, ok := c.driverConfig["allow_host_access"].(bool); ok {
		allowHostAccess = val
	}
	if !allowHostAccess && !strings.HasPrefix(finalPath, c.root) {
		// path is outside the repo root
		return "", fmt.Errorf("file connector cannot ingest source: path is outside repo root")
	}
	return finalPath, nil
}
