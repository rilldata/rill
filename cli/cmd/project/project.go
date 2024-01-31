package project

import (
	"context"
	"path/filepath"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ProjectCmd(ch *cmdutil.Helper) *cobra.Command {
	cfg := ch.Config
	projectCmd := &cobra.Command{
		Use:               "project",
		Short:             "Manage projects",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	projectCmd.PersistentFlags().StringVar(&cfg.Org, "org", cfg.Org, "Organization Name")
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

func inferProjectName(ctx context.Context, adminClient *client.Client, org, path string) (string, error) {
	// Verify projectPath is a Git repo with remote on Github
	_, githubURL, err := gitutil.ExtractGitRemote(path, "")
	if err != nil {
		return "", err
	}

	// fetch project names for github url
	names, err := cmdutil.ProjectNamesByGithubURL(ctx, adminClient, org, githubURL)
	if err != nil {
		return "", err
	}

	if len(names) == 1 {
		return names[0], nil
	}
	// prompt for name from user
	return cmdutil.SelectPrompt("Select project", names, ""), nil
}

type project struct {
	Name         string `header:"name" json:"name"`
	Public       bool   `header:"public" json:"public"`
	GithubURL    string `header:"github" json:"github"`
	Organization string `header:"organization" json:"organization"`
}
