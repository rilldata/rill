package service

import (
	"time"

	"github.com/rilldata/rill/cli/cmd/service/token"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ServiceCmd(ch *cmdutil.Helper) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:               "service",
		Short:             "Manage service accounts",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	serviceCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")

	serviceCmd.AddCommand(ListCmd(ch))
	serviceCmd.AddCommand(CreateCmd(ch))
	serviceCmd.AddCommand(RenameCmd(ch))
	serviceCmd.AddCommand(DeleteCmd(ch))
	serviceCmd.AddCommand(token.TokenCmd(ch))

	return serviceCmd
}

func toTable(sv []*adminv1.Service) []*service {
	services := make([]*service, 0, len(sv))

	for _, s := range sv {
		services = append(services, toRow(s))
	}

	return services
}

func toRow(s *adminv1.Service) *service {
	return &service{
		Name:      s.Name,
		OrgName:   s.OrgName,
		CreatedAt: s.CreatedOn.AsTime().Format(time.DateTime),
	}
}

type service struct {
	Name      string `header:"name" json:"name"`
	OrgName   string `header:"org_name" json:"org_name"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}
