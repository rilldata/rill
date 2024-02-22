package project

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ProjectCmd(ch *cmdutil.Helper) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Short:             "Manage projects",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	projectCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	projectCmd.AddCommand(ShowCmd(ch))
	projectCmd.AddCommand(StatusCmd(ch))
	projectCmd.AddCommand(DescribeCmd(ch))
	projectCmd.AddCommand(EditCmd(ch))
	projectCmd.AddCommand(DeleteCmd(ch))
	projectCmd.AddCommand(ListCmd(ch))
	projectCmd.AddCommand(ReconcileCmd(ch))
	projectCmd.AddCommand(RefreshCmd(ch))
	projectCmd.AddCommand(ResetCmd(ch))
	projectCmd.AddCommand(JwtCmd(ch))
	projectCmd.AddCommand(RenameCmd(ch))
	projectCmd.AddCommand(LogsCmd(ch))

	return projectCmd
}

func projectNames(ctx context.Context, ch *cmdutil.Helper) ([]string, error) {
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	org := ch.Org

	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: org})
	if err != nil {
		return nil, err
	}

	if len(resp.Projects) == 0 {
		return nil, fmt.Errorf("no projects found for org %q", org)
	}

	var projNames []string
	for _, proj := range resp.Projects {
		projNames = append(projNames, proj.Name)
	}

	return projNames, nil
}

func toTable(projects []*adminv1.Project) []*project {
	projs := make([]*project, 0, len(projects))

	for _, proj := range projects {
		projs = append(projs, toRow(proj))
	}

	return projs
}

func toRow(o *adminv1.Project) *project {
	githubURL := o.GithubUrl
	if o.Subpath != "" {
		githubURL = filepath.Join(o.GithubUrl, "tree", o.ProdBranch, o.Subpath)
	}

	return &project{
		Name:         o.Name,
		Public:       o.Public,
		GithubURL:    githubURL,
		Organization: o.OrgName,
	}
}

type project struct {
	Name         string `header:"name" json:"name"`
	Public       bool   `header:"public" json:"public"`
	GithubURL    string `header:"github" json:"github"`
	Organization string `header:"organization" json:"organization"`
}
