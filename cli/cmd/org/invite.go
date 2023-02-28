package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func InviteCmd(ver version.Version) *cobra.Command {
	inviteCmd := &cobra.Command{
		Use:   "invite",
		Args:  cobra.ExactArgs(1),
		Short: "Invite",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	return inviteCmd
}
