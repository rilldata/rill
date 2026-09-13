package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var prodSlots, devSlots int
	var prodVersion string
	var overrideDiskGB int64
	var cloudEditingDisabled bool

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

			isProjectEditRequested := false
			if cmd.Flags().Changed("prod-slots") {
				if prodSlots <= 0 {
					return fmt.Errorf("--prod-slots must be greater than zero")
				}
				prodSlotsInt64 := int64(prodSlots)
				req.ProdSlots = &prodSlotsInt64
				isProjectEditRequested = true
			}
			if cmd.Flags().Changed("prod-version") {
				req.ProdVersion = &prodVersion
				isProjectEditRequested = true
			}
			if cmd.Flags().Changed("dev-slots") {
				if devSlots <= 0 {
					return fmt.Errorf("--dev-slots must be greater than zero")
				}
				devSlotsInt64 := int64(devSlots)
				req.DevSlots = &devSlotsInt64
				isProjectEditRequested = true
			}
			if cmd.Flags().Changed("override-disk-gb") {
				if overrideDiskGB < 0 {
					return fmt.Errorf("--override-disk-gb must be >= 0 (use 0 to clear the override)")
				}
				v := overrideDiskGB
				req.OverrideDiskGb = &v
				isProjectEditRequested = true
			}

			isCloudEditingEditRequested := cmd.Flags().Changed("cloud-editing-disabled")
			if !isProjectEditRequested && !isCloudEditingEditRequested {
				ch.Printf("No edit requested\n")
				return nil
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			var updatedProject *adminv1.Project
			if isProjectEditRequested {
				res, err := client.UpdateProject(ctx, req)
				if err != nil {
					return err
				}
				updatedProject = res.Project
			}

			if isCloudEditingEditRequested {
				res, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
					Org:                  args[0],
					Project:              args[1],
					SuperuserForceAccess: true,
				})
				if err != nil {
					return err
				}

				annotations, changed := setCloudEditingDisabledAnnotation(res.Project.Annotations, cloudEditingDisabled)
				if changed {
					updatedAnnotations, err := client.SudoUpdateAnnotations(ctx, &adminv1.SudoUpdateAnnotationsRequest{
						Org:         args[0],
						Project:     args[1],
						Annotations: annotations,
					})
					if err != nil {
						return err
					}
					updatedProject = updatedAnnotations.Project
				} else {
					updatedProject = res.Project
				}
			}

			ch.PrintfSuccess("Updated project\n")
			ch.PrintProjects([]*adminv1.Project{updatedProject})

			return nil
		},
	}

	editCmd.Flags().IntVar(&prodSlots, "prod-slots", 0, "Slots to allocate for production deployments")
	editCmd.Flags().IntVar(&devSlots, "dev-slots", 0, "Slots to allocate for dev deployments")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version for production deployment")
	editCmd.Flags().Int64Var(&overrideDiskGB, "override-disk-gb", 0, "Override disk size in GB for prod and dev deployments (0 clears the override)")
	editCmd.Flags().BoolVar(&cloudEditingDisabled, "cloud-editing-disabled", false, "Hide cloud editing in the UI even when enabled in rill.yaml")
	return editCmd
}

func setCloudEditingDisabledAnnotation(annotations map[string]string, disabled bool) (map[string]string, bool) {
	res := make(map[string]string, len(annotations)+1)
	for k, v := range annotations {
		res[k] = v
	}

	if disabled {
		if res[runtime.CloudEditingDisabledAnnotation] == "true" {
			return res, false
		}
		res[runtime.CloudEditingDisabledAnnotation] = "true"
		return res, true
	}

	if _, ok := res[runtime.CloudEditingDisabledAnnotation]; !ok {
		return res, false
	}
	delete(res, runtime.CloudEditingDisabledAnnotation)
	return res, true
}
