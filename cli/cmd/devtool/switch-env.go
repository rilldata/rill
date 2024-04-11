package devtool

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
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
			backupToken, err := dotrill.GetBackupToken()
			if err != nil {
				return err
			}
			if backupToken != "" {
				return fmt.Errorf("can't switch environment when assuming another user (run `rill sudo user unassume` and try again)")
			}

			fromEnv, err := inferEnv(ch)
			if err != nil {
				return err
			}

			var toEnv string
			if len(args) > 0 {
				toEnv = args[0]
			} else {
				toEnv, err = cmdutil.SelectPrompt("Select environment", maps.Keys(envURLs), fromEnv)
				if err != nil {
					return err
				}
			}

			err = switchEnv(ch, fromEnv, toEnv)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Set default env to %q (%q)\n", toEnv, adminURLForEnv(toEnv))

			return auth.SelectOrgFlow(cmd.Context(), ch, true)
		},
	}

	return cmd
}

var envURLs = map[string]string{
	"prod":  "https://admin.rilldata.com",
	"stage": "https://admin.rilldata.io",
	"test":  "https://admin.rilldata.in",
	"dev":   "http://localhost:9090",
}

func inferEnv(ch *cmdutil.Helper) (string, error) {
	for env, url := range envURLs {
		if url == ch.AdminURL {
			return env, nil
		}
	}
	return "", fmt.Errorf("could not infer env from admin URL %q", ch.AdminURL)
}

func adminURLForEnv(env string) string {
	u, ok := envURLs[env]
	if !ok {
		panic(fmt.Errorf("invalid environment %q", env))
	}
	return u
}

func switchEnv(ch *cmdutil.Helper, fromEnv, toEnv string) error {
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
	ch.AdminTokenDefault = toToken // Also set the cfg's token to the one we just got

	toURL := adminURLForEnv(toEnv)
	err = dotrill.SetDefaultAdminURL(toURL)
	if err != nil {
		return err
	}
	ch.AdminURL = toURL

	return nil
}

// switchEnvToDevTemporarily switches the CLI to the "dev" environment (if not already there),
// and then switches it back and returns when the context is cancelled.
func switchEnvToDevTemporarily(ctx context.Context, ch *cmdutil.Helper) {
	env, err := inferEnv(ch)
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
		err = auth.SelectOrgFlow(ctx, ch, false)
		if err != nil {
			logWarn.Printf("Failed to select org in dev environment: %v\n", err)
		}
	} else {
		// Since dev environments are frequently reset, clear the token if it's invalid
		_ = dotrill.SetAccessToken("")
		_ = dotrill.SetDefaultOrg("")
		ch.AdminTokenDefault = ""
	}

	// Wait for ctx cancellation, then switch back to the previous environment before returning.
	<-ctx.Done()

	err = switchEnv(ch, "dev", env)
	if err != nil {
		logErr.Printf("Failed to switch CLI back to %s environment: %v\n", env, err)
		return
	}

	logInfo.Printf("Switched CLI back to %s environment\n", env)

	err = dotrill.SetDefaultOrg(prevOrg)
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
