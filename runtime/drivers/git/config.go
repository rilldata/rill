package git

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/go-github/v50/github"
)

const (
	_githubAppKey         = "github_app_key"
	_githubInstallationID = "github_installation_id"
	_githubAppID          = "github_app_id"
)

func parseDSN(dsn string) (string, *InstallationAuth, error) {
	// Parse DSN as URL
	uri, err := url.Parse(dsn)
	if err != nil {
		return "", nil, fmt.Errorf("could not parse dsn: %w", err)
	}
	qry, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return "", nil, fmt.Errorf("could not parse dsn: %w", err)
	}

	var auth *InstallationAuth
	if qry.Has(_githubInstallationID) {
		id, err := strconv.ParseInt(qry.Get(_githubInstallationID), 10, 64)
		if err != nil {
			return "", nil, fmt.Errorf("invalid github installation id %v ", qry.Get(_githubInstallationID))
		}

		f, _ := os.Open("/Users/kanshul/Downloads/test-rill-webhooks.2023-02-27.private-key.pem")
		content, _ := io.ReadAll(f)
		auth, err = NewInstallationAuth(id, content, "299102")
		if err != nil {
			return "", nil, err
		}

		// Remove from query string (so not passed into github url)
		qry.Del(_githubInstallationID)
		// update url with installation token to access repos
		// uri.User = url.UserPassword("x-access-token", installationToken.GetToken())
	}

	// Rebuild DSN
	uri.RawQuery = qry.Encode()
	dsn = uri.String()
	return dsn, auth, nil
}

// InstallationAuth represent a HTTP basic auth
type InstallationAuth struct {
	token                *github.InstallationToken
	githubInstallationID int64
	appKey               []byte
	githubAppID          string
}

func NewInstallationAuth(installationID int64, appKey []byte, githubAppID string) (*InstallationAuth, error) {
	token, err := NewInstallationToken(installationID, appKey, githubAppID)
	if err != nil {
		return nil, err
	}

	return &InstallationAuth{token: token, githubInstallationID: installationID, appKey: appKey, githubAppID: githubAppID}, nil
}

func (a *InstallationAuth) SetAuth(r *http.Request) {
	if a == nil {
		return
	}

	if a.token.ExpiresAt.Before(time.Now().Add(time.Minute)) {
		newToken, err := NewInstallationToken(a.githubInstallationID, a.appKey, a.githubAppID)
		if err != nil {
			return
		}
		a.token = newToken
	}
	r.Header.Set("Authorization", "Basic "+a.token.GetToken())
}

// Name is name of the auth
func (a *InstallationAuth) Name() string {
	return "http-token-auth"
}

func (a *InstallationAuth) String() string {
	masked := "*******"
	if a.token == nil {
		masked = "<empty>"
	}

	return fmt.Sprintf("%s - %s", a.Name(), masked)
}
