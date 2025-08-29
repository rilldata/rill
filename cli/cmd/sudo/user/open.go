package user

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func OpenCmd(ch *cmdutil.Helper) *cobra.Command {
	var ttlMinutes int
	openCmd := &cobra.Command{
		Use:   "open",
		Args:  cobra.NoArgs,
		Short: "Open browser as the assume user",
		RunE: func(cmd *cobra.Command, args []string) error {
			authURL := ch.AdminURL()
			assumeOpenURI, err := url.JoinPath(authURL, "auth", "assume-open")
			if err != nil {
				return err
			}
			representingUser, err := ch.DotRill.GetRepresentingUser()
			if err != nil {
				return err
			}
			if representingUser == "" {
				return errors.New("no representing user configured, please assume a user first")
			}
			qry := map[string]string{
				"representing_user": representingUser,
				"ttl_minutes":       strconv.Itoa(ttlMinutes),
			}
			assumeOpenURI, err = urlutil.WithQuery(assumeOpenURI, qry)
			if err != nil {
				return err
			}
			_ = browser.Open(assumeOpenURI)

			return nil
		},
	}
	openCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 60, "Minutes until the token should expire")
	return openCmd
}
