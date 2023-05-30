package user

import (
	"net/url"
	"strings"

	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func OpenCmd(cfg *config.Config) *cobra.Command {
	openCmd := &cobra.Command{
		Use:   "open",
		Args:  cobra.NoArgs,
		Short: "Open",
		RunE: func(cmd *cobra.Command, args []string) error {
			authURL := cfg.AdminURL
			if strings.Contains(authURL, "http://localhost:9090") {
				authURL = "http://localhost:8080"
			}

			withTokenURI, err := url.JoinPath(authURL, "auth/with-token")
			if err != nil {
				return err
			}

			qry := map[string]string{"token": cfg.AdminTokenDefault}
			// What about local
			qry["redirect"] = "ui.rilldata.com"

			withTokenURL, err := urlutil.WithQuery(withTokenURI, qry)
			if err != nil {
				return err
			}

			_ = browser.Open(withTokenURL)

			return nil
		},
	}
	return openCmd
}
