package project

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/spf13/cobra"
)

func StatusCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, path string
	var local bool

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
			if local {
				if cmd.Flags().Changed("project") || cmd.Flags().Changed("path") {
					return fmt.Errorf("the --local flag cannot be used with --project or --path")
				}
				if len(args) > 0 {
					return fmt.Errorf("the --local flag cannot be used with <project-name> positional argument")
				}
			} else {
				if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
					name, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
					if err != nil {
						return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
					}
				}
				// Project info and deployment info not available --local mode
				proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
					Org:     ch.Org,
					Project: name,
				})
				if err != nil {
					return err
				}

				var gitRemote string
				if proj.Project.ManagedGitId == "" {
					gitRemote = proj.Project.GitRemote
				}

				// 1. Print project info
				ch.PrintfSuccess("Project info\n\n")
				fmt.Printf("  Name: %s\n", proj.Project.Name)
				fmt.Printf("  Organization: %v\n", proj.Project.OrgName)
				fmt.Printf("  Public: %v\n", proj.Project.Public)
				fmt.Printf("  Git: %v\n", gitRemote)
				fmt.Printf("  Created: %s\n", proj.Project.CreatedOn.AsTime().Local().Format(time.RFC3339))
				fmt.Printf("  Updated: %s\n", proj.Project.UpdatedOn.AsTime().Local().Format(time.RFC3339))

				depl := proj.ProdDeployment
				if depl == nil {
					return nil
				}

				// 2. Print deployment info
				ch.PrintfSuccess("\nDeployment info\n\n")
				fmt.Printf("  Web: %s\n", proj.Project.FrontendUrl)
				fmt.Printf("  Runtime: %s\n", depl.RuntimeHost)
				fmt.Printf("  Instance: %s\n", depl.RuntimeInstanceId)
				fmt.Printf("  Slots: %d\n", proj.Project.ProdSlots)
				fmt.Printf("  Branch: %s\n", depl.Branch)
				if proj.Project.Subpath != "" {
					fmt.Printf("  Subpath: %s\n", proj.Project.Subpath)
				}
				fmt.Printf("  Created: %s\n", depl.CreatedOn.AsTime().Local().Format(time.RFC3339))
				fmt.Printf("  Updated: %s\n", depl.UpdatedOn.AsTime().Local().Format(time.RFC3339))
				if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
					fmt.Printf("  Status: %s\n", depl.Status.String())
					fmt.Printf("  Status Message: %s\n", depl.StatusMessage)

					// Deployment not available
					return nil
				}
				fmt.Println("")
			}

			// 3. Print parser and resources info
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, name, local)
			if err != nil {
				return err
			}

			res, err := rt.ListResources(cmd.Context(), &runtimev1.ListResourcesRequest{
				InstanceId:         instanceID,
				SkipSecurityChecks: true,
			})
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			var parser *runtimev1.Resource
			var table []*resourceTableRow

			for _, r := range res.Resources {
				if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
					parser = r
				}
				if r.Meta.Hidden {
					continue
				}

				table = append(table, newResourceTableRow(r))
			}

			ch.PrintfSuccess("Resources\n\n")
			ch.PrintData(table)

			if parser != nil {
				state := parser.GetProjectParser().State

				var table []*parseErrorTableRow
				if parser.Meta.ReconcileError != "" {
					table = append(table, &parseErrorTableRow{
						Path:  "<meta>",
						Error: parser.Meta.ReconcileError,
					})
				}
				if state != nil {
					for _, e := range state.ParseErrors {
						table = append(table, &parseErrorTableRow{
							Path:  e.FilePath,
							Error: e.Message,
						})
					}
				}

				if len(table) > 0 {
					ch.PrintfSuccess("\nParse errors\n\n")
					ch.PrintData(table)
				}
			}

			return nil
		},
	}

	statusCmd.Flags().StringVar(&name, "project", "", "Project Name")
	statusCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	statusCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")

	return statusCmd
}

type resourceTableRow struct {
	Type   string `header:"type"`
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
		Type:   runtime.PrettifyResourceKind(r.Meta.Name.Kind),
		Name:   r.Meta.Name.Name,
		Status: runtime.PrettifyReconcileStatus(r.Meta.ReconcileStatus),
		Error:  truncErr,
	}
}

type parseErrorTableRow struct {
	Path  string `header:"path"`
	Error string `header:"error"`
}
