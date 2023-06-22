package sudo

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func gitCloneCmd(cfg *config.Config) *cobra.Command {
	gitCloneCmd := &cobra.Command{
		Use:   "git-clone <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Create git clone token",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.GetGitCredentials(ctx, &adminv1.GetGitCredentialsRequest{
				Organization: args[0],
				Project:      args[1],
			})
			if err != nil {
				return err
			}

			ep, err := transport.NewEndpoint(res.RepoUrl)
			if err != nil {
				return err
			}
			ep.User = res.Username
			ep.Password = res.Password
			cloneURL := ep.String()

			fmt.Println("Clone command:")
			fmt.Printf("\tgit clone %s\n\n", cloneURL)
			fmt.Println("Full details:")
			fmt.Printf("\tRepo URL: %s\n", res.RepoUrl)
			fmt.Printf("\tUsername: %s\n", res.Username)
			fmt.Printf("\tPassword: %s\n", res.Password)
			fmt.Printf("\tSubpath: %s\n", res.Subpath)
			fmt.Printf("\tProd branch: %s\n", res.ProdBranch)
			fmt.Printf("\nNote the credentials are only valid for a limited duration of time.\n")

			return nil
		},
	}

	return gitCloneCmd
}
