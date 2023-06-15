package admin

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

const (
	prodAdminURL    = "https://admin.rilldata.com"
	stagingAdminURL = "https://admin.rilldata.io"
	devAdminURL     = "http://localhost:9090"
)

func SwitchCmd(cfg *config.Config) *cobra.Command {
	var env string
	switchCmd := &cobra.Command{
		Use:   "switch {stage|prod|dev}",
		Short: "switch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				env = args[0]
			}

			var url string

			switch env {
			case "prod":
				url = prodAdminURL
			case "stage":
				url = stagingAdminURL
			case "dev":
				url = devAdminURL
			default:
				return fmt.Errorf("invalid args provided, valid args are {stage|prod|dev}")
			}

			err := dotrill.SetDefaultAdminURL(url)
			if err != nil {
				return err
			}

			err = auth.Logout(cmd.Context(), cfg)
			if err != nil {
				return err
			}

			cfg.AdminURL = url

			cmdutil.PrintlnSuccess(fmt.Sprintf("Set default env to %q, url is %q", env, url))

			return nil
		},
	}

	return switchCmd
}
