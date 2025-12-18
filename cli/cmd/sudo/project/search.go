package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
)

func SearchCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var annotations map[string]string
	var statusFlag bool

	searchCmd := &cobra.Command{
		Use:   "search [<pattern>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Search projects by pattern",
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

			if !statusFlag {
				ch.PrintData(res.Names)
			} else {
				// We need to fetch the status of each project by connecting to their individual runtime instances.
				// Using an errgroup to parallelize the requests.
				table := make([]*projectStatusTableRow, len(res.Names))
				grp, ctx := errgroup.WithContext(ctx)
				for idx, name := range res.Names {
					org := strings.Split(name, "/")[0]
					project := strings.Split(name, "/")[1]

					idx := idx
					grp.Go(func() error {
						row, err := newProjectStatusTableRow(ctx, client, org, project)
						if err != nil {
							return err
						}
						row.DeploymentStatus = truncMessage(row.DeploymentStatus, 35)
						table[idx] = row
						return nil
					})
				}

				err := grp.Wait()
				if err != nil {
					return err
				}

				ch.PrintData(table)
			}

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	searchCmd.Flags().BoolVar(&statusFlag, "status", false, "Include project status")
	searchCmd.Flags().StringToStringVar(&annotations, "annotation", nil, "Annotations to filter projects by (supports wildcard values)")
	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}

type projectStatusTableRow struct {
	Org                  string `header:"org"`
	Project              string `header:"project"`
	DeploymentStatus     string `header:"deployment"`
	IdleCount            int    `header:"idle"`
	PendingCount         int    `header:"pending"`
	RunningCount         int    `header:"running"`
	ReconcileErrorsCount int    `header:"reconcile errors"`
	ParseErrorsCount     int    `header:"parse errors"`
}

func newProjectStatusTableRow(ctx context.Context, c *client.Client, org, project string) (*projectStatusTableRow, error) {
	proj, err := c.GetProject(ctx, &adminv1.GetProjectRequest{
		Org:                  org,
		Project:              project,
		SuperuserForceAccess: true,
		IssueSuperuserToken:  true,
	})
	if err != nil {
		return nil, err
	}

	depl := proj.ProdDeployment

	if depl == nil {
		return &projectStatusTableRow{
			Org:              org,
			Project:          project,
			DeploymentStatus: "Hibernated",
		}, nil
	}

	if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
		var deplStatus string
		switch depl.Status {
		case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_PENDING:
			deplStatus = "Pending"
		case adminv1.DeploymentStatus_DEPLOYMENT_STATUS_ERRORED:
			deplStatus = "Errored"
		default:
			deplStatus = depl.Status.String()
		}

		return &projectStatusTableRow{
			Org:              org,
			Project:          project,
			DeploymentStatus: deplStatus,
		}, nil
	}

	rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
	if err != nil {
		return &projectStatusTableRow{
			Org:              org,
			Project:          project,
			DeploymentStatus: fmt.Sprintf("Connection error: %v", err),
		}, nil
	}

	res, err := rt.ListResources(ctx, &runtimev1.ListResourcesRequest{InstanceId: depl.RuntimeInstanceId})
	if err != nil {
		msg := err.Error()
		if s, ok := status.FromError(err); ok {
			msg = s.Message()
		}

		return &projectStatusTableRow{
			Org:              org,
			Project:          project,
			DeploymentStatus: fmt.Sprintf("Runtime error: %v", msg),
		}, nil
	}

	var parser *runtimev1.Resource
	var parseErrorsCount int
	var idleCount int
	var reconcileErrorsCount int
	var pendingCount int
	var runningCount int

	for _, r := range res.Resources {
		if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
			parser = r
		}
		if r.Meta.Hidden {
			continue
		}

		switch r.Meta.ReconcileStatus {
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
			idleCount++
			if r.Meta.GetReconcileError() != "" {
				reconcileErrorsCount++
			}
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
			pendingCount++
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
			runningCount++
		}
	}

	// check if there are any parser errors
	if parser == nil {
		return &projectStatusTableRow{
			Org:              org,
			Project:          project,
			DeploymentStatus: "Parser not found (corrupted)",
		}, nil
	}
	if parser.GetProjectParser().State != nil {
		parseErrorsCount = len(parser.GetProjectParser().State.ParseErrors)
	}
	if parser.Meta.ReconcileError != "" {
		parseErrorsCount++
	}

	return &projectStatusTableRow{
		Org:                  org,
		Project:              project,
		DeploymentStatus:     "OK",
		IdleCount:            idleCount,
		PendingCount:         pendingCount,
		RunningCount:         runningCount,
		ReconcileErrorsCount: reconcileErrorsCount,
		ParseErrorsCount:     parseErrorsCount,
	}, nil
}

func truncMessage(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:(n-3)] + "..."
}
