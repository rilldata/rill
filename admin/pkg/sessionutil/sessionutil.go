package sessionutil

import (
	"errors"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// GetCookie returns a session for the given name.
// It returns a new session if the cookie could not be decoded and validated.
func GetCookie(cookies *sessions.CookieStore, r *http.Request, name string) (*sessions.Session, error) {
	sess, err := cookies.Get(r, name)
	if err != nil {
		var cookieErr securecookie.Error
		if errors.As(err, &cookieErr) && cookieErr.IsDecode() {
			session := sessions.NewSession(cookies, name)
			session.Options = cookies.Options
			session.IsNew = true
			return session, nil
		}
		return nil, err
	}
	return sess, nil
}
