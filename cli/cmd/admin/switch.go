package admin

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

const (
	prodAdminURL    = "https://admin.rilldata.com"
	stagingAdminURL = "https://admin.rilldata.io"
)

func SwitchCmd(cfg *config.Config) *cobra.Command {
	var env string
	switchCmd := &cobra.Command{
		Use:    "switch {stage|prod}",
		Short:  "switch",
		Args:   cobra.MaximumNArgs(1),
		Hidden: !cfg.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				env = args[0]
			}

			if len(args) == 0 && cfg.Interactive {
				env = cmdutil.SelectPrompt("Select org to rename", []string{"prod", "stage"}, "")
			}

			var url string

			switch env {
			case "prod":
				url = prodAdminURL
			case "stage":
				url = stagingAdminURL
			default:
				url = cfg.AdminURL
			}

			err := dotrill.SetDefaultAdminURL(url)
			if err != nil {
				return err
			}
			cfg.AdminURL = url

			cmdutil.SuccessPrinter(fmt.Sprintf("Set default env to %q, url is %q", env, url))

			return nil
		},
	}

	return switchCmd
}
