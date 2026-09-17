package project

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// dumpInstancesConcurrency bounds how many runtimes we query in parallel.
	dumpInstancesConcurrency = 16

	// dumpInstancesMaxErrLen truncates error messages from the runtime.
	// A deployment pointing at a decommissioned runtime host returns an HTML error page,
	// which gRPC embeds in the status message and would otherwise flood the output.
	dumpInstancesMaxErrLen = 200
)

func DumpInstances(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var annotations map[string]string
	var sensitive bool

	dumpCmd := &cobra.Command{
		Use:   "dump-instances [<project-pattern>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Dump runtime instance info for projects by pattern",
		Long:  "Dump the runtime instance of each matching project, including its feature flags, connectors, variables and annotations.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := ch.Client()
			if err != nil {
				return err
			}

			pattern := "%"
			if len(args) > 0 {
				pattern = args[0]
			}

			res, err := c.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
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
			var skipped int
			failedProjects := map[string]string{}
			var instances []*projectInstance
			grp, ctx := errgroup.WithContext(ctx)
			grp.SetLimit(dumpInstancesConcurrency)
			for _, name := range res.Names {
				org, project, ok := strings.Cut(name, "/")
				if !ok {
					continue
				}

				grp.Go(func() error {
					inst, err := instanceForProject(ctx, c, org, project, sensitive)
					if err != nil {
						m.Lock()
						failedProjects[name] = shortErrMsg(err)
						m.Unlock()
						return nil
					}
					if inst == nil {
						m.Lock()
						skipped++
						m.Unlock()
						return nil
					}
					m.Lock()
					instances = append(instances, inst)
					m.Unlock()
					return nil
				})
			}

			err = grp.Wait()
			if err != nil {
				return err
			}

			printInstances(ch, instances)

			if skipped > 0 {
				ch.Println()
				ch.Printf("Skipped %d project(s) without a running deployment\n", skipped)
			}
			if len(failedProjects) > 0 {
				ch.Println()
			}
			for name, msg := range failedProjects {
				ch.PrintfWarn("Failed to dump instance for project %v: %s\n", name, msg)
			}
			if res.NextPageToken != "" {
				ch.Println()
				ch.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	dumpCmd.Flags().StringToStringVar(&annotations, "annotation", nil, "Annotations to filter projects by (supports wildcard values)")
	dumpCmd.Flags().Uint32Var(&pageSize, "page-size", 1000, "Number of projects to return per page")
	dumpCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	dumpCmd.Flags().BoolVar(&sensitive, "sensitive", false, "Include sensitive values, such as connector configs and variables")

	return dumpCmd
}

// projectInstance pairs a runtime instance with the project it belongs to.
type projectInstance struct {
	Org         string
	Project     string
	RuntimeHost string
	Instance    *runtimev1.Instance
}

func instanceForProject(ctx context.Context, c *client.Client, org, project string, sensitive bool) (*projectInstance, error) {
	proj, err := c.GetProject(ctx, &adminv1.GetProjectRequest{
		Org:                  org,
		Project:              project,
		SuperuserForceAccess: true,
		IssueSuperuserToken:  true,
	})
	if err != nil {
		return nil, err
	}

	// Skip projects without a running deployment.
	// A stopped deployment still carries a runtime host and instance ID, but its runtime has been hibernated or deprovisioned,
	// so dialing it fails. Note that hibernating a project clears its primary deployment, but stopping a deployment does not.
	depl := proj.Deployment
	if depl == nil || depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
		return nil, nil
	}

	rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to runtime: %w", err)
	}
	defer rt.Close()

	res, err := rt.GetInstance(ctx, &runtimev1.GetInstanceRequest{
		InstanceId: depl.RuntimeInstanceId,
		Sensitive:  sensitive,
	})
	if err != nil {
		msg := err.Error()
		if s, ok := status.FromError(err); ok {
			msg = s.Message()
		}
		return nil, fmt.Errorf("runtime error, failed to get instance: %v", msg)
	}

	return &projectInstance{
		Org:         org,
		Project:     project,
		RuntimeHost: depl.RuntimeHost,
		Instance:    res.Instance,
	}, nil
}

func printInstances(ch *cmdutil.Helper, instances []*projectInstance) {
	if len(instances) == 0 {
		ch.PrintfWarn("No instances found\n")
		return
	}

	// The instances are collected concurrently, so sort them for a stable output.
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].Org != instances[j].Org {
			return instances[i].Org < instances[j].Org
		}
		return instances[i].Project < instances[j].Project
	})

	// Use the proto field names so keys match the API, e.g. "feature_flags" instead of "featureFlags".
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	rows := make([]map[string]any, 0, len(instances))
	for _, pi := range instances {
		rowJSON, err := marshaler.Marshal(pi.Instance)
		if err != nil {
			ch.PrintfWarn("Failed to marshal instance for org %v, project %v: %v\n", pi.Org, pi.Project, err)
			continue
		}
		row := make(map[string]any)
		err = json.Unmarshal(rowJSON, &row)
		if err != nil {
			ch.PrintfWarn("Failed to unmarshal instance for org %v, project %v: %v\n", pi.Org, pi.Project, err)
			continue
		}

		row["org"] = pi.Org
		row["project"] = pi.Project
		row["runtime_host"] = pi.RuntimeHost
		rows = append(rows, row)
	}

	jsonData, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		ch.PrintfWarn("Failed to marshal instances: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// shortErrMsg formats an error as a single truncated line.
func shortErrMsg(err error) string {
	msg := strings.Join(strings.Fields(err.Error()), " ")
	if len(msg) > dumpInstancesMaxErrLen {
		// ToValidUTF8 drops a rune that the cut may have split in half.
		msg = strings.ToValidUTF8(msg[:dumpInstancesMaxErrLen], "") + "..."
	}
	return msg
}
