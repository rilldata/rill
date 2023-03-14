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
		PersistentPreRunE: cmdutil.CheckAuth(cfg),
	}
	projectCmd.AddCommand(ShowCmd(cfg))
	projectCmd.AddCommand(StatusCmd(cfg))
	projectCmd.AddCommand(ConnectCmd(cfg))
	projectCmd.AddCommand(EditCmd(cfg))
	projectCmd.AddCommand(DeleteCmd(cfg))
	projectCmd.AddCommand(ListCmd(cfg))
	projectCmd.AddCommand(EnvCmd(cfg))
	return projectCmd
}

func toTable(projects []*adminv1.Project) []*project {
	orgs := make([]*project, 0, len(projects))

	for _, org := range projects {
		orgs = append(orgs, toRow(org))
	}

	return orgs
}

func toRow(o *adminv1.Project) *project {
	return &project{
		Name:      o.Name,
		Public:    o.Public,
		GithubURL: o.GithubUrl,
		CreatedAt: o.CreatedOn.AsTime().String(),
	}
}

type project struct {
	Name      string `header:"name" json:"name"`
	Public    bool   `header:"public" json:"public"`
	GithubURL string `header:"github" json:"github"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}
