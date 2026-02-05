package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, path string

	showCmd := &cobra.Command{
		Use:   "show [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Show project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				name, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				Org:     ch.Org,
				Project: name,
			})
			if err != nil {
				return err
			}

			ch.PrintProjects([]*adminv1.Project{proj.Project})

			return nil
		},
	}

	showCmd.Flags().SortFlags = false
	showCmd.Flags().StringVar(&name, "project", "", "Name")
	showCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return showCmd
}
