package embed

import (
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func OpenCmd(ch *cmdutil.Helper) *cobra.Command {
	var branch string
	var ttlSeconds uint32
	var userID string
	var userEmail string
	var userAttributes string
	var externalUserID string
	var resourceType string
	var resource string
	var theme string
	var themeMode string
	var navigation bool
	var query map[string]string
	var noOpen bool

	openCmd := &cobra.Command{
		Use:   "open <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Open an embedded dashboard in the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			org := args[0]
			project := args[1]

			req := &adminv1.GetIFrameRequest{
				Org:                  org,
				Project:              project,
				Branch:               branch,
				TtlSeconds:           ttlSeconds,
				ExternalUserId:       externalUserID,
				Type:                 resourceType,
				Resource:             resource,
				Theme:                theme,
				ThemeMode:            themeMode,
				Navigation:           navigation,
				Query:                query,
				SuperuserForceAccess: true,
			}

			// Set user identity: only one of user_id, user_email, or user_attributes can be specified.
			n := 0
			if userID != "" {
				n++
				req.For = &adminv1.GetIFrameRequest_UserId{UserId: userID}
			}
			if userEmail != "" {
				n++
				req.For = &adminv1.GetIFrameRequest_UserEmail{UserEmail: userEmail}
			}
			if userAttributes != "" {
				n++
				var attrs map[string]any
				if err := json.Unmarshal([]byte(userAttributes), &attrs); err != nil {
					return fmt.Errorf("invalid --user-attributes JSON: %w", err)
				}
				s, err := structpb.NewStruct(attrs)
				if err != nil {
					return fmt.Errorf("failed to parse --user-attributes: %w", err)
				}
				req.For = &adminv1.GetIFrameRequest_Attributes{Attributes: s}
			}
			if n > 1 {
				return fmt.Errorf("only one of --user-id, --user-email, or --user-attributes can be specified")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetIFrame(cmd.Context(), req)
			if err != nil {
				return err
			}

			if noOpen || !ch.Interactive {
				ch.Printf("Open browser at: %s\n", res.IframeSrc)
			} else {
				ch.Printf("Opening browser at: %s\n", res.IframeSrc)
				_ = browser.Open(res.IframeSrc)
			}

			return nil
		},
	}

	openCmd.Flags().StringVar(&branch, "branch", "", "Branch to embed (defaults to the primary branch)")
	openCmd.Flags().Uint32Var(&ttlSeconds, "ttl-seconds", 0, "TTL for the access token in seconds")
	openCmd.Flags().StringVar(&userID, "user-id", "", "Rill user ID to assume")
	openCmd.Flags().StringVar(&userEmail, "user-email", "", "User email to assume")
	openCmd.Flags().StringVar(&userAttributes, "user-attributes", "", "User attributes as JSON (e.g. '{\"domain\":\"example.com\"}')")
	openCmd.Flags().StringVar(&externalUserID, "external-user-id", "", "External user ID for per-user state")
	openCmd.Flags().StringVar(&resourceType, "resource-type", "", "Type of resource to embed")
	openCmd.Flags().StringVar(&resource, "resource", "", "Name of the resource to embed")
	openCmd.Flags().StringVar(&theme, "theme", "", "Theme for the embedded resource")
	openCmd.Flags().StringVar(&themeMode, "theme-mode", "", "Theme mode")
	openCmd.Flags().BoolVar(&navigation, "navigation", false, "Enable navigation between resources")
	openCmd.Flags().StringToStringVar(&query, "query", nil, "Additional query parameters (key=value)")
	openCmd.Flags().BoolVar(&noOpen, "no-open", false, "Print the URL without opening the browser")

	return openCmd
}
