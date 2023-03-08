package org

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func OrgCmd(cfg *config.Config) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:               "org",
		Hidden:            !cfg.IsDev(),
		Short:             "Manage organisations",
		PersistentPreRunE: cmdutil.CheckAuth(cfg),
	}
	orgCmd.AddCommand(CreateCmd(cfg))
	orgCmd.AddCommand(EditCmd(cfg))
	orgCmd.AddCommand(ShowCmd(cfg))
	orgCmd.AddCommand(CloseCmd(cfg))
	orgCmd.AddCommand(InviteCmd(cfg))
	orgCmd.AddCommand(MembersCmd(cfg))
	orgCmd.AddCommand(SwitchCmd(cfg))
	orgCmd.AddCommand(ListCmd(cfg))
	orgCmd.AddCommand(DeleteCmd(cfg))

	return orgCmd
}

func toOrgs(organizations []*adminv1.Organization) []*organization {
	orgs := make([]*organization, 0, len(organizations))

	for _, org := range organizations {
		orgs = append(orgs, toOrg(org))
	}

	return orgs
}

func toOrg(o *adminv1.Organization) *organization {
	return &organization{
		ID:          o.Id,
		Name:        o.Name,
		Description: o.Description,
		CreatedAt:   o.CreatedOn.AsTime().String(),
		UpdatedAt:   o.UpdatedOn.AsTime().String(),
	}
}

type organization struct {
	ID          string `header:"id" json:"id"`
	Name        string `header:"name" json:"name"`
	Description string `header:"description" json:"description"`
	CreatedAt   string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
	UpdatedAt   string `header:"updated_at,timestamp(ms|utc|human)" json:"updated_at"`
}
