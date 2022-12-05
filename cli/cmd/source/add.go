package source

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// addCmd represents the add command, it requires min 1 args as source name
func AddCmd(ver string) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:   "add <source>",
		Short: "Add a source",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("add called with args:%s", strings.Join(args, " "))
			return nil
		},
	}
	return addCmd
}
