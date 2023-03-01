package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func MembersCmd(cfg *config.Config) *cobra.Command {
	membersCmd := &cobra.Command{
		Use:   "members",
		Short: "Members",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return membersCmd
}
