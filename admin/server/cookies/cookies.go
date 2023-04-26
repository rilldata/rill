package cookies

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type CookieStore struct {
	*sessions.CookieStore
	logger *zap.Logger
}

func NewCookieStore(logger *zap.Logger, keyPairs ...[]byte) *CookieStore {
	return &CookieStore{
		CookieStore: sessions.NewCookieStore(keyPairs...),
		logger:      logger,
	}
}

// getCookie returns a session for the given name.
// It returns a new session if the cookie could not be decoded and validated.
func (c *CookieStore) Get(r *http.Request, name string) *sessions.Session {
	sess, err := c.CookieStore.Get(r, name)
	if err != nil {
		c.logger.Error("failed to get cookie", zap.String("cookie_name", name), zap.Error(err), observability.ZapCtx(r.Context()))
		var cookieErr securecookie.Error
		if errors.As(err, &cookieErr) && cookieErr.IsDecode() {
			session := sessions.NewSession(c.CookieStore, name)
			session.Options = c.CookieStore.Options
			session.IsNew = true
			return session
		}
		panic(fmt.Errorf("unable to get cookie name %s error: %w", name, err))
	}
	return sess
}
