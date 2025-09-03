package cookies

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type Store struct {
	*sessions.CookieStore
	logger *zap.Logger
}

func New(logger *zap.Logger, keyPairs ...[]byte) *Store {
	return &Store{
		CookieStore: sessions.NewCookieStore(keyPairs...),
		logger:      logger,
	}
}

// Get returns a session for the given name.
// It returns a new session if the cookie could not be decoded and validated.
func (s *Store) Get(r *http.Request, name string) *sessions.Session {
	sess, err := s.CookieStore.Get(r, name)
	if err != nil {
		var cookieErr securecookie.Error
		if errors.As(err, &cookieErr) && cookieErr.IsDecode() {
			if strings.Contains(err.Error(), "expired timestamp") {
				s.logger.Info("cookie expired", zap.String("cookie_name", name), zap.Error(err), observability.ZapCtx(r.Context()))
			} else {
				s.logger.Error("failed to get cookie", zap.String("cookie_name", name), zap.Error(err), observability.ZapCtx(r.Context()))
			}
			session := sessions.NewSession(s.CookieStore, name)
			session.Options = s.CookieStore.Options
			session.IsNew = true
			return session
		}
		panic(fmt.Errorf("unable to get cookie name %q error: %w", name, err))
	}
	return sess
}
