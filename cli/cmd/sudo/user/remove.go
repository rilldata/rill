package user

import (
	"errors"
	"net/mail"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove email",
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

			if !valid(email) {
				return errors.New("invalid email format")
			}

			_, err = client.DeleteUser(ctx, &adminv1.DeleteUserRequest{
				Email: email,
			})
			if err != nil {
				return err
			}

			cmd.Printf("User %q has been removed\n", email)

			return nil
		},
	}
	return cmd
}
