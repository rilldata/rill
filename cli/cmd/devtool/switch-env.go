package devtool

import (
	"context"
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

func SwitchEnvCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch-env [prod|stage|dev]",
		Short: "Switch between environments",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config

			backupToken, err := dotrill.GetBackupToken()
			if err != nil {
				return err
			}
			if backupToken != "" {
				return fmt.Errorf("can't switch environment when assuming another user (run `rill sudo user unassume` and try again)")
			}

			fromEnv, err := inferEnv(cfg)
			if err != nil {
				return err
			}

			var toEnv string
			if len(args) > 0 {
				toEnv = args[0]
			} else {
				toEnv = cmdutil.SelectPrompt("Select environment", []string{"prod", "stage", "dev"}, fromEnv)
			}

			err = switchEnv(cfg, fromEnv, toEnv)
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess(fmt.Sprintf("Set default env to %q (%q)", toEnv, adminURLForEnv(toEnv)))

			return auth.SelectOrgFlow(cmd.Context(), ch, true)
		},
	}

	return cmd
}

func inferEnv(cfg *config.Config) (string, error) {
	switch cfg.AdminURL {
	case prodAdminURL:
		return "prod", nil
	case stagingAdminURL:
		return "stage", nil
	case devAdminURL:
		return "dev", nil
	default:
		return "", fmt.Errorf("could not infer env from admin URL %q", cfg.AdminURL)
	}
}

func adminURLForEnv(env string) string {
	switch env {
	case "prod":
		return prodAdminURL
	case "stage":
		return stagingAdminURL
	case "dev":
		return devAdminURL
	default:
		panic(fmt.Errorf("invalid environment %q", env))
	}
}

func switchEnv(cfg *config.Config, fromEnv, toEnv string) error {
	token, err := dotrill.GetAccessToken()
	if err != nil {
		return err
	}

	err = dotrill.SetEnvToken(fromEnv, token)
	if err != nil {
		return err
	}

	toToken, err := dotrill.GetEnvToken(toEnv)
	if err != nil {
		return err
	}

	err = dotrill.SetAccessToken(toToken)
	if err != nil {
		return err
	}
	cfg.AdminTokenDefault = toToken // Also set the cfg's token to the one we just got

	toURL := adminURLForEnv(toEnv)
	err = dotrill.SetDefaultAdminURL(toURL)
	if err != nil {
		return err
	}
	cfg.AdminURL = toURL

	return nil
}

func switchEnvToDevTemporarily(ctx context.Context, ch *cmdutil.Helper) {
	env, err := inferEnv(ch.Config)
	if err != nil {
		logWarn.Printf("Did not switch CLI to dev environment: failed to infer environment (error: %v)\n", err)
		return
	}

	if env == "dev" {
		logInfo.Printf("CLI already configured for dev environment\n")
		return
	}

	err = switchEnv(ch.Config, env, "dev")
	if err != nil {
		logWarn.Printf("Did not switch CLI to dev environment: failed to switch environment (error: %v)\n", err)
		return
	}

	logInfo.Printf("Switched CLI to dev environment\n")

	// TODO: Clear token if not authenticated

	prevOrg := ch.Config.Org
	err = auth.SelectOrgFlow(ctx, ch, false)
	if err != nil {
		logWarn.Printf("Failed to select org in dev environment: %v\n", err)
	}

	<-ctx.Done()

	err = switchEnv(ch.Config, "dev", env)
	if err != nil {
		logErr.Printf("Failed to switch CLI back to %s environment: %v\n", env, err)
		return
	}

	logInfo.Printf("Switched CLI back to to %s environment\n", env)

	err = dotrill.SetDefaultOrg(prevOrg)
	if err != nil {
		logErr.Printf("Failed to set default org back to %q: %v\n", prevOrg, err)
		return
	}
	ch.Config.Org = prevOrg
}
