package cmdutil

import (
	"errors"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func CheckAuth(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cfg.IsAuthenticated() {
			return nil
		}

		return errors.New("not authenticated")
	}
}
