package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
)

func SplitsCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path, model string
	var local bool
	var pageSize uint32
	var pageToken string

	splitsCmd := &cobra.Command{
		Use:   "splits [<project>] <model>",
		Args:  cobra.MaximumNArgs(2),
		Short: "List splits for a model",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				model = args[0]
			} else if len(args) == 2 {
				project = args[0]
				model = args[1]
			}

			if !local && !cmd.Flags().Changed("project") && len(args) <= 1 && ch.Interactive {
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			var host, instanceID, jwt string
			if local {
				// This is the default port that Rill localhost uses for gRPC.
				// TODO: In the future, we should capture the gRPC port in ~/.rill and use it here.
				host = "http://localhost:49009"
				instanceID = "default"
			} else {
				proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
					OrganizationName: ch.Org,
					Name:             project,
				})
				if err != nil {
					return err
				}

				depl := proj.ProdDeployment
				if depl == nil {
					return fmt.Errorf("project %q is not currently deployed", project)
				}
				if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_OK {
					ch.PrintfWarn("Deployment status not OK: %s\n", depl.Status.String())
					return nil
				}

				host = depl.RuntimeHost
				instanceID = depl.RuntimeInstanceId
				jwt = proj.Jwt
			}

			rt, err := runtimeclient.New(host, jwt)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}

			res, err := rt.GetModelSplits(cmd.Context(), &runtimev1.GetModelSplitsRequest{
				InstanceId: instanceID,
				Model:      model,
				PageSize:   pageSize,
				PageToken:  pageToken,
			})
			if err != nil {
				return fmt.Errorf("failed to get model splits: %w", err)
			}

			ch.PrintModelSplits(res.Splits)

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}

	splitsCmd.Flags().SortFlags = false
	splitsCmd.Flags().StringVar(&project, "project", "", "Project Name")
	splitsCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	splitsCmd.Flags().StringVar(&model, "model", "", "Model Name")
	splitsCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")
	splitsCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of splits to return per page")
	splitsCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return splitsCmd
}
