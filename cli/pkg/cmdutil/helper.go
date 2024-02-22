package cmdutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
)

type Helper struct {
	Version            Version
	AdminURL           string
	AdminTokenOverride string
	AdminTokenDefault  string
	Org                string
	Interactive        bool
	Printer            *printer.Printer
	admin              *client.Client
}

func (h *Helper) Close() error {
	if h.admin != nil {
		return h.admin.Close()
	}
	return nil
}

func (h *Helper) IsDev() bool {
	return h.Version.IsDev()
}

func (h *Helper) IsAuthenticated() bool {
	return h.AdminToken() != ""
}

func (h *Helper) AdminToken() string {
	if h.AdminTokenOverride != "" {
		return h.AdminTokenOverride
	}
	return h.AdminTokenDefault
}

func (h *Helper) Client() (*client.Client, error) {
	if h.admin == nil {
		cliVersion := h.Version.Number
		if cliVersion == "" {
			cliVersion = "unknown"
		}

		userAgent := fmt.Sprintf("rill-cli/%v", cliVersion)
		c, err := client.New(h.AdminURL, h.AdminToken(), userAgent)
		if err != nil {
			return nil, err
		}

		h.admin = c
	}

	return h.admin, nil
}

func (h *Helper) CurrentUser(ctx context.Context) (*adminv1.User, error) {
	c, err := h.Client()
	if err != nil {
		return nil, err
	}

	res, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return nil, err
	}

	return res.User, nil
}

func (h *Helper) ProjectNamesByGithubURL(ctx context.Context, org, githubURL string) ([]string, error) {
	c, err := h.Client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{
		OrganizationName: org,
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, p := range resp.Projects {
		if strings.EqualFold(p.GithubUrl, githubURL) {
			names = append(names, p.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no project with githubURL %q exist in org %q", githubURL, org)
	}

	return names, nil
}

func (h *Helper) InferProjectName(ctx context.Context, org, path string) (string, error) {
	// Verify projectPath is a Git repo with remote on Github
	_, githubURL, err := gitutil.ExtractGitRemote(path, "")
	if err != nil {
		return "", err
	}

	// fetch project names for github url
	names, err := h.ProjectNamesByGithubURL(ctx, org, githubURL)
	if err != nil {
		return "", err
	}

	if len(names) == 1 {
		return names[0], nil
	}
	// prompt for name from user
	return SelectPrompt("Select project", names, ""), nil
}
