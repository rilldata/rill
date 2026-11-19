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

// unauthorized returns true if the last response rejected the credentials (401), which a new authorization or refreshed token can fix.
// A 403 is not included: it means the credentials are valid but lack access (e.g. RFC 6750 "insufficient_scope"),
// which a refresh cannot fix, and treating it as an authorization failure would make callers discard a working authorization.
func (t *authRoundTripper) unauthorized() bool {
	return t.lastStatus.Load() == http.StatusUnauthorized
}

// rejected returns true if the last response was 401 or 403, i.e. the server is up but refused the request's credentials.
func (t *authRoundTripper) rejected() bool {
	s := t.lastStatus.Load()
	return s == http.StatusUnauthorized || s == http.StatusForbidden
}
