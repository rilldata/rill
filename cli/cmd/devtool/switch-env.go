package devtool

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/adminenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SwitchEnvCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch-env [prod|stage|test|dev]",
		Short: "Switch between admin environments",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupToken, err := ch.DotRill.GetBackupToken()
			if err != nil {
				return err
			}
			if backupToken != "" {
				return fmt.Errorf("can't switch environment when assuming another user (run `rill sudo user unassume` and try again)")
			}

			fromEnv, err := adminenv.Infer(ch.AdminURL())
			if err != nil {
				return err
			}

			var toEnv string
			if len(args) > 0 {
				toEnv = args[0]
			} else {
				toEnv, err = cmdutil.SelectPrompt("Select environment", maps.Keys(adminenv.EnvURLs), fromEnv)
				if err != nil {
					return err
				}
			}

			err = switchEnv(ch, fromEnv, toEnv)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Set default env to %q (%q)\n", toEnv, adminenv.AdminURL(toEnv))

			return auth.SelectOrgFlow(cmd.Context(), ch, ch.Interactive, "")
		},
	}

	return cmd
}

func switchEnv(ch *cmdutil.Helper, fromEnv, toEnv string) error {
	token, err := ch.DotRill.GetAccessToken()
	if err != nil {
		return err
	}

	err = ch.DotRill.SetEnvToken(fromEnv, token)
	if err != nil {
		return err
	}

	toToken, err := ch.DotRill.GetEnvToken(toEnv)
	if err != nil {
		return err
	}

	err = ch.DotRill.SetAccessToken(toToken)
	if err != nil {
		return err
	}

	toURL := adminenv.AdminURL(toEnv)
	err = ch.DotRill.SetDefaultAdminURL(toURL)
	if err != nil {
		return err
	}

	err = ch.ReloadAdminConfig()
	if err != nil {
		return err
	}

	return nil
}

// switchEnvToDevTemporarily switches the CLI to the "dev" environment (if not already there),
// and then switches it back and returns when the context is cancelled.
func switchEnvToDevTemporarily(ctx context.Context, ch *cmdutil.Helper) {
	env, err := adminenv.Infer(ch.AdminURL())
	if err != nil {
		logWarn.Printf("Did not switch CLI to dev environment: failed to infer environment (error: %v)\n", err)
		return
	}

	if env == "dev" {
		logInfo.Printf("CLI already configured for dev environment\n")
		return
	}

	err = switchEnv(ch, env, "dev")
	if err != nil {
		logWarn.Printf("Did not switch CLI to dev environment: failed to switch environment (error: %v)\n", err)
		return
	}

	logInfo.Printf("Switched CLI to dev environment\n")

	authenticated, err := checkAuthenticated(ctx, ch)
	if err != nil {
		logErr.Printf("Failed to check if authenticated: %v\n", err)
	}

	var prevOrg string
	if authenticated {
		prevOrg = ch.Org
		err = auth.SelectOrgFlow(ctx, ch, false, "")
		if err != nil {
			logWarn.Printf("Failed to select org in dev environment: %v\n", err)
		}
	} else {
		// Since dev environments are frequently reset, clear the token if it's invalid
		_ = ch.DotRill.SetAccessToken("")
		_ = ch.ReloadAdminConfig()

		_ = ch.DotRill.SetDefaultOrg("")
		ch.Org = ""
	}

	// Wait for ctx cancellation, then switch back to the previous environment before returning.
	<-ctx.Done()

	err = switchEnv(ch, "dev", env)
	if err != nil {
		logErr.Printf("Failed to switch CLI back to %s environment: %v\n", env, err)
		return
	}

	logInfo.Printf("Switched CLI back to %s environment\n", env)

	err = ch.DotRill.SetDefaultOrg(prevOrg)
	if err != nil {
		logErr.Printf("Failed to set default org back to %q: %v\n", prevOrg, err)
		return
	}
	ch.Org = prevOrg
}

func checkAuthenticated(ctx context.Context, ch *cmdutil.Helper) (bool, error) {
	client, err := ch.Client()
	if err != nil {
		return false, err
	}

	res, err := client.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.Unauthenticated {
			return false, nil
		}

		return false, err
	}

	return res.User != nil, nil
}
