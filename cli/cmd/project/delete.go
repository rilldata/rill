package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, path string
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Delete the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				name, err = ch.InferProjectName(cmd.Context(), path, "use --project to specify the name")
				if err != nil {
					return err
				}
			}

			if !force && ch.Interactive {
				ch.PrintfWarn("Warn: Deleting the project %q will remove all metadata associated with the project\n", name)

				msg := fmt.Sprintf("Type %q to confirm deletion", name)
				project, err := cmdutil.InputPrompt(msg, "")
				if err != nil {
					return err
				}

				if project != name {
					return fmt.Errorf("entered incorrect name : %q, expected value is %q", project, name)
				}
			}

			_, err = client.DeleteProject(cmd.Context(), &adminv1.DeleteProjectRequest{
				Org:     ch.Org,
				Project: name,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Deleted project: %v\n", name)
			return nil
		},
	}

	deleteCmd.Flags().SortFlags = false
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete forcefully, skips the confirmation")
	deleteCmd.Flags().StringVar(&name, "project", "", "Project Name")
	deleteCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return deleteCmd
}
