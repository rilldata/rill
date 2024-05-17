package uninstall

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/installscript"
	"github.com/spf13/cobra"
)

func UninstallCmd(ch *cmdutil.Helper) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the Rill binary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return installscript.Uninstall(cmd.Context())
		},
	}
}
