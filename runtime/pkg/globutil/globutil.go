package globutil

import (
	"fmt"
	"strings"
)

type URL struct {
	Scheme string
	Host   string
	Path   string
}

// ParseBucketURL splits urls with globs into scheme, hostname and rest of the url(as glob)
// url.Parse removes `?` considering it is query param
// `?` is valid meta in glob patterns
func ParseBucketURL(path string) (*URL, error) {
	scheme, path, ok := strings.Cut(path, "://")
	if !ok {
		return nil, fmt.Errorf("failed to parse URL '%q'", path)
	}

	host, path, ok := strings.Cut(path, "/")
	if !ok {
		// This is actually a valid URL, just not a valid object storage URL.
		return nil, fmt.Errorf("failed to parse URL '%q'", path)
	}

	return &URL{Scheme: scheme, Host: host, Path: path}, nil
}
