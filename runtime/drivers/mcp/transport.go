package mcp

import (
	"net/http"
	"sync/atomic"
)

// authRoundTripper injects the connector's headers and the caller's bearer token into every request,
// and remembers the status code of the last response so callers can recognize authorization failures
// that the MCP SDK only reports as generic errors.
type authRoundTripper struct {
	base       http.RoundTripper
	headers    map[string]string
	token      string
	lastStatus atomic.Int32
}

func (t *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	resp, err := t.base.RoundTrip(req)
	if resp != nil {
		t.lastStatus.Store(int32(resp.StatusCode))
	}
	return resp, err
}

// unauthorized returns true if the last response indicated an authorization failure.
func (t *authRoundTripper) unauthorized() bool {
	s := t.lastStatus.Load()
	return s == http.StatusUnauthorized || s == http.StatusForbidden
}
