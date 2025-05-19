package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	ErrNotAuthenticated = fmt.Errorf("not authenticated, please run 'rill login'")
	ErrNoOrganization   = fmt.Errorf("no organization is set, pass `--org` or run `rill org switch`")
)

// PreRunCheck is called before a command is run.
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

// CheckAuth checks if the user is authenticated.
func CheckAuth(ch *Helper) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		// This will just check if token is present in the config
		if !ch.IsAuthenticated() {
			return fmt.Errorf("command '%s': %w", cmd.Name(), ErrNotAuthenticated)
		}
		return nil
	}
}

// CheckOrganization checks if the user has an organization set.
func CheckOrganization(ch *Helper) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		// If the command is run in local mode, skip the check.
		if cmd.Flags().Lookup("local").Changed {
			return nil
		}

		if ch.Org != "" {
			return nil
		}

		return ErrNoOrganization
	}
}
