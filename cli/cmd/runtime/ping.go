package runtime

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
)

func PingCmd(cfg *config.Config) *cobra.Command {
	var runtimeURL string

	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(runtimeURL, cfg.AdminToken)
			if err != nil {
				return err
			}
			defer client.Close()

			pong, err := client.Ping(context.Background(), &runtimev1.PingRequest{})
			if err != nil {
				return err
			}

			fmt.Printf("Pong: %s\n", pong.Time.AsTime().String())
			return nil
		},
	}

	pingCmd.Flags().StringVar(&runtimeURL, "base-url", "http://localhost:9010", "Base URL for the runtime")

	return pingCmd
}
