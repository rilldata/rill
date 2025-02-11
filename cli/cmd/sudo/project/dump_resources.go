package project

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
)

func DumpResources(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var annotations map[string]string
	var typ string

	searchCmd := &cobra.Command{
		Use:   "dump-resources [<pattern>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Dump resources for projects by pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			var pattern string
			// If args is not empty, use the first element as the pattern
			if len(args) > 0 {
				pattern = args[0]
			} else {
				pattern = "%"
			}

			res, err := client.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
				NamePattern: pattern,
				Annotations: annotations,
				PageSize:    pageSize,
				PageToken:   pageToken,
			})
			if err != nil {
				return err
			}

			if len(res.Names) == 0 {
				ch.PrintfWarn("No projects found\n")
				return nil
			}

			var m sync.Mutex
			failedProjects := map[string]error{}
			resources := map[string]map[string][]*runtimev1.Resource{}
			grp, ctx := errgroup.WithContext(ctx)
			for _, name := range res.Names {
				org := strings.Split(name, "/")[0]
				project := strings.Split(name, "/")[1]

				grp.Go(func() error {
					row, err := resourcesForProject(ctx, client, org, project, typ)
					if err != nil {
						m.Lock()
						failedProjects[name] = err
						m.Unlock()
						return nil
					}
					m.Lock()
					projects, ok := resources[org]
					if !ok {
						projects = map[string][]*runtimev1.Resource{}
						resources[org] = projects
					}
					projects[project] = row
					m.Unlock()
					return nil
				})
			}

			err = grp.Wait()
			if err != nil {
				return err
			}

			printer.NewPrinter(printer.FormatJSON).PrintResource(resources)

			for name, err := range failedProjects {
				ch.Println()
				ch.PrintfWarn("Failed to dump resources for project %v: %s\n", name, err)
			}
			if res.NextPageToken != "" {
				ch.Println()
				ch.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	searchCmd.Flags().StringVar(&typ, "type", "", "Filter for resources of a specific type")
	searchCmd.Flags().StringToStringVar(&annotations, "annotation", nil, "Annotations to filter projects by (supports wildcard values)")
	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 1000, "Number of projects to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}

func resourcesForProject(ctx context.Context, c *client.Client, org, project, filter string) ([]*runtimev1.Resource, error) {
	proj, err := c.GetProject(ctx, &adminv1.GetProjectRequest{
		OrganizationName:    org,
		Name:                project,
		IssueSuperuserToken: true,
	})
	if err != nil {
		return nil, err
	}

	depl := proj.ProdDeployment
	if depl == nil {
		return nil, nil
	}

	rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to runtime: %w", err)
	}

	req := &runtimev1.ListResourcesRequest{
		InstanceId: depl.RuntimeInstanceId,
	}
	if filter != "" {
		req.Kind = parseResourceKind(filter)
	}
	res, err := rt.ListResources(ctx, req)
	if err != nil {
		msg := err.Error()
		if s, ok := status.FromError(err); ok {
			msg = s.Message()
		}
		return nil, fmt.Errorf("runtime error, failed to list resources: %v", msg)
	}

	return res.Resources, nil
}

func parseResourceKind(k string) string {
	switch strings.ToLower(strings.TrimSpace(k)) {
	case "source":
		return runtime.ResourceKindSource
	case "model":
		return runtime.ResourceKindModel
	case "metricsview", "metrics_view":
		return runtime.ResourceKindMetricsView
	case "explore":
		return runtime.ResourceKindExplore
	case "migration":
		return runtime.ResourceKindMigration
	case "report":
		return runtime.ResourceKindReport
	case "alert":
		return runtime.ResourceKindAlert
	case "theme":
		return runtime.ResourceKindTheme
	case "component":
		return runtime.ResourceKindComponent
	case "canvas":
		return runtime.ResourceKindCanvas
	case "api":
		return runtime.ResourceKindAPI
	case "connector":
		return runtime.ResourceKindConnector
	default:
		return k
	}
}
