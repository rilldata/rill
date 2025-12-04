package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func DescribeCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string

	statusCmd := &cobra.Command{
		Use:   "describe [<project-name>] <type> <name>",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(2), cobra.MaximumNArgs(3)),
		Short: "Retrieve detailed state for a resource",
		Long:  "Retrieve detailed state for a specific resource (source, model, dashboard, ...)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 3 {
				project = args[0]
			}
			if !cmd.Flags().Changed("project") && len(args) == 2 && ch.Interactive {
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			kind := runtime.ResourceKindFromShorthand(args[len(args)-2])
			name := args[len(args)-1]

			proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				Org:     ch.Org,
				Project: project,
			})
			if err != nil {
				return err
			}

			depl := proj.ProdDeployment
			if depl == nil {
				return fmt.Errorf("no production deployment found for project %q", project)
			}

			rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}

			res, err := rt.GetResource(cmd.Context(), &runtimev1.GetResourceRequest{
				InstanceId: depl.RuntimeInstanceId,
				Name: &runtimev1.ResourceName{
					Kind: kind,
					Name: name,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			enc := protojson.MarshalOptions{
				Multiline:       true,
				EmitUnpopulated: true,
			}

			data, err := enc.Marshal(res.Resource)
			if err != nil {
				return fmt.Errorf("failed to marshal resource as JSON: %w", err)
			}

			fmt.Println(string(data))

			return nil
		},
	}

	statusCmd.Flags().StringVar(&project, "project", "", "Project name")
	statusCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return statusCmd
}
