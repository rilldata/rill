package runtime

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func ListInstancesCmd(ch *cmdutil.Helper) *cobra.Command {
	var audience string
	var pageSize uint32
	var pageToken string

	cmd := &cobra.Command{
		Use:   "list-instances <host>",
		Args:  cobra.ExactArgs(1),
		Short: "Lists full details about the instances on a runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse args
			host := args[0]
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

			// List instances
			res, err := rt.ListInstances(cmd.Context(), &runtimev1.ListInstancesRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			// Pretty print as JSON
			enc := protojson.MarshalOptions{
				Multiline:       true,
				EmitUnpopulated: true,
			}
			data, err := enc.Marshal(res)
			if err != nil {
				return fmt.Errorf("failed to marshal as JSON: %w", err)
			}
			fmt.Println(string(data))

			return nil
		},
	}

	cmd.Flags().StringVar(&audience, "audience", "", "Override JWT audience if it differs from the host")
	cmd.Flags().Uint32Var(&pageSize, "page-size", 100, "Number of instances per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return cmd
}
