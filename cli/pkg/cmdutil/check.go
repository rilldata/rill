package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

type PreRunCheck func(cmd *cobra.Command, args []string) error

func CheckChain(chain ...PreRunCheck) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		for _, fn := range chain {
			err := fn(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func CheckAuth(ch *Helper) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		// This will just check if token is present in the config
		if ch.IsAuthenticated() {
			return nil
		}

		return fmt.Errorf("not authenticated, please run 'rill login'")
	}
}

func CheckOrganization(ch *Helper) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		if ch.Org != "" {
			return nil
		}

		return fmt.Errorf("no organization is set, pass `--org` or run `rill org switch`")
	}
}
