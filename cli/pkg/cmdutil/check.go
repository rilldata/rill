package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
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
		localFlag := cmd.Flags().Lookup("local")
		if localFlag != nil && localFlag.Changed {
			return nil
		}

		// This will just check if token is present in the config
		if !ch.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'rill login'")
		}
		return nil
	}
}

// CheckOrganization checks if the user has an organization set.
func CheckOrganization(ch *Helper) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		localFlag := cmd.Flags().Lookup("local")
		if localFlag != nil && localFlag.Changed {
			return nil
		}

		if ch.Org != "" {
			return nil
		}

		return fmt.Errorf("no organization is set, pass `--org` or run `rill org switch`")
	}
}
