package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var prodSlots, infraSlots, clusterSlots, rillSlots int
	var prodVersion string

	editCmd := &cobra.Command{
		Use:   "edit <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if cmd.Flags().Changed("rill-slots") && cmd.Flags().Changed("prod-slots") {
				return fmt.Errorf("--rill-slots and --prod-slots are mutually exclusive")
			}

			req := &adminv1.UpdateProjectRequest{
				Org:                  args[0],
				Project:              args[1],
				SuperuserForceAccess: true,
			}

			isEditRequested := false

			if cmd.Flags().Changed("prod-slots") {
				if prodSlots < 0 {
					return fmt.Errorf("--prod-slots must be >= 0")
				}
				v := int64(prodSlots)
				req.ProdSlots = &v
				isEditRequested = true
			}
			if cmd.Flags().Changed("rill-slots") {
				if rillSlots < 0 {
					return fmt.Errorf("--rill-slots must be >= 0")
				}
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

			// --rill-slots: resolve prod_slots = cluster_slots + rill_slots.
			// For Rill Managed (DuckDB): prod_slots = rill_slots directly (no cluster slots).
			// For Live Connect: fetch cluster_slots from DB or use --cluster-slots flag.
			if cmd.Flags().Changed("rill-slots") {
				proj, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
					Org:                  args[0],
					Project:              args[1],
					SuperuserForceAccess: true,
				})
				if err != nil {
					return fmt.Errorf("fetching project for --rill-slots: %w", err)
				}
				isRillManaged := isRillManagedProject(proj.Project)
				if isRillManaged {
					// Managed: prod_slots = rill_slots (no cluster component)
					newProdSlots := int64(rillSlots)
					req.ProdSlots = &newProdSlots
				} else {
					// Live Connect: prod_slots = cluster_slots + rill_slots
					var clusterSlotsForCalc int64
					if cmd.Flags().Changed("cluster-slots") {
						clusterSlotsForCalc = int64(clusterSlots)
					} else if proj.Project.ClusterSlots != nil {
						clusterSlotsForCalc = *proj.Project.ClusterSlots
					} else {
						return fmt.Errorf("cluster_slots is not set for this project -- set it first with --cluster-slots")
					}
					newProdSlots := clusterSlotsForCalc + int64(rillSlots)
					req.ProdSlots = &newProdSlots
				}
			}

			updatedProj, err := client.UpdateProject(ctx, req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated project\n")
			proj := updatedProj.Project
			isManaged := isRillManagedProject(proj)
			if isManaged {
				fmt.Printf("Rill slots:    %d\n", proj.ProdSlots)
				fmt.Printf("Cluster slots: 0\n")
				fmt.Printf("Infra slots:   0\n")
				fmt.Printf("Prod slots:    %d\n", proj.ProdSlots)
			} else {
				infraLabel := "(default)"
				infraVal := int64(4)
				if proj.InfraSlots != nil {
					infraLabel = ""
					infraVal = *proj.InfraSlots
				}
				clusterSlotsVal := int64(4)
				clusterLabel := "(default)"
				if proj.ClusterSlots != nil {
					clusterSlotsVal = *proj.ClusterSlots
					clusterLabel = ""
				}
				rillSlotsVal := proj.ProdSlots - clusterSlotsVal
				if rillSlotsVal < 0 {
					rillSlotsVal = 0
				}
				fmt.Printf("Rill slots:    %d\n", rillSlotsVal)
				fmt.Printf("Cluster slots: %d %s\n", clusterSlotsVal, clusterLabel)
				fmt.Printf("Infra slots:   %d %s\n", infraVal, infraLabel)
				fmt.Printf("Prod slots:    %d (cluster + rill)\n", proj.ProdSlots)
			}

			return nil
		},
	}

	editCmd.Flags().IntVar(&prodSlots, "prod-slots", 0, "Total prod slots (cluster + rill); use --rill-slots for Live Connect projects")
	editCmd.Flags().IntVar(&rillSlots, "rill-slots", 0, "Rill (user) slots on top of cluster slots; sets prod_slots = cluster_slots + rill_slots")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version for production deployment")
	editCmd.Flags().IntVar(&infraSlots, "infra-slots", 0, "Rill infra overhead slot allocation (Live Connect only; 0 = use default of 4)")
	editCmd.Flags().IntVar(&clusterSlots, "cluster-slots", 0, "Cluster slot allocation override (stored as rill_min_slots in DB)")
	return editCmd
}
