package globutil

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseURL splits urls with globs into scheme, hostname and rest of the url(as glob)
// url.Parse removes `?` considering it is query param
// `?` is valid meta in glob patterns
func ParseURL(path string) (string, string, string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", "", "", err
	}
	// everything after `Host/` is glob
	_, glob, found := strings.Cut(path, fmt.Sprintf("%s/", u.Host))
	if !found {
		return "", "", "", fmt.Errorf("failed to parse url %s", path)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", "", "", fmt.Errorf("invalid url %s", path)
	}

	return u.Scheme, u.Host, glob, nil
}
