package sudo

import (
	"github.com/rilldata/rill/cli/cmd/sudo/project"
	"github.com/rilldata/rill/cli/cmd/sudo/quota"
	"github.com/rilldata/rill/cli/cmd/sudo/superuser"
	"github.com/rilldata/rill/cli/cmd/sudo/tags"
	"github.com/rilldata/rill/cli/cmd/sudo/user"
	"github.com/rilldata/rill/cli/cmd/sudo/whitelist"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func SudoCmd(ch *cmdutil.Helper) *cobra.Command {
	sudoCmd := &cobra.Command{
		Use:    "sudo",
		Short:  "sudo commands for superusers",
		Hidden: true,
	}
	sudoCmd.AddCommand(whitelist.WhitelistCmd(ch))
	sudoCmd.AddCommand(superuser.SuperuserCmd(ch))
	sudoCmd.AddCommand(user.UserCmd(ch))
	sudoCmd.AddCommand(quota.QuotaCmd(ch))
	sudoCmd.AddCommand(gitCloneCmd(ch))
	sudoCmd.AddCommand(lookupCmd(ch))
	sudoCmd.AddCommand(project.ProjectCmd(ch))
	sudoCmd.AddCommand(tags.TagsCmd(ch))

	return sudoCmd
}
