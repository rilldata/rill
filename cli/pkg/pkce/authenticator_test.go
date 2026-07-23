package pkce

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/stretchr/testify/require"
)

func TestGenerateCodeVerifier(t *testing.T) {
	t.Parallel()
	// PKCE verifiers must satisfy RFC 7636 length and alphabet constraints on every generation.

	for i := 0; i < 1000; i++ {
		code, err := generateCodeVerifier()
		require.NoError(t, err)
		require.NotEmpty(t, code)
		require.GreaterOrEqual(t, len(code), 43)
		require.LessOrEqual(t, len(code), 128)
		// only contains A-Z, a-z, 0-9, and the punctuation characters -._~ (hyphen, period, underscore, and tilde)
		for _, c := range code {
			require.Contains(t, charset, string(c))
		}
	}
}

func TestGetAuthURLContract(t *testing.T) {
	t.Parallel()
	// The browser URL is the CLI/server contract, including encoded state and the S256 challenge.

	authenticator := &Authenticator{
		baseAuthURL:  "https://auth.example",
		redirectURL:  "http://localhost:9009/auth/callback?open=true",
		codeVerifier: "0123456789abcdefghijklmnopqrstuvwxyz-._~ABCDEFG",
		clientID:     "rill-local",
	}
	authURL, err := url.Parse(authenticator.GetAuthURL("state with spaces"))
	require.NoError(t, err)
	require.Equal(t, "https", authURL.Scheme)
	require.Equal(t, "auth.example", authURL.Host)
	require.Equal(t, "/auth/oauth/authorize", authURL.Path)
	require.Equal(t, url.Values{
		"client_id":             {"rill-local"},
		"redirect_uri":          {"http://localhost:9009/auth/callback?open=true"},
		"response_type":         {"code"},
		"code_challenge":        {createCodeChallenge(authenticator.codeVerifier)},
		"code_challenge_method": {codeChallengeMethod},
		"state":                 {"state with spaces"},
	}, authURL.Query())
}

func TestExchangeCodeForTokenContract(t *testing.T) {
	t.Parallel()
	// Exercise successful and malformed token responses while proving failed exchanges do not replace credentials.

	tests := []struct {
		name          string
		status        int
		body          string
		wantToken     string
		wantErrText   string
		wantPersisted string
	}{
		{
			name:          "success",
			status:        http.StatusOK,
			body:          `{"access_token":"new-token","token_type":"Bearer"}`,
			wantToken:     "new-token",
			wantPersisted: "new-token",
		},
		{
			name:          "malformed JSON",
			status:        http.StatusOK,
			body:          `{"access_token":`,
			wantErrText:   "unexpected EOF",
			wantPersisted: "old-token",
		},
		{
			name:          "empty token response",
			status:        http.StatusOK,
			body:          `{}`,
			wantErrText:   "missing access_token",
			wantPersisted: "old-token",
		},
		{
			name:          "server error",
			status:        http.StatusInternalServerError,
			body:          "backend unavailable",
			wantErrText:   "unexpected status code: 500",
			wantPersisted: "old-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var persisted atomic.Value
			persisted.Store("old-token")
			var changedBeforeSuccess atomic.Bool
			requests := make(chan pkceRequestSnapshot, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if persisted.Load().(string) != "old-token" {
					changedBeforeSuccess.Store(true)
				}
				requests <- snapshotPKCERequest(r)
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			t.Cleanup(server.Close)

			authenticator, err := NewAuthenticator(
				server.URL,
				"http://localhost:9009/auth/callback?open=true",
				"rill-local",
				"/project",
			)
			require.NoError(t, err)
			authenticator.client = server.Client()
			authenticator.codeVerifier = "0123456789abcdefghijklmnopqrstuvwxyz-._~ABCDEFG"

			token, err := authenticator.ExchangeCodeForToken("authorization-code")
			// This mirrors the callback persistence boundary: the old credential is
			// retained unless a complete exchange returns a non-empty token.
			if err == nil {
				persisted.Store(token)
			}

			if tt.wantErrText == "" {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, token)
			} else {
				require.ErrorContains(t, err, tt.wantErrText)
				require.Empty(t, token)
			}
			require.Equal(t, tt.wantPersisted, persisted.Load())
			require.False(t, changedBeforeSuccess.Load())

			request := <-requests
			require.NoError(t, request.parseErr)
			require.Equal(t, http.MethodPost, request.method)
			require.Equal(t, "/auth/oauth/token", request.path)
			require.Equal(t, oauth.FormMediaType, request.contentType)
			require.Equal(t, oauth.JSONMediaType, request.accept)
			require.Equal(t, url.Values{
				"grant_type":             {"authorization_code"},
				"code":                   {"authorization-code"},
				"client_id":              {"rill-local"},
				"redirect_uri":           {"http://localhost:9009/auth/callback?open=true"},
				"code_verifier":          {authenticator.codeVerifier},
				"token_response_version": {"standard"},
			}, request.form)
		})
	}
}

func TestExchangeCodeForTokenCancellation(t *testing.T) {
	t.Parallel()
	// Canceling an in-flight HTTP exchange must promptly return without producing a token.

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(releaseRequest) })

	authenticator, err := NewAuthenticator(server.URL, "http://localhost/callback", "rill-local", "/")
	require.NoError(t, err)
	authenticator.client = server.Client()

	ctx, cancel := context.WithCancel(context.Background())
	type result struct {
		token string
		err   error
	}
	resultCh := make(chan result, 1)
	go func() {
		token, err := authenticator.ExchangeCodeForTokenContext(ctx, "authorization-code")
		resultCh <- result{token: token, err: err}
	}()

	<-requestStarted
	cancel()
	select {
	case result := <-resultCh:
		require.ErrorIs(t, result.err, context.Canceled)
		require.Empty(t, result.token)
	case <-time.After(2 * time.Second):
		t.Fatal("PKCE token exchange did not stop after cancellation")
	}
}

type pkceRequestSnapshot struct {
	method      string
	path        string
	contentType string
	accept      string
	form        url.Values
	parseErr    error
}

func snapshotPKCERequest(r *http.Request) pkceRequestSnapshot {
	err := r.ParseForm()
	return pkceRequestSnapshot{
		method:      r.Method,
		path:        r.URL.Path,
		contentType: r.Header.Get("Content-Type"),
		accept:      r.Header.Get("Accept"),
		form:        r.PostForm,
		parseErr:    err,
	}
}
