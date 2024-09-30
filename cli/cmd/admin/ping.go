package admin

import (
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func PingCmd(ch *cmdutil.Helper) *cobra.Command {
	var adminURL string

	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Must set here to avoid flag parser overriding it globally
			ch.AdminURLOverride = adminURL

			client, err := ch.Client()
			if err != nil {
				return err
			}

			ch.Printf("Ping: %s\n", time.Now().UTC().String())

			pong, err := client.Ping(cmd.Context(), &adminv1.PingRequest{})
			if err != nil {
				return err
			}

			ch.Printf("Pong: %s\n", pong.Time.AsTime().String())
			return nil
		},
	}

	pingCmd.Flags().StringVar(&adminURL, "url", "https://admin.rilldata.com", "Base URL for the admin API")

	return pingCmd
}
