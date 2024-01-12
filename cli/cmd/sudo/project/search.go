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
)

func SearchCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize uint32
	var pageToken string
	var tags []string
	var status bool

	searchCmd := &cobra.Command{
		Use:   "search [<pattern>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Search projects by pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			pattern := "%"
			// If args is not empty, use the first element as the pattern
			if len(args) > 0 {
				pattern = args[0]
			}

			res, err := client.SearchProjectNames(ctx, &adminv1.SearchProjectNamesRequest{
				NamePattern: pattern,
				Tags:        tags,
				PageSize:    pageSize,
				PageToken:   pageToken,
			})
			if err != nil {
				return err
			}
			if len(res.Names) == 0 {
				ch.Printer.PrintlnWarn("No projects found")
				return nil
			}

			if status {
				var table []*projectStatusTableRow
				ch.Printer.Println()
				for _, name := range res.Names {
					org := strings.Split(name, "/")[0]
					project := strings.Split(name, "/")[1]

					row, err := newProjectStatusTableRow(ctx, client, org, project)
					if err != nil {
						if strings.Contains(err.Error(), "project has no prod deployment") {
							continue
						}
						return err
					}
					table = append(table, row)
				}

				ch.Printer.PrintlnSuccess("\nProject status\n")
				err = ch.Printer.PrintResource(table)
				if err != nil {
					return err
				}
			} else {
				err = ch.Printer.PrintResource(res.Names)
				if err != nil {
					return err
				}
			}

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}
	searchCmd.Flags().BoolVar(&status, "status", false, "Include project status")
	searchCmd.Flags().StringSliceVar(&tags, "tag", []string{}, "Tags to filter projects by")
	searchCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	searchCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return searchCmd
}

type projectStatusTableRow struct {
	Name                string `header:"name"`
	Org                 string `header:"org"`
	IdleCount           int    `header:"idle"`
	IdleWithErrorsCount int    `header:"idle with errors"`
	PendingCount        int    `header:"pending"`
	RunningCount        int    `header:"running"`
	ParserErrorsCount   int    `header:"parser errors"`
}

func newProjectStatusTableRow(ctx context.Context, c *client.Client, org, project string) (*projectStatusTableRow, error) {
	proj, err := c.GetProject(ctx, &adminv1.GetProjectRequest{
		OrganizationName: org,
		Name:             project,
	})
	if err != nil {
		return nil, err
	}

	depl := proj.ProdDeployment
	if depl == nil {
		return nil, fmt.Errorf("project has no prod deployment")
	}

	rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to runtime: %w", err)
	}

	res, err := rt.ListResources(ctx, &runtimev1.ListResourcesRequest{InstanceId: depl.RuntimeInstanceId})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var parser *runtimev1.ProjectParser
	var parserErrorsCount int
	var idleCount int
	var idleWithErrorsCount int
	var pendingCount int
	var runningCount int

	for _, r := range res.Resources {
		if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
			parser = r.GetProjectParser()
		}
		if r.Meta.Hidden {
			continue
		}

		switch r.Meta.ReconcileStatus {
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
			// if it is idle, check if there are any errors
			if r.Meta.GetReconcileError() != "" {
				idleWithErrorsCount++
			} else {
				idleCount++
			}
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
			pendingCount++
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
			runningCount++
		}
	}

	// check if there are any parser errors
	if parser.State != nil && len(parser.State.ParseErrors) != 0 {
		parserErrorsCount++
	}

	return &projectStatusTableRow{
		Name:                project,
		Org:                 org,
		IdleCount:           idleCount,
		IdleWithErrorsCount: idleWithErrorsCount,
		PendingCount:        pendingCount,
		RunningCount:        runningCount,
		ParserErrorsCount:   parserErrorsCount,
	}, nil
}
