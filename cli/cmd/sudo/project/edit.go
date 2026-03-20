package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var rillSlots int
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

			if cmd.Flags().Changed("rill-slots") {
				if rillSlots < 2 {
					return fmt.Errorf("--rill-slots must be >= 2")
				}
				v := int64(rillSlots)
				req.ProdSlots = &v
				isEditRequested = true
			}
			if cmd.Flags().Changed("prod-version") {
				req.ProdVersion = &prodVersion
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
			proj := updatedProj.Project
			fmt.Printf("Rill Slots: %d\n", proj.ProdSlots)

			return nil
		},
	}

	editCmd.Flags().IntVar(&rillSlots, "rill-slots", 0, "Rill slots (minimum 2); sets prod_slots directly")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version for production deployment")
	return editCmd
}
