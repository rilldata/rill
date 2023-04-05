package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

type Logs struct {
	Errors        []string `json:"Errors"`
	AffectedPaths []string `json:"AffectedPaths"`
}

func StatusCmd(cfg *config.Config) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.ExactArgs(1),
		Short: "Status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             args[0],
			})
			if err != nil {
				return err
			}

			cmdutil.TextPrinter("Found project\n")
			cmdutil.TablePrinter(toRow(proj.Project))

			depl := proj.ProductionDeployment
			if depl != nil {
				logs, err := logsFormatter(depl.Logs)
				if err != nil {
					return err
				}

				cmdutil.TextPrinter("Deplyment info\n")
				fmt.Printf("  Runtime: %s\n", depl.RuntimeHost)
				fmt.Printf("  Instance: %s\n", depl.RuntimeInstanceId)
				fmt.Printf("  Slots: %d\n", depl.Slots)
				fmt.Printf("  Branch: %s\n", depl.Branch)
				fmt.Printf("  Status: %s\n", depl.Status.String())
				fmt.Printf("  Logs: %s\n\n", logs)
			}

			return nil
		},
	}

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

	logs := fmt.Sprintf("Errors:\n\t%s\n\n\tAffectedPaths:\n\t%s",
		strings.Join(errors, "\n\t"),
		strings.Join(res.AffectedPaths, "\n\t"))
	return logs, nil
}
