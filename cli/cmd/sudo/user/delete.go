package user

import (
	"errors"
	"fmt"
	"net/mail"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <email>",
		Short: "Delete a user",
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

			_, err = client.GetUser(ctx, &adminv1.GetUserRequest{
				Email: email,
			})
			if err != nil {
				return fmt.Errorf("user %q not found: %w", email, err)
			}

			_, err = client.DeleteUser(ctx, &adminv1.DeleteUserRequest{Email: email})
			if err != nil {
				return fmt.Errorf("failed to delete user %q: %w", email, err)
			}

			fmt.Printf("User %q deleted successfully\n", email)

			return nil
		},
	}
	return cmd
}
