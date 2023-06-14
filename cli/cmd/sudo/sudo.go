package sudo

import (
	"github.com/rilldata/rill/cli/cmd/sudo/superuser"
	"github.com/rilldata/rill/cli/cmd/sudo/user"
	"github.com/rilldata/rill/cli/cmd/sudo/whitelist"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func SudoCmd(cfg *config.Config) *cobra.Command {
	sudoCmd := &cobra.Command{
		Use:    "sudo",
		Short:  "sudo commands for superusers",
		Hidden: true,
	}
	sudoCmd.AddCommand(whitelist.WhitelistCmd(cfg))
	sudoCmd.AddCommand(superuser.SuperuserCmd(cfg))
	sudoCmd.AddCommand(user.UserCmd(cfg))
	sudoCmd.AddCommand(gitCloneCmd(cfg))
	sudoCmd.AddCommand(lookupCmd(cfg))

	return sudoCmd
}
