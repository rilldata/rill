package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func HibernateCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var redeploy, force bool

	hibernateCmd := &cobra.Command{
		Use:   "hibernate [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Hibernate project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project name
			if len(args) > 0 {
				project = args[0]
			}
			// Only prompt interactively if no flags or args are provided
			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Get client
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Hibernate
			if !redeploy {
				_, err = client.HibernateProject(cmd.Context(), &adminv1.HibernateProjectRequest{
					Org:     ch.Org,
					Project: project,
				})
				if err != nil {
					return err
				}
				return nil
			}

			// Redeploy
			if !force {
				res, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
					Org:     ch.Org,
					Project: project,
				})
				if err != nil {
					return err
				}
				if res.ProdDeployment != nil {
					return fmt.Errorf("the project %q in the organization %q is not hibernated", project, ch.Org)
				}
			}
			_, err = client.RedeployProject(cmd.Context(), &adminv1.RedeployProjectRequest{
				Org:     ch.Org,
				Project: project,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	hibernateCmd.Flags().SortFlags = false
	hibernateCmd.Flags().StringVar(&project, "project", "", "Name")
	hibernateCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	hibernateCmd.Flags().BoolVar(&redeploy, "redeploy", false, "Restore a previously hibernated project")
	hibernateCmd.Flags().BoolVar(&force, "force", false, "Force a redeploy for a non-hibernated project")

	return hibernateCmd
}
