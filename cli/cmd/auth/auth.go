package auth

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func AuthCmd(ver version.Version) *cobra.Command {
	authCmd := &cobra.Command{
		Use:    "auth",
		Hidden: !ver.IsDev(),
		Short:  "Manage authentication",
	}
	authCmd.AddCommand(LoginCmd(ver))
	authCmd.AddCommand(LogoutCmd(ver))

	return authCmd
}
