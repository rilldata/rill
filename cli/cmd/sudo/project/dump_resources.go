package project

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
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

			pattern := "%"
			// If args is not empty, use the first element as the pattern
			if len(args) > 0 {
				pattern = args[0]
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

			failedProjects := make([]string, len(res.Names))
			errors := make([]error, len(res.Names))

			var m sync.Mutex
			resources := map[string]map[string]json.RawMessage{}
			grp, ctx := errgroup.WithContext(ctx)
			for idx, name := range res.Names {
				org := strings.Split(name, "/")[0]
				project := strings.Split(name, "/")[1]

				grp.Go(func() error {
					row, err := dumpResourcesForProject(ctx, client, org, project, typ)
					if err != nil {
						failedProjects[idx] = name
						errors[idx] = err
						return nil
					}
					m.Lock()
					resources[name] = row
					m.Unlock()
					return nil
				})
			}

			err = grp.Wait()
			if err != nil {
				return err
			}

			// marshal as json with indentation
			jsonData, err := json.MarshalIndent(resources, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal resources: %w", err)
			}
			fmt.Println(string(jsonData))

			for idx, failed := range failedProjects {
				if failed != "" {
					ch.PrintfWarn("Failed to dump resources for project %v: %s\n", failed, errors[idx])
				}
			}
			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
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

func dumpResourcesForProject(ctx context.Context, c *client.Client, org, project, filter string) (map[string]json.RawMessage, error) {
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

	res, err := rt.ListResources(ctx, &runtimev1.ListResourcesRequest{InstanceId: depl.RuntimeInstanceId})
	if err != nil {
		msg := err.Error()
		if s, ok := status.FromError(err); ok {
			msg = s.Message()
		}
		return nil, fmt.Errorf("runtime error, failed to list resources: %v", msg)
	}

	result := make(map[string]json.RawMessage)
	for _, r := range res.Resources {
		jsonData, err := protojson.MarshalOptions{Indent: " "}.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal resource: %w", err)
		}

		kind := runtime.PrettifyResourceKind(r.Meta.Name.Kind)
		if filter != "" && !strings.EqualFold(kind, filter) {
			continue
		}
		result[kind+"/"+r.Meta.Name.Name] = jsonData
	}
	return result, nil
}
