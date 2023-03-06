package project

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var displayName string
	createCmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.ExactArgs(1),
		Short: "Create",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.GetSpinner(4, "Creating project...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(1 * time.Second)

			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.CreateProject(context.Background(), &adminv1.CreateProjectRequest{
				Organization: cfg.DefaultOrg,
				Name:         args[0],
				Description:  displayName,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Created project: %v\n", proj)
			sp.Stop()
			return nil
		},
	}

	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")

	return createCmd
}
