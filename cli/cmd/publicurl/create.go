package publicurl

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var ttlMinutes int
	var filter string
	var fields []string

	createCmd := &cobra.Command{
		Use:   "create [<project-name>] <explore>",
		Short: "Create a public URL",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 2 {
				project = args[0]
			}
			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			explore := args[len(args)-1]

			var filterExpr *runtimev1.Expression
			if filter != "" {
				filterExpr = &runtimev1.Expression{}
				err := protojson.Unmarshal([]byte(filter), filterExpr)
				if err != nil {
					return fmt.Errorf("failed to parse filter expression: %w", err)
				}
			}

			err = validateExplore(cmd.Context(), ch, project, explore, fields)
			if err != nil {
				return err
			}

			res, err := client.IssueMagicAuthToken(cmd.Context(), &adminv1.IssueMagicAuthTokenRequest{
				Org:          ch.Org,
				Project:      project,
				TtlMinutes:   int64(ttlMinutes),
				ResourceType: runtime.ResourceKindExplore,
				ResourceName: explore,
				Filter:       filterExpr,
				Fields:       fields,
			})
			if err != nil {
				return err
			}

			ch.Printf("%s\n", res.Url)
			return nil
		},
	}

	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&project, "project", "", "Project name")
	createCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	createCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 0, "Duration until the token expires (use 0 for no expiry)")
	createCmd.Flags().StringVar(&filter, "filter", "", "Limit access to the provided filter (json)")
	createCmd.Flags().StringSliceVar(&fields, "fields", nil, "Limit access to the provided fields")

	return createCmd
}

func validateExplore(ctx context.Context, ch *cmdutil.Helper, project, explore string, fields []string) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	proj, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
		Org:     ch.Org,
		Project: project,
	})
	if err != nil {
		return err
	}

	if proj.ProdDeployment == nil {
		ch.PrintfWarn("Could not validate metrics view: project has no production deployment")
		return nil
	}
	depl := proj.ProdDeployment

	rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
	if err != nil {
		return fmt.Errorf("failed to connect to runtime: %w", err)
	}

	expl, err := rt.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceId,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindExplore,
			Name: explore,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get explore %q: %w", explore, err)
	}

	spec := expl.Resource.GetExplore().State.ValidSpec
	if spec == nil {
		return fmt.Errorf("explore %q is invalid", explore)
	}

	mv, err := rt.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceId,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindMetricsView,
			Name: spec.MetricsView,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get metrics view %q: %w", spec.MetricsView, err)
	}

	for _, f := range fields {
		if strings.EqualFold(f, mv.Resource.GetMetricsView().Spec.TimeDimension) {
			continue
		}

		found := false
		for _, dim := range spec.Dimensions {
			if strings.EqualFold(f, dim) {
				found = true
				break
			}
		}
		if found {
			continue
		}

		for _, m := range spec.Measures {
			if strings.EqualFold(f, m) {
				found = true
				break
			}
		}
		if found {
			continue
		}

		return fmt.Errorf("field %q not found in explore %q", f, explore)
	}

	return nil
}
