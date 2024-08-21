package admin

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rilldata/rill/admin/pkg/urlutil"
)

// URLs centralizes parsing and formatting of URLs for the admin service.
//
// There are several complexities around URL handling in Rill:
// 1. The frontend may run on a different host than the admin service (e.g. ui.rilldata.com vs. admin.rilldata.com).
// 2. We support custom domains for specific orgs (e.g. analytics.mycompany.com instead of ui.rilldata.com/mycompany).
// 3. The admin service sends transactional emails that link to the frontend, such as project invites.
// 4. The admin service is also responsible for sending transactional emails on behalf of the runtime, which also link to the frontend, such as for alerts and reports.
// 5. We need to ensure correct redirects and callbacks for the auth service (on auth.rilldata.com) and Github.
//
// For orgs with a custom domain configured (using the CLI command `rill sudo org set-custom-domain`),
// the admin service and frontend must be reachable on the custom domain using the following load balancer rules:
// 1. The admin service must be reachable at the `/api` path prefix on the custom domain. The `/api` prefix should be removed by the load balancer before proxying to the admin service.
// 2. The frontend must be reachable at all other paths on the custom domain.
type URLs struct {
	externalURL    *url.URL
	externalURLRaw string
	frontendURL    *url.URL
	frontendURLRaw string
}

// NewURLs creates a new URLs. The provided URLs should include the scheme, host, optional port, and optional path prefix.
// The provided URLs should be the primary external and frontend URL for the Rill service. The returned *URLs will rewrite them as needed for custom domains.
func NewURLs(externalURL, frontendURL string) (*URLs, error) {
	eu, err := url.Parse(externalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse external URL: %w", err)
	}

	fu, err := url.Parse(frontendURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontend URL: %w", err)
	}

	return &URLs{
		externalURL:    eu,
		externalURLRaw: externalURL,
		frontendURL:    fu,
		frontendURLRaw: frontendURL,
	}, nil
}

// IsHTTPS returns true if the admin service's external URL uses HTTPS.
func (u *URLs) IsHTTPS() bool {
	return u.externalURL.Scheme == "https"
}

// ExternalURL returns the external URL for the admin service.
func (u *URLs) External() string {
	return u.externalURLRaw
}

// FrontendURL returns the frontend URL for the admin service.
func (u *URLs) Frontend() string {
	return u.frontendURLRaw
}

// AuthLogin returns the URL that starts the redirects to the auth service for login.
func (u *URLs) AuthLogin(redirect string) string {
	res := urlutil.MustJoinURL(u.externalURLRaw, "/auth/login")
	if redirect != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"redirect": redirect})
	}
	return res
}

// AuthLoginCallback returns the URL for the OAuth2 callback.
func (u *URLs) AuthLoginCallback() string {
	return urlutil.MustJoinURL(u.externalURLRaw, "/auth/callback")
}

// AuthLogoutCallback returns the URL for the logout callback.
func (u *URLs) AuthLogoutCallback() string {
	return urlutil.MustJoinURL(u.externalURLRaw, "/auth/logout/callback")
}

// VerifyEmailUI returns the frontend URL for the verify email page.
func (u *URLs) VerifyEmailUI() string {
	return urlutil.MustJoinURL(u.frontendURLRaw, "/-/auth/verify-email")
}

// DeviceAuthVerification returns the frontend URL for the device auth verification page.
func (u *URLs) DeviceAuthVerification(query map[string]string) string {
	return urlutil.MustWithQuery(urlutil.MustJoinURL(u.frontendURLRaw, "/-/auth/device"), query)
}

// CLIAuthSuccess returns the frontend URL to redirect to after successful CLI authentication.
func (u *URLs) CLIAuthSuccess() string {
	return urlutil.MustJoinURL(u.frontendURLRaw, "/-/auth/cli/success")
}

