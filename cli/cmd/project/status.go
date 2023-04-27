package project

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func StatusCmd(cfg *config.Config) *cobra.Command {
	var name, path string

	statusCmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Short: "Project deployment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("project") {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Found project")
			cmdutil.TablePrinter(toRow(proj.Project))

			depl := proj.ProdDeployment
			if depl != nil {
				logs, err := logsFormatter(depl.Logs)
				if err != nil {
					logs = fmt.Sprintf("  Logs: %s\n\n", depl.Logs)
				}

				cmdutil.SuccessPrinter("Deployment info")
				fmt.Printf("  Runtime: %s\n", depl.RuntimeHost)
				fmt.Printf("  Instance: %s\n", depl.RuntimeInstanceId)
				fmt.Printf("  Slots: %d\n", depl.Slots)
				fmt.Printf("  Branch: %s\n", depl.Branch)
				fmt.Printf("  Status: %s\n", depl.Status.String())
				if proj.ProjectPermissions.ReadProdStatus {
					fmt.Println(logs)
				}
			}

			return nil
		},
	}

	statusCmd.Flags().StringVar(&name, "project", "", "Name")
	statusCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return statusCmd
}

func logsFormatter(jsonStr string) (string, error) {
	res := runtimev1.ReconcileResponse{}
	err := protojson.Unmarshal([]byte(jsonStr), &res)
	if err != nil {
		return "", fmt.Errorf("error in reconcileResponse logs formatting, Error %w", err)
	}

	var errors []string
	for i := range res.Errors {
		errors = append(errors, res.Errors[i].String())
	}

	var logs []string
	if len(errors) != 0 {
		logs = append(logs, fmt.Sprintf("  Errors:\n\t%s", strings.Join(errors, "\n\t")))
	}

	if len(res.AffectedPaths) != 0 {
		logs = append(logs, fmt.Sprintf("  Affected paths:\n\t%s", strings.Join(res.AffectedPaths, "\n\t")))
	}
	return strings.Join(logs, "\n"), nil
}
