package org

import (
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func OrgCmd(ch *cmdutil.Helper) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:               "org",
		Short:             "Manage organisations",
		PersistentPreRunE: cmdutil.CheckAuth(ch.Config),
	}

	orgCmd.AddCommand(CreateCmd(ch))
	orgCmd.AddCommand(EditCmd(ch))
	orgCmd.AddCommand(SwitchCmd(ch))
	orgCmd.AddCommand(ListCmd(ch))
	orgCmd.AddCommand(DeleteCmd(ch))
	orgCmd.AddCommand(RenameCmd(ch))

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
		CreatedAt: o.CreatedOn.AsTime().Format(cmdutil.TSFormatLayout),
	}
}

type organization struct {
	Name      string `header:"name" json:"name"`
	CreatedAt string `header:"created_at,timestamp(ms|utc|human)" json:"created_at"`
}
