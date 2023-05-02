package admin

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func PingCmd(cfg *config.Config) *cobra.Command {
	var adminURL string

	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			pong, err := client.Ping(context.Background(), &adminv1.PingRequest{})
			if err != nil {
				return err
			}

			fmt.Printf("Pong: %s\n", pong.Time.AsTime().String())
			return nil
		},
	}

	pingCmd.Flags().StringVar(&adminURL, "base-url", "https://admin.rilldata.io", "Base URL for the admin API")

	return pingCmd
}
