package project

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/spf13/cobra"
)

const _connectURL = "%s/github-connect/organizations/%s/projects?remote=remote&project_name=project&prod_branch=branch"

func ConnectCmd(cfg *config.Config) *cobra.Command {
	var name, displayName, prodBranch, projectPath string
	var public bool

	connectCmd := &cobra.Command{
		Use:   "connect",
		Args:  cobra.ExactArgs(1),
		Short: "Connect project to rill cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				projectPath = args[0]
			}
			remote, err := gitutil.ExtractRemotes(projectPath)
			// todo :: throw cli error for other errors, and return this for no remote error only
			if err != nil && len(remote) == 0 {
				return fmt.Errorf("Please push project to github and then try connect again")
			}

			endpoint, err := transport.NewEndpoint(remote[0].URL)
			if err != nil {
				return err
			}

			if !strings.Contains(endpoint.Host, "github") {
				return fmt.Errorf("Only github hosted repos are supported at this point, please push repo to github")
			}

			org := cfg.DefaultOrg
			connectURL, err := url.Parse(fmt.Sprintf(_connectURL, cfg.AdminHTTPURL, org))
			if err != nil {
				return err
			}

			q := connectURL.Query()
			q.Set("remote", url.QueryEscape(remote[0].URL))
			q.Set("project_name", name)
			q.Set("prod_branch", prodBranch)
			connectURL.RawQuery = q.Encode()
			return browser.OpenURL(connectURL.String())
		},
	}

	connectCmd.Flags().SortFlags = false

	connectCmd.Flags().StringVar(&name, "name", "noname", "Name")
	connectCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")
	connectCmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Production branch name")
	connectCmd.Flags().BoolVar(&public, "public", false, "Public")
	connectCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")

	return connectCmd
}
