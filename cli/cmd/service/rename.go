package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var newName string

	renameCmd := &cobra.Command{
		Use:   "rename <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Rename service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			req := &adminv1.UpdateServiceRequest{
				Name:             args[0],
				OrganizationName: cfg.Org,
			}

			if newName != "" {
				req.NewName = &newName
			}

			res, err := client.UpdateService(cmd.Context(), req)
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess("Renamed service")
			return ch.Printer.PrintResource([]*service{toRow(res.Service)})
		},
	}
	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Service Name")

	return renameCmd
}
