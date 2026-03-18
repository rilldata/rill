package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var prodSlots, infraSlots, clusterSlots int
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
			if cmd.Flags().Changed("infra-slots") {
				v := int64(infraSlots)
				req.InfraSlots = &v
				isEditRequested = true
			}
			if cmd.Flags().Changed("cluster-slots") {
				v := int64(clusterSlots)
				req.ClusterSlots = &v
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
			fmt.Printf("Prod slots: %d\n", proj.ProdSlots)
			if proj.InfraSlots != nil {
				fmt.Printf("Infra slots: %d\n", *proj.InfraSlots)
			}
			if proj.RillMinSlots != nil {
				fmt.Printf("Cluster slots: %d\n", *proj.RillMinSlots)
			}

			return nil
		},
	}

	editCmd.Flags().IntVar(&prodSlots, "prod-slots", 0, "Total slots for production deployments")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version for production deployment")
	editCmd.Flags().IntVar(&infraSlots, "infra-slots", 0, "Rill infra overhead slot allocation (Live Connect only; 0 = use default of 4)")
	editCmd.Flags().IntVar(&clusterSlots, "cluster-slots", 0, "Cluster slot allocation override (maps to rill_min_slots)")
	return editCmd
}
