package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
)

func StatusCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, path string

	statusCmd := &cobra.Command{
		Use:   "status [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Project deployment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				name, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				OrganizationName: ch.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			// 1. Print project info
			ch.Printer.PrintlnSuccess("Project info\n")
			fmt.Printf("  Name: %s\n", proj.Project.Name)
			fmt.Printf("  Organization: %v\n", proj.Project.OrgName)
			fmt.Printf("  Public: %v\n", proj.Project.Public)
			fmt.Printf("  Github: %v\n", proj.Project.GithubUrl)
			fmt.Printf("  Created: %s\n", proj.Project.CreatedOn.AsTime().Local().Format(time.RFC3339))
			fmt.Printf("  Updated: %s\n", proj.Project.UpdatedOn.AsTime().Local().Format(time.RFC3339))

			depl := proj.ProdDeployment
			if depl == nil {
				return nil
			}

			// 2. Print deployment info
			ch.Printer.PrintlnSuccess("\nDeployment info\n")
			fmt.Printf("  Web: %s\n", proj.Project.FrontendUrl)
			fmt.Printf("  Runtime: %s\n", depl.RuntimeHost)
			fmt.Printf("  Instance: %s\n", depl.RuntimeInstanceId)
			fmt.Printf("  Driver: %s\n", proj.Project.ProdOlapDriver)
			if proj.Project.ProdOlapDsn != "" {
				fmt.Printf("  OLAP DSN: %s\n", proj.Project.ProdOlapDsn)
			}
			fmt.Printf("  Slots: %d\n", depl.Slots)
			fmt.Printf("  Branch: %s\n", depl.Branch)
			if proj.Project.Subpath != "" {
				fmt.Printf("  Subpath: %s\n", proj.Project.Subpath)
			}
			fmt.Printf("  Created: %s\n", depl.CreatedOn.AsTime().Local().Format(time.RFC3339))
			fmt.Printf("  Updated: %s\n", depl.UpdatedOn.AsTime().Local().Format(time.RFC3339))
			if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_OK {
				fmt.Printf("  Status: %s\n", depl.Status.String())
				fmt.Printf("  Status Message: %s\n", depl.StatusMessage)

				// Deployment not available
				return nil
			}

			// 3. Print parser and resources info
			rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}

			res, err := rt.ListResources(cmd.Context(), &runtimev1.ListResourcesRequest{InstanceId: depl.RuntimeInstanceId})
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			var parser *runtimev1.ProjectParser
			var table []*resourceTableRow

			for _, r := range res.Resources {
				if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
					parser = r.GetProjectParser()
				}
				if r.Meta.Hidden {
					continue
				}

				table = append(table, newResourceTableRow(r))
			}

			ch.Printer.PrintlnSuccess("\nResources\n")
			ch.Printer.PrintData(table)

			if parser.State != nil && len(parser.State.ParseErrors) != 0 {
				var table []*parseErrorTableRow
				for _, e := range parser.State.ParseErrors {
					table = append(table, newParseErrorTableRow(e))
				}

				ch.Printer.PrintlnSuccess("\nParse errors\n")
				ch.Printer.PrintData(table)
			}

			return nil
		},
	}

	statusCmd.Flags().StringVar(&name, "project", "", "Project Name")
	statusCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return statusCmd
}

type resourceTableRow struct {
	Kind   string `header:"kind"`
	Name   string `header:"name"`
	Status string `header:"status"`
	Error  string `header:"error"`
}

func newResourceTableRow(r *runtimev1.Resource) *resourceTableRow {
	truncErr := r.Meta.ReconcileError
	if len(truncErr) > 80 {
		truncErr = truncErr[:80] + "..."
	}

	return &resourceTableRow{
		Kind:   formatResourceKind(r.Meta.Name.Kind),
		Name:   r.Meta.Name.Name,
		Status: formatReconcileStatus(r.Meta.ReconcileStatus),
		Error:  truncErr,
	}
}

func formatResourceKind(k string) string {
	k = strings.TrimPrefix(k, "rill.runtime.v1.")
	k = strings.TrimSuffix(k, "V2")
	return k
}

func formatReconcileStatus(s runtimev1.ReconcileStatus) string {
	switch s {
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_UNSPECIFIED:
		return "Unknown"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
		return "Idle"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
		return "Pending"
	case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
		return "Running"
	default:
		panic(fmt.Errorf("unknown reconcile status: %s", s.String()))
	}
}

type parseErrorTableRow struct {
	Path  string `header:"path"`
	Error string `header:"error"`
}

func newParseErrorTableRow(e *runtimev1.ParseError) *parseErrorTableRow {
	return &parseErrorTableRow{
		Path:  e.FilePath,
		Error: e.Message,
	}
}
