package source

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// dropCmd represents the drop command, it requires min 1 args as source path
func DropCmd() *cobra.Command {
	var dropCmd = &cobra.Command{
		Use:   "drop <source>",
		Short: "Drop a source",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("drop called with args:%s", strings.Join(args, " "))
			return nil
		},
	}

	return dropCmd
}
