package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/server/cookies"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// newTestAuthenticator returns an Authenticator wired just enough to exercise the login and logout
// redirect handlers without a database or a live auth provider.
func newTestAuthenticator(t *testing.T, authDomain string) *Authenticator {
	urls, err := admin.NewURLs("http://localhost:8080", "http://localhost:3000")
	require.NoError(t, err)

	// Cookie options as set in server.New
	cookieStore := cookies.New(zap.NewNop(), []byte("0123456789abcdef0123456789abcdef"), []byte("0123456789abcdef"))
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.SameSite = http.SameSiteLaxMode

	return &Authenticator{
		logger:  zap.NewNop(),
		admin:   &admin.Service{URLs: urls},
		cookies: cookieStore,
		opts: &AuthenticatorOptions{
			AuthDomain:   authDomain,
			AuthClientID: "rill-client",
		},
		oauth2: oauth2.Config{
			ClientID:    "rill-client",
			RedirectURL: urls.AuthLoginCallback(),
			Endpoint:    oauth2.Endpoint{AuthURL: "https://idp.example.com/authorize"},
		},
	}
}

func TestAuthStartSignup(t *testing.T) {
	a := newTestAuthenticator(t, "idp.example.com")

	for _, signup := range []bool{false, true} {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/auth/login", nil)
		w := httptest.NewRecorder()
		a.authStart(w, req, signup)
		require.Equal(t, http.StatusTemporaryRedirect, w.Code)

		loc, err := url.Parse(w.Header().Get("Location"))
		require.NoError(t, err)
		q := loc.Query()
		if !signup {
			require.Empty(t, q.Get("prompt"))
			require.Empty(t, q.Get("screen_hint"))
			continue
		}
		// Auth0 only honors screen_hint, standard OIDC providers only honor prompt=create.
		require.Equal(t, "create", q.Get("prompt"))
		require.Equal(t, "signup", q.Get("screen_hint"))
	}
}

func TestAuthLogoutProvider(t *testing.T) {
	tests := []struct {
		name               string
		authDomain         string
		endSessionEndpoint string
		want               string
	}{
		{
			// Must stay identical to what Auth0 deployments get today.
			name:       "auth0 domain without end_session_endpoint",
			authDomain: "rill.auth0.com",
			want:       "https://rill.auth0.com/v2/logout?client_id=rill-client&returnTo=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Flogout%2Fcallback",
		},
		{
			name:       "issuer URL without end_session_endpoint",
			authDomain: "https://idp.example.com/realms/rill",
			want:       "http://localhost:8080/auth/logout/callback",
		},
		{
			name:               "issuer URL with end_session_endpoint",
			authDomain:         "https://idp.example.com/realms/rill",
			endSessionEndpoint: "https://idp.example.com/realms/rill/protocol/openid-connect/logout",
			want:               "https://idp.example.com/realms/rill/protocol/openid-connect/logout?client_id=rill-client&post_logout_redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Flogout%2Fcallback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAuthenticator(t, tt.authDomain)
			a.endSessionEndpoint = tt.endSessionEndpoint

			req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/auth/logout/provider", nil)
			w := httptest.NewRecorder()
			a.authLogoutProvider(w, req)
			require.Equal(t, http.StatusTemporaryRedirect, w.Code)
			require.Equal(t, tt.want, w.Header().Get("Location"))
		})
	}
}

