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

			backupToken, err := dotrill.GetBackupToken()
			if err != nil {
				return err
			}

			if backupToken != "" {
				return fmt.Errorf("Can't switch environment when assuming another user. Run `rill sudo user unassume` and try again")
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

			err = switchEnvTokens(env, cfg)
			if err != nil {
				return err
			}

			err = dotrill.SetDefaultAdminURL(url)
			if err != nil {
				return err
			}

			cfg.AdminURL = url

			cmdutil.PrintlnSuccess(fmt.Sprintf("Set default env to %q, url is %q", env, url))
			err = auth.SelectOrgFlow(cmd.Context(), cfg)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return switchCmd
}

func switchEnvTokens(env string, cfg *config.Config) error {
	token, err := dotrill.GetAccessToken()
	if err != nil {
		return err
	}

	switch cfg.AdminURL {
	case prodAdminURL:
		err := dotrill.SetEnvToken("prod", token)
		if err != nil {
			return err
		}
	case stagingAdminURL:
		err := dotrill.SetEnvToken("stage", token)
		if err != nil {
			return err
		}
	case devAdminURL:
		err := dotrill.SetEnvToken("dev", token)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid admin url")
	}

	newToken, err := dotrill.GetEnvToken(env)
	if err != nil {
		return err
	}

	err = dotrill.SetAccessToken(newToken)
	if err != nil {
		return err
	}

	// set the default token to the one we just got
	cfg.AdminTokenDefault = newToken

	return nil
}