// GithubConnect returns the URL that starts the Github connect redirects.
func (u *URLs) GithubConnect(remote string) string {
	res := urlutil.MustJoinURL(u.externalURLRaw, "/github/connect")
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubAuthLogin returns the URL that starts the Github auth redirects.
func (u *URLs) GithubAuthLogin(remote string) string {
	res := urlutil.MustJoinURL(u.externalURLRaw, "/github/auth/login")
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubAuthCallback returns the URL for the Github auth callback.
func (u *URLs) GithubAuthCallback() string {
	return urlutil.MustJoinURL(u.externalURLRaw, "/github/auth/callback")
}

// GithubConnectUI returns the page in the Rill frontend for starting the Github connect flow.
func (u *URLs) GithubConnectUI(redirect string) string {
	res := urlutil.MustJoinURL(u.frontendURLRaw, "/-/github/connect")
	if redirect != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"redirect": redirect})
	}
	return res
}

// GithubConnectRetryUI returns the page in the Rill frontend for retrying the Github connect flow.
func (u *URLs) GithubConnectRetryUI(remote string) string {
	res := urlutil.MustJoinURL(u.frontendURLRaw, "/-/github/connect/retry-install")
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubConnectRequestUI returns the page in the Rill frontend for requesting a Github connect.
func (u *URLs) GithubConnectRequestUI(remote string) string {
	res := urlutil.MustJoinURL(u.frontendURLRaw, "/-/github/connect/request")
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	return res
}

// GithubConnectSuccessUI returns the page in the Rill frontend for a successful Github connect.
func (u *URLs) GithubConnectSuccessUI(autoclose bool) string {
	res := urlutil.MustJoinURL(u.frontendURLRaw, "/-/github/connect/success")
	if autoclose {
		res = urlutil.MustWithQuery(res, map[string]string{"autoclose": "true"})
	}
	return res
}

// GithubRetryAuthUI returns the page in the Rill frontend for retrying the Github auth flow.
func (u *URLs) GithubRetryAuthUI(remote, username string) string {
	res := urlutil.MustJoinURL(u.frontendURLRaw, "/-/github/connect/retry-auth")
	if remote != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"remote": remote})
	}
	if username != "" {
		res = urlutil.MustWithQuery(res, map[string]string{"githubUsername": username})
	}
	return res
}

// Project returns the URL for a project in the frontend.
func (u *URLs) Project(org, project string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project)
}

// Embed creates a URL for embedding the frontend in an iframe.
func (u *URLs) Embed(query map[string]string) (string, error) {
	return urlutil.WithQuery(urlutil.MustJoinURL(u.frontendURLRaw, "-", "embed"), query)
}

// MagicAuthTokenOpen returns the frontend URL for opening a magic auth token.
func (u *URLs) MagicAuthTokenOpen(org, project, token string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "share", token)
}

// ApproveProjectAccess returns the frontend URL for approving a project access request.
func (u *URLs) ApproveProjectAccess(org, project, id string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "request-access", id, "approve")
}

// DenyProjectAccess returns the frontend URL for denying a project access request.
func (u *URLs) DenyProjectAccess(org, project, id string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "request-access", id, "deny")
}

// ReportOpen returns the URL for opening a report in the frontend.
func (u *URLs) ReportOpen(org, project, report string, executionTime time.Time) string {
	reportURL := urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "reports", report, "open")
	reportURL += fmt.Sprintf("?execution_time=%s", executionTime.UTC().Format(time.RFC3339))
	return reportURL
}

// ReportExport returns the URL for exporting a report in the frontend.
func (u *URLs) ReportExport(org, project, report string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "reports", report, "export")
}

// ReportEdit returns the URL for editing a report in the frontend.
func (u *URLs) ReportEdit(org, project, report string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "reports", report)
}

// AlertOpen returns the URL for opening an alert in the frontend.
func (u *URLs) AlertOpen(org, project, alert string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "alerts", alert, "open")
}

// AlertEdit returns the URL for editing an alert in the frontend.
func (u *URLs) AlertEdit(org, project, alert string) string {
	return urlutil.MustJoinURL(u.frontendURLRaw, org, project, "-", "alerts", alert)
}
