package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var prodSlots, devSlots int
	var prodVersion string

	editCmd := &cobra.Command{
		Use:   "edit <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			req := &adminv1.UpdateProjectRequest{
				Org:                  args[0],
				Project:              args[1],
				SuperuserForceAccess: true,
			}

			isEditRequested := false
			if cmd.Flags().Changed("prod-slots") {
				if prodSlots <= 0 {
					return fmt.Errorf("--prod-slots must be greater than zero")
				}
				prodSlotsInt64 := int64(prodSlots)
				req.ProdSlots = &prodSlotsInt64
				isEditRequested = true
			}
			if cmd.Flags().Changed("prod-version") {
				req.ProdVersion = &prodVersion
				isEditRequested = true
			}
			if cmd.Flags().Changed("dev-slots") {
				if devSlots <= 0 {
					return fmt.Errorf("--dev-slots must be greater than zero")
				}
				devSlotsInt64 := int64(devSlots)
				req.DevSlots = &devSlotsInt64
				isEditRequested = true
			}

			if !isEditRequested {
				ch.Printf("No edit requested\n")
				return nil
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			updatedProj, err := client.UpdateProject(ctx, req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project\n")
			ch.PrintProjects([]*adminv1.Project{updatedProj.Project})

			return nil
		},
	}

	editCmd.Flags().IntVar(&prodSlots, "prod-slots", 0, "Slots to allocate for production deployments")
	editCmd.Flags().IntVar(&devSlots, "dev-slots", 0, "Slots to allocate for dev deployments")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version for production deployment")
	return editCmd
}
