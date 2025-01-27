package user

import (
	"errors"
	"fmt"
	"net/mail"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <email>",
		Short: "Remove a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return err
			}

			email := args[0]
			if email == "" {
				return errors.New("email is required")
			}

			if _, err := mail.ParseAddress(email); err != nil {
				return fmt.Errorf("invalid email: %w", err)
			}

			orgs, err := client.ListOrganizationsByUser(ctx, &adminv1.ListOrganizationsByUserRequest{
				Email: email,
			})
			if err != nil {
				return fmt.Errorf("failed to list organizations for user %q: %w", email, err)
			}

			for _, org := range orgs.Organizations {
				_, err = client.DeleteUser(ctx, &adminv1.DeleteUserRequest{Email: email, Organization: org.Id})
				if err != nil {
					return fmt.Errorf("failed to remove user %q from organization %q: %w", email, org.Id, err)
				}
			}

			cmd.Printf("User %q removed from all organizations\n", email)

			return nil
		},
	}
	return cmd
}
