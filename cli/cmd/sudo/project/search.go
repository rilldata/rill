package project

import (
	"fmt"
	"strings"

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

			err = ch.Printer.PrintResource(res.Names)
			if err != nil {
				return err
			}

			if status {
				var table []*projectStatusTableRow
				ch.Printer.Println()
				for _, name := range res.Names {
					project := strings.Split(name, "/")[0]
					org := strings.Split(name, "/")[1]

					proj, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
						OrganizationName: org,
						Name:             project,
					})
					if err != nil {
						return err
					}

					depl := proj.ProdDeployment
					if depl == nil {
						continue
					}

					rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
					if err != nil {
						return fmt.Errorf("failed to connect to runtime: %w", err)
					}

					res, err := rt.ListResources(cmd.Context(), &runtimev1.ListResourcesRequest{InstanceId: depl.RuntimeInstanceId})
					if err != nil {
						return fmt.Errorf("failed to list resources: %w", err)
					}

					var parser *runtimev1.ProjectParser
					var ParserErrorCount int32
					var IdleCount int32
					var IdleWithErrorsCount int32
					var PendingCount int32
					var RunningCount int32

					for _, r := range res.Resources {
						if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
							parser = r.GetProjectParser()
						}
						if r.Meta.Hidden {
							continue
						}

						// check if there are any parser errors
						if parser.State != nil && len(parser.State.ParseErrors) != 0 {
							ParserErrorCount++
						}

						switch r.Meta.ReconcileStatus {
						case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
							// if it is idle, check if there are any errors
							if r.Meta.GetReconcileError() != "" {
								IdleWithErrorsCount++
							} else {
								IdleCount++
							}
						case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
							PendingCount++
						case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
							RunningCount++
						}
					}

					table = append(table, &projectStatusTableRow{
						Name:                name,
						Organization:        org,
						IdelCount:           IdleCount,
						IdelWithErrorsCount: IdleWithErrorsCount,
						PendingCount:        PendingCount,
						RunningCount:        RunningCount,
						ParserErrorCount:    ParserErrorCount,
					})
				}

				ch.Printer.PrintlnSuccess("\nProject status\n")
				err = ch.Printer.PrintResource(table)
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
	Organization        string `header:"organization"`
	IdelCount           int32  `header:"idle"`
	IdelWithErrorsCount int32  `header:"idle with errors"`
	PendingCount        int32  `header:"pending"`
	RunningCount        int32  `header:"running"`
	ParserErrorCount    int32  `header:"parser error"`
	Error               string `header:"error"`
}
