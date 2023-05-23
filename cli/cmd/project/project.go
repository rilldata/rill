package project

import (
	"context"
	"path/filepath"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ProjectCmd(cfg *config.Config) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Short:             "Manage projects",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	projectCmd.PersistentFlags().StringVar(&cfg.Org, "org", cfg.Org, "Organization Name")
	projectCmd.AddCommand(ShowCmd(cfg))
	projectCmd.AddCommand(StatusCmd(cfg))
	projectCmd.AddCommand(EditCmd(cfg))
	projectCmd.AddCommand(DeleteCmd(cfg))
	projectCmd.AddCommand(ListCmd(cfg))
	projectCmd.AddCommand(ReconcileCmd(cfg))
	projectCmd.AddCommand(JwtCmd(cfg))
	projectCmd.AddCommand(RenameCmd(cfg))
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
		githubURL = filepath.Join(o.GithubUrl, "tree/main", o.Subpath)
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
