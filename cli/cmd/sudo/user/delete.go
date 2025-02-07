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
	var force bool

	deleteCmd := &cobra.Command{
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

			ch.PrintfError("\nDeleting a user is a permanent action and cannot be undone.\n")
			ch.PrintfError("The user will be removed from all organizations and their data will be deleted.\n")
			confirm, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Are you sure you want to delete user %q?", email), "", false)
			if err != nil {
				return err
			}

			if confirm {
				_, err = client.DeleteUser(ctx, &adminv1.DeleteUserRequest{Email: email})
				if err != nil {
					return fmt.Errorf("failed to delete user %q: %w", email, err)
				}
				fmt.Printf("User %q deleted successfully\n", email)
			}

			return nil
		},
	}

	deleteCmd.Flags().BoolVar(&force, "force", false, "Allows superusers to bypass certain checks")
	_ = deleteCmd.Flags().MarkHidden("force")

	return deleteCmd
}
