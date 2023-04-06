package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func JwtCmd(cfg *config.Config) *cobra.Command {
	jwtCmd := &cobra.Command{
		Use:    "jwt",
		Args:   cobra.ExactArgs(1),
		Short:  "Jwt",
		Hidden: true,
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

			cmdutil.TextPrinter("Runtime info\n")
			fmt.Printf("  Host: %s\n", proj.ProductionDeployment.RuntimeHost)
			fmt.Printf("  Instance: %s\n", proj.ProductionDeployment.RuntimeInstanceId)
			fmt.Printf("  JWT: %s\n", proj.Jwt)

			return nil
		},
	}

	return jwtCmd
}