func TestIDTokenHint(t *testing.T) {
	p := newTestProvider(t, nil)
	provider, err := oidc.NewProvider(context.Background(), p.URL)
	require.NoError(t, err)

	newAuthenticator := func(endSessionEndpoint string) *Authenticator {
		a := newTestAuthenticator(t, p.URL)
		a.oidc = provider
		a.endSessionEndpoint = endSessionEndpoint
		return a
	}
	idToken := func(key *rsa.PrivateKey, exp time.Time) string {
		return p.signIDToken(t, key, map[string]any{"iss": p.URL, "aud": "rill-client", "sub": "user", "iat": exp.Add(-5 * time.Minute).Unix(), "exp": exp.Unix()})
	}
	// save runs saveIDTokenHint as the login callback would and returns the cookie it set, if any.
	save := func(a *Authenticator, rawIDToken string) *http.Cookie {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/auth/callback", nil)
		w := httptest.NewRecorder()
		a.saveIDTokenHint(w, req, rawIDToken)
		require.Equal(t, http.StatusOK, w.Code) // It never fails the login
		for _, c := range w.Result().Cookies() {
			if c.Name == idTokenCookieName {
				return c
			}
		}
		return nil
	}
	// logout runs authLogoutProvider with the given cookie and returns the id_token_hint it sent and whether it cleared the cookie.
	logout := func(a *Authenticator, c *http.Cookie) (string, bool) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/auth/logout/provider", nil)
		req.AddCookie(c)
		w := httptest.NewRecorder()
		a.authLogoutProvider(w, req)
		require.Equal(t, http.StatusTemporaryRedirect, w.Code)
		loc, err := url.Parse(w.Header().Get("Location"))
		require.NoError(t, err)
		cleared := false
		for _, rc := range w.Result().Cookies() {
			if rc.Name == idTokenCookieName && rc.MaxAge < 0 && rc.Path == "/auth/logout/provider" {
				cleared = true
			}
		}
		return loc.Query().Get("id_token_hint"), cleared
	}

	t.Run("saved only for providers with an end_session_endpoint", func(t *testing.T) {
		require.Nil(t, save(newAuthenticator(""), idToken(nil, time.Now().Add(5*time.Minute))))
	})

	t.Run("scoped to the logout path", func(t *testing.T) {
		c := save(newAuthenticator(p.URL+"/logout"), idToken(nil, time.Now().Add(5*time.Minute)))
		require.NotNil(t, c)
		require.Equal(t, "/auth/logout/provider", c.Path)
		require.True(t, c.HttpOnly)
	})

	t.Run("skipped when too large for a cookie", func(t *testing.T) {
		require.Nil(t, save(newAuthenticator(p.URL+"/logout"), strings.Repeat("x", 5000)))
	})

	tests := []struct {
		name     string
		idToken  string
		wantHint bool
	}{
		{"valid", idToken(nil, time.Now().Add(5*time.Minute)), true},
		// The ID token expires within minutes but the Rill session lasts for weeks, and providers accept expired hints.
		{"expired", idToken(nil, time.Now().Add(-24*time.Hour)), true},
		// An invalid hint makes Keycloak fail the logout with an error page, so it is dropped.
		{"signed by another key", idToken(mustRSAKey(t), time.Now().Add(5*time.Minute)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAuthenticator(p.URL + "/logout")
			c := save(a, tt.idToken)
			require.NotNil(t, c)

			hint, cleared := logout(a, c)
			require.True(t, cleared)
			if tt.wantHint {
				require.Equal(t, tt.idToken, hint)
			} else {
				require.Empty(t, hint)
			}
		})
	}

	t.Run("undecodable cookie leaves the store options alone", func(t *testing.T) {
		a := newAuthenticator(p.URL + "/logout")
		before := *a.cookies.Options

		hint, cleared := logout(a, &http.Cookie{Name: idTokenCookieName, Value: "garbage"})
		require.Empty(t, hint)
		require.True(t, cleared)
		require.Equal(t, before, *a.cookies.Options)
	})
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func TestParseUserProfile(t *testing.T) {
	tests := []struct {
		name    string
		claims  map[string]any
		want    *userProfile
		wantErr string
	}{
		{
			name:   "all claims",
			claims: map[string]any{"email": "a@example.com", "email_verified": true, "name": "A", "picture": "https://example.com/a.png"},
			want:   &userProfile{email: "a@example.com", emailVerified: true, name: "A", photoURL: "https://example.com/a.png"},
		},
		{
			// Dex never emits picture, and Keycloak omits it for users without a picture attribute.
			name:   "without picture",
			claims: map[string]any{"email": "a@example.com", "email_verified": true, "name": "A"},
			want:   &userProfile{email: "a@example.com", emailVerified: true, name: "A"},
		},
		{
			name:   "email_verified as a string",
			claims: map[string]any{"email": "a@example.com", "email_verified": "false", "name": "A", "picture": ""},
			want:   &userProfile{email: "a@example.com", emailVerified: false, name: "A"},
		},
		{
			name:    "without email",
			claims:  map[string]any{"email_verified": true, "name": "A"},
			wantErr: "claim 'email' not found",
		},
		{
			name:    "without email_verified",
			claims:  map[string]any{"email": "a@example.com", "name": "A"},
			wantErr: "claim 'email_verified' not found",
		},
		{
			name:    "without name",
			claims:  map[string]any{"email": "a@example.com", "email_verified": true},
			wantErr: "claim 'name' not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUserProfile(tt.claims)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
