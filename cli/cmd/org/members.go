package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func MembersCmd(ver version.Version) *cobra.Command {
	membersCmd := &cobra.Command{
		Use:   "members",
		Short: "Members",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return membersCmd
}
