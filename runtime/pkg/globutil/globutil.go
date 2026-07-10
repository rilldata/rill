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
func ParseBucketURL(input string) (*URL, error) {
	return parseBucketURL(input, false)
}

// ParseBucketURLLenient parses a URL similar to ParseBucketURL, but is tolerant of inputs
// that do not include a scheme (i.e., missing "://").
// In such cases, the entire input is treated as the path, and Scheme and Host are returned as empty strings.
func ParseBucketURLLenient(input string) (*URL, error) {
	return parseBucketURL(input, true)
}

func parseBucketURL(input string, lenient bool) (*URL, error) {
	scheme, path, ok := strings.Cut(input, "://")
	if !ok {
		if lenient {
			return &URL{Scheme: "", Host: "", Path: input}, nil
		}
		return nil, fmt.Errorf("failed to parse URL '%q'", input)
	}

	host, path, ok := strings.Cut(path, "/")
	if !ok {
		if lenient {
			return &URL{Scheme: "", Host: "", Path: input}, nil
		}
		// This is actually a valid URL, just not a valid object storage URL.
		return nil, fmt.Errorf("failed to parse URL '%q'", path)
	}

	return &URL{Scheme: scheme, Host: host, Path: path}, nil
}
