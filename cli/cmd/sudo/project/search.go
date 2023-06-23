package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SearchCmd(cfg *config.Config) *cobra.Command {
	searchCmd := &cobra.Command{
		Use:   "search",
		Args:  cobra.ExactArgs(1),
		Short: "Search projects by pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
				NamePattern: args[0],
			})
			if err != nil {
				return err
			}

			cmdutil.TablePrinter(toTable(res.Projects))
			return nil
		},
	}

	return searchCmd
}

type projectName struct {
	Organization string `header:"organization" json:"organization"`
	ProjectName  string `header:"project name" json:"project_name"`
}

func toTable(projects []*adminv1.ProjectName) []*projectName {
	projs := make([]*projectName, 0, len(projects))

	for _, proj := range projects {
		projs = append(projs, toRow(proj))
	}

	return projs
}

func toRow(o *adminv1.ProjectName) *projectName {
	return &projectName{
		Organization: o.OrgName,
		ProjectName:  o.ProjectName,
	}
}
