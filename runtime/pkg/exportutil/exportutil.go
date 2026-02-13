package exportutil

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

func ParsePath(path string) (bucket, key, fullPath string, err error) {
	uri, err := url.Parse(path)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse path %q: %w", path, err)
	}

	if uri.Scheme != "s3" && uri.Scheme != "gs" && uri.Scheme != "gcs" {
		return "", "", "", fmt.Errorf("unsupported object store schema %q", uri.Scheme)
	}

	// Trim the leading slash from the path to get the key, since S3 keys should not start with a slash
	key = strings.TrimPrefix(uri.Path, "/")

	// Check for file extension, to determine if the path is a directory or a file.
	// If it's a directory, we will add a timestamped filename with .parquet extension
	ext := filepath.Ext(uri.Path)
	if ext == "" {
		key, err = url.JoinPath(key, fmt.Sprintf("%s.parquet", time.Now().Format(time.RFC3339)))
		if err != nil {
			return "", "", "", err
		}
	} else if ext != ".parquet" {
		return "", "", "", fmt.Errorf("invalid path %q: file extension must be .parquet", path)
	}

	// Construct full path
	fullPath = uri.Scheme + "://" + uri.Host + "/" + key

	return uri.Host, key, fullPath, nil
}
