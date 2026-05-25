package file

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var errProjectNotFound = errors.New("not connected to a rill project")

// loadGitConfig loads the git configuration for the repository
// Should be called with c.gitMu held.
func (c *connection) loadGitConfig(ctx context.Context) (*gitutil.Config, error) {
	if c.gitConfig != nil && !c.gitConfig.IsExpired() {
		return c.gitConfig, nil
	}

	// get authenticated admin client
	client, err := c.getAdminClient()
	if err != nil {
		return nil, err
	}

	// Build request
	req := &adminv1.ListProjectsForFingerprintRequest{
		DirectoryName: filepath.Base(c.root),
	}

	// extract subpath
	repoRoot, subpath, err := gitutil.InferRepoRootAndSubpath(c.root)
	if err == nil {
		req.SubPath = subpath
	}

	// extract remotes
	remote, err := gitutil.ExtractRemotes(repoRoot, false)
	if err == nil {
		for _, r := range remote {
			if r.Name == "__rill_remote" {
				req.RillMgdGitRemote = r.URL
			} else {
				gitRemote, err := r.Github()
				if err == nil {
					req.GitRemote = gitRemote
				}
			}
		}
	}
	resp, err := client.ListProjectsForFingerprint(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Projects) == 0 {
		return nil, errProjectNotFound
	}

	// filter by org
	org, err := c.dotRill.GetDefaultOrg()
	if err != nil {
		return nil, err
	}
	orgFiltered := make([]*adminv1.Project, 0)
	for _, p := range resp.Projects {
		if p.OrgName == org {
			orgFiltered = append(orgFiltered, p)
		}
	}
	if len(orgFiltered) == 0 {
		return nil, errProjectNotFound
	}
	p := orgFiltered[0]
	creds, err := client.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
		Org:     p.OrgName,
		Project: p.Name,
	})
	if err != nil {
		return nil, err
	}

	c.gitConfig = &gitutil.Config{
		Remote:            creds.GitRepoUrl,
		Username:          creds.GitUsername,
		Password:          creds.GitPassword,
		PasswordExpiresAt: creds.GitPasswordExpiresAt.AsTime(),
		DefaultBranch:     creds.GitPrimaryBranch,
		Subpath:           creds.GitSubpath,
		ManagedRepo:       creds.GitManagedRepo,
	}
	return c.gitConfig, nil
}

// gitSignature returns the git author name and email to use for commits.
// It checks the local git config first, then falls back to the logged-in Rill user.
func (c *connection) gitSignature(ctx context.Context, client *client.Client, path string) (name, email string, _ error) {
	n, e := gitReadUserConfig(path)
	if n != "" && e != "" {
		return n, e, nil
	}

	if client == nil {
		return "Rill", "noreply@rilldata.com", nil
	}
	userResp, err := client.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return "", "", err
	}
	return userResp.User.DisplayName, userResp.User.Email, nil
}

// gitReadUserConfig reads user.name and user.email from the git config for the given path.
// Returns ("", "") if not configured.
func gitReadUserConfig(repoDir string) (name, email string) {
	nameOut, err := exec.Command("git", "-C", repoDir, "config", "user.name").Output()
	if err == nil {
		name = strings.TrimSpace(string(nameOut))
	}
	emailOut, err := exec.Command("git", "-C", repoDir, "config", "user.email").Output()
	if err == nil {
		email = strings.TrimSpace(string(emailOut))
	}
	return name, email
}

func (c *connection) getAdminClient() (*client.Client, error) {
	if c.admin != nil {
		return c.admin, nil
	}
	accessToken, err := c.adminToken()
	if err != nil {
		return nil, err
	}
	if accessToken == "" {
		return nil, drivers.ErrNotAuthenticated
	}
	adminURL, err := c.adminURL()
	if err != nil {
		return nil, err
	}
	admin, err := client.New(adminURL, accessToken, "rill-runtime")
	if err != nil {
		return nil, err
	}
	c.admin = admin
	return c.admin, nil
}

func (c *connection) adminToken() (string, error) {
	if c.driverConfig.AccessTokenOverride != "" {
		return c.driverConfig.AccessTokenOverride, nil
	}
	return c.dotRill.GetAccessToken()
}

func (c *connection) adminURL() (string, error) {
	if c.driverConfig.AdminURLOverride != "" {
		return c.driverConfig.AdminURLOverride, nil
	}
	adminURL, err := c.dotRill.GetDefaultAdminURL()
	if err != nil {
		return "", err
	}
	if adminURL == "" {
		adminURL = defaultAdminURL
	}
	return adminURL, nil
}

const defaultAdminURL = "https://admin.rilldata.com"
