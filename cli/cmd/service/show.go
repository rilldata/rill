package service

import (
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show <service-name>",
		Short: "Show service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetService(cmd.Context(), &adminv1.GetServiceRequest{
				Name: args[0],
				Org:  ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Service Details:\n")
			ch.Println(" Name: ", res.Service.Name)
			ch.Println(" Org Name: ", res.Service.OrgName)
			ch.Println(" Org Role: ", res.Service.RoleName)
			ch.Print(" Attributes: ")
			attrBytes, err := json.Marshal(res.Service.Attributes)
			if err != nil {
				panic(fmt.Errorf("failed to marshal service attributes: %w", err))
			}
			ch.Println(string(attrBytes))

			ch.Println(" Created On: ", res.Service.CreatedOn.AsTime().Format("2006-01-02 15:04:05"))
			ch.Println(" Updated On: ", res.Service.UpdatedOn.AsTime().Format("2006-01-02 15:04:05"))

			ch.Printf("\n")
			if len(res.ProjectMemberships) > 0 {
				ch.PrintfSuccess("Project Memberships:\n")
				ch.PrintProjectMemberServices(res.ProjectMemberships)
			}

			return nil
		},
	}
	return showCmd
}
