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
	var noOpen bool
	var ttlMinutes int
	openCmd := &cobra.Command{
		Use:   "open [<email>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Open browser as an assumed user",
		RunE: func(cmd *cobra.Command, args []string) error {
			authURL := ch.AdminURL()
			assumeOpenURI, err := url.JoinPath(authURL, "auth", "assume-open")
			if err != nil {
				return err
			}

			var email string
			if len(args) == 1 {
				email = args[0]
			}
			if email == "" {
				email, err = ch.DotRill.GetRepresentingUser()
				if err != nil {
					return err
				}
			}
			if email == "" {
				return errors.New("no user specified; you must specify a user's email or separately assume a user with `rill sudo user assume`")
			}

			qry := map[string]string{
				"representing_user": email,
				"ttl_minutes":       strconv.Itoa(ttlMinutes),
			}
			assumeOpenURI, err = urlutil.WithQuery(assumeOpenURI, qry)
			if err != nil {
				return err
			}

			if !noOpen {
				ch.Printf("Opening browser at: %s\n", assumeOpenURI)
				_ = browser.Open(assumeOpenURI)
			} else {
				ch.Printf("Open browser at: %s\n", assumeOpenURI)
			}

			return nil
		},
	}
	openCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the browser automatically")
	openCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 60, "Minutes until the token should expire")
	return openCmd
}
