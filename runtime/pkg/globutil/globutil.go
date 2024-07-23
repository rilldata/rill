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

func (u *URL) String() string {
	var res string
	if u.Scheme != "" && u.Host != "" {
		res = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	} else if u.Host != "" {
		res = u.Host
	}
	if u.Path != "" {
		if res == "" {
			res = u.Path
		} else {
			res += "/" + u.Path
		}
	}
	return res
}

// ParseBucketURL parses a URL while preserving glob patterns in the URL's path.
// For example, url.Parse removes `?` as a query string, which is problematic because `?` is a valid character in glob patterns.
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
