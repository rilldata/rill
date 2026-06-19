package sudo

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/gitutil"
	"github.com/spf13/cobra"
)

func cloneCmd(ch *cmdutil.Helper) *cobra.Command {
	cloneCmd := &cobra.Command{
		Use:   "clone <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Get clone instrunctions for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetCloneCredentials(ctx, &adminv1.GetCloneCredentialsRequest{
				Org:                  args[0],
				Project:              args[1],
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			fmt.Println("Clone command:")
			if res.ArchiveDownloadUrl != "" {
				fmt.Printf("\tcurl -o %s__%s.tar.gz '%s'\n\n", args[0], args[1], res.ArchiveDownloadUrl)
				return nil
			}

			config := &gitutil.Config{
				Remote:   res.GitRepoUrl,
				Username: res.GitUsername,
				Password: res.GitPassword,
			}
			cloneURL, err := config.FullyQualifiedRemote()
			if err != nil {
				return err
			}

			fmt.Printf("\tgit clone %s\n\n", cloneURL)
			fmt.Println("Full details:")
			fmt.Printf("\tRepo URL: %s\n", res.GitRepoUrl)
			fmt.Printf("\tUsername: %s\n", res.GitUsername)
			fmt.Printf("\tPassword: %s\n", res.GitPassword)
			fmt.Printf("\tSubpath: %s\n", res.GitSubpath)
			fmt.Printf("\tPrimary branch: %s\n", res.GitPrimaryBranch)
			fmt.Printf("\nNote the credentials are only valid for a limited duration of time.\n")

			return nil
		},
	}

	return cloneCmd
}
