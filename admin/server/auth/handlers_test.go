package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

	return &Authenticator{
		logger:  zap.NewNop(),
		admin:   &admin.Service{URLs: urls},
		cookies: cookies.New(zap.NewNop(), []byte("0123456789abcdef0123456789abcdef"), []byte("0123456789abcdef")),
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
