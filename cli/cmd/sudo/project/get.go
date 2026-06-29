package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Get project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}
			res, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				Org:                  args[0],
				Project:              args[1],
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			annotations := make([]string, 0, len(res.Project.Annotations))
			for k, v := range res.Project.Annotations {
				annotations = append(annotations, fmt.Sprintf("%s=%s", k, v))
			}

			project := res.Project
			fmt.Printf("Name: %s (ID: %s)\n", project.Name, project.Id)
			fmt.Printf("Org: %s (ID: %s)\n", project.OrgName, project.OrgId)
			fmt.Printf("Created on: %s\n", project.CreatedOn.AsTime().Format(time.RFC3339Nano))
			fmt.Printf("Public: %t\n", project.Public)
			fmt.Printf("Created by user ID: %s\n", project.CreatedByUserId)
			fmt.Printf("Provisioner: %s\n", project.Provisioner)
			fmt.Printf("Git remote: %s\n", project.GitRemote)
			fmt.Printf("Subpath: %s\n", project.Subpath)
			fmt.Printf("Prod version: %s\n", project.ProdVersion)
			fmt.Printf("Primary branch: %s\n", project.PrimaryBranch)
			fmt.Printf("Prod slots: %d\n", project.ProdSlots)
			fmt.Printf("Primary deployment ID: %s\n", project.PrimaryDeploymentId)
			fmt.Printf("Prod hibernation TTL: %s\n", time.Duration(project.ProdTtlSeconds)*time.Second)
			fmt.Printf("Prod deployment status: %s\n", prodDeploymentStatus(res.Deployment))
			fmt.Printf("Annotations: %s\n", strings.Join(annotations, "; "))

			return nil
		},
	}

	return getCmd
}

// prodDeploymentStatus returns a human-readable description of the prod deployment's status.
// A project with no prod deployment, or one that is stopped, is hibernated.
func prodDeploymentStatus(depl *adminv1.Deployment) string {
	if depl == nil {
		return "Hibernated (no deployment)"
	}
	switch depl.Status {
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_PENDING:
		return "Pending"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING:
		return "Running"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_ERRORED:
		return "Errored"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPED:
		return "Hibernated (stopped)"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_UPDATING:
		return "Updating"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_STOPPING:
		return "Hibernating (stopping)"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_DELETING:
		return "Deleting"
	case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_DELETED:
		return "Deleted"
	default:
		return "Unknown"
	}
}
