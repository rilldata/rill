package org

import (
	"strings"

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
	orgCmd.AddCommand(SwitchCmd(cfg))
	orgCmd.AddCommand(ListCmd(cfg))
	orgCmd.AddCommand(DeleteCmd(cfg))
	orgCmd.AddCommand(RenameCmd(cfg))

	return orgCmd
}

func toTable(organizations []*adminv1.Organization, defaultOrg string) []*organization {
	orgs := make([]*organization, 0, len(organizations))

	for _, org := range organizations {
		if strings.EqualFold(org.Name, defaultOrg) {
			org.Name += " (default)"
		}
		orgs = append(orgs, toRow(org))
	}

	return orgs
}

func toRow(o *adminv1.Organization) *organization {
	return &organization{
		Name:      o.Name,
		CreatedAt: o.CreatedOn.AsTime().String(),
	}
}

type organization struct {
	Name      string `header:"name" json:"name"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}
