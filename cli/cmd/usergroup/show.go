package usergroup

import (
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var name string

	showCmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			name = args[0]

			res, err := client.GetUsergroup(cmd.Context(), &adminv1.GetUsergroupRequest{
				Org:       ch.Org,
				Usergroup: name,
			})
			if err != nil {
				return err
			}

			ch.Printf("User group info\n")
			ch.Printf("  Name: %s\n", res.Usergroup.GroupName)
			ch.Printf("  Description: %s\n", res.Usergroup.GroupDescription)
			ch.Printf("  Created on: %s\n", res.Usergroup.CreatedOn.AsTime().Local().Format(time.DateTime))
			ch.Printf("  Updated on: %s\n", res.Usergroup.UpdatedOn.AsTime().Local().Format(time.DateTime))

			return nil
		},
	}

	showCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return showCmd
}
