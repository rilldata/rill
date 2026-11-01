package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/server/cookies"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConnectorOAuthFlowSession(t *testing.T) {
	urls, err := admin.NewURLs("https://admin.example.com", "https://ui.example.com")
	require.NoError(t, err)
	s := &Server{
		admin:   &admin.Service{URLs: urls},
		cookies: cookies.New(zap.NewNop(), []byte("0123456789abcdef0123456789abcdef")),
	}

	// Each flow gets its own cookie, scoped to the callback and expiring with the flow
	save := func(state, value string) *http.Cookie {
		w := httptest.NewRecorder()
		sess := s.connectorOAuthFlowSession(httptest.NewRequest(http.MethodGet, "/connectors/oauth/connect", nil), state)
		sess.Values["flow"] = value
		require.NoError(t, sess.Save(httptest.NewRequest(http.MethodGet, "/connectors/oauth/connect", nil), w))
		res := w.Result().Cookies()
		require.Len(t, res, 1)
		return res[0]
	}
	a := save("state-a", "flow-a")
	b := save("state-b", "flow-b")
	require.Equal(t, connectorOAuthCookieName+"_state-a", a.Name)
	require.Equal(t, "/connectors/oauth/callback", a.Path)
	require.Equal(t, int(connectorOAuthFlowTTL.Seconds()), a.MaxAge)
	require.NotEqual(t, a.Name, b.Name)

	// The callback finds each flow by its state
	r := httptest.NewRequest(http.MethodGet, "/connectors/oauth/callback", nil)
	r.AddCookie(a)
	r.AddCookie(b)
	require.Equal(t, "flow-a", s.connectorOAuthFlowSession(r, "state-a").Values["flow"])
	require.Equal(t, "flow-b", s.connectorOAuthFlowSession(r, "state-b").Values["flow"])
	require.Nil(t, s.connectorOAuthFlowSession(r, "state-c").Values["flow"])

	// Options are not shared with the store
	require.NotEqual(t, "/connectors/oauth/callback", s.cookies.Options.Path)
}
