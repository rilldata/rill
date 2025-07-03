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

			res, err := client.ShowService(cmd.Context(), &adminv1.ShowServiceRequest{
				Name:             args[0],
				OrganizationName: ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Service Details:\n")
			ch.Println(" Name: ", res.OrgService.Name)
			ch.Println(" Org Name: ", res.OrgService.OrgName)
			ch.Println(" Org Role: ", res.OrgService.RoleName)
			ch.Print(" Attributes: ")
			attrBytes, err := json.Marshal(res.OrgService.Attributes)
			if err != nil {
				panic(fmt.Errorf("failed to marshal service attributes: %w", err))
			}
			ch.Println(string(attrBytes))

			ch.Println(" Created On: ", res.OrgService.CreatedOn.AsTime().Format("2006-01-02 15:04:05"))
			ch.Println(" Updated On: ", res.OrgService.UpdatedOn.AsTime().Format("2006-01-02 15:04:05"))

			ch.Printf("\n")
			if len(res.ProjectServices) > 0 {
				ch.PrintfSuccess("Project Memberships:\n")
				ch.PrintProjectMemberServices(res.ProjectServices)
			}

			return nil
		},
	}
	return showCmd
}
