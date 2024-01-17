package annotations

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AnnotationsCmd(ch *cmdutil.Helper) *cobra.Command {
	annotationsCmd := &cobra.Command{
		Use:   "annotations",
		Short: "Manage annotations for project in an organization",
	}

	annotationsCmd.AddCommand(GetCmd(ch))
	annotationsCmd.AddCommand(SetCmd(ch))

	return annotationsCmd
}
