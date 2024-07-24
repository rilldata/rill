package runtime

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
)

func DeleteInstanceCmd(ch *cmdutil.Helper) *cobra.Command {
	var audience string

	cmd := &cobra.Command{
		Use:   "delete-instance <host> <instance_id>",
		Args:  cobra.ExactArgs(2),
		Short: "Forcefully deletes an instance on a runtime",
		Long:  "Forcefully deletes an instance on a runtime. Should only be used to clean up orphaned instances. Use `sudo project hibernate` or `project delete` to tear down healthy deployments.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse args
			host := args[0]
			instanceID := args[1]
			if audience == "" {
				audience = host
			}

			// Obtain a manager token for the host
			client, err := ch.Client()
			if err != nil {
				return err
			}
			tokenRes, err := client.SudoIssueRuntimeManagerToken(cmd.Context(), &adminv1.SudoIssueRuntimeManagerTokenRequest{
				Host: audience,
			})
			if err != nil {
				return err
			}
			token := tokenRes.Token

			// Open a connection to the runtime
			rt, err := runtimeclient.New(host, token)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}

			// Delete the instance
			_, err = rt.DeleteInstance(cmd.Context(), &runtimev1.DeleteInstanceRequest{InstanceId: instanceID})
			if err != nil {
				return fmt.Errorf("failed to delete instance: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&audience, "audience", "", "Override JWT audience if it differs from the host")

	return cmd
}
