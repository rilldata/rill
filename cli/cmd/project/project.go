package project

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ProjectCmd(cfg *config.Config) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Hidden:            !cfg.IsDev(),
		Short:             "Manage projects",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	projectCmd.PersistentFlags().StringVar(&cfg.Org, "org", cfg.Org, "Organization Name")
	projectCmd.AddCommand(ShowCmd(cfg))
	projectCmd.AddCommand(StatusCmd(cfg))
	projectCmd.AddCommand(EditCmd(cfg))
	projectCmd.AddCommand(DeleteCmd(cfg))
	projectCmd.AddCommand(ListCmd(cfg))
	projectCmd.AddCommand(EnvCmd(cfg))
	projectCmd.AddCommand(MembersCmd(cfg))
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
	return &project{
		Name:         o.Name,
		Public:       o.Public,
		GithubURL:    o.GithubUrl,
		Organization: o.OrgName,
	}
}

type project struct {
	Name         string `header:"name" json:"name"`
	Public       bool   `header:"public" json:"public"`
	GithubURL    string `header:"github" json:"github"`
	Organization string `header:"organization" json:"organization"`
}
