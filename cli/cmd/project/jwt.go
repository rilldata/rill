package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func JwtCmd(cfg *config.Config) *cobra.Command {
	jwtCmd := &cobra.Command{
		Use:    "jwt <project>",
		Args:   cobra.ExactArgs(1),
		Short:  "Generate token for connecting directly to the deployment",
		Hidden: !cfg.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             args[0],
			})
			if err != nil {
				return err
			}
			if res.ProdDeployment == nil {
				cmdutil.WarnPrinter("Project does not have a production deployment")
				return nil
			}

			cmdutil.SuccessPrinter("Runtime info")
			fmt.Printf("  Host: %s\n", res.ProdDeployment.RuntimeHost)
			fmt.Printf("  Instance: %s\n", res.ProdDeployment.RuntimeInstanceId)
			fmt.Printf("  JWT: %s\n", res.Jwt)

			return nil
		},
	}

	return jwtCmd
}
