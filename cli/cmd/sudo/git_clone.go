package sudo

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func gitCloneCmd(_ *cmdutil.Helper) *cobra.Command {
	return &cobra.Command{
		Use:        "git-clone <org> <project>",
		Args:       cobra.ExactArgs(2),
		Short:      "Create git clone token",
		Deprecated: "Command is deprecated. Use `rill sudo clone <org> <project>` instead.",
	}
}

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

			res, err := client.GetArtifactsURL(ctx, &adminv1.GetArtifactsURLRequest{
				Organization: args[0],
				Project:      args[1],
			})
			if err != nil {
				return err
			}

			fmt.Println("Clone command:")
			if res.UploadPath != "" {
				fmt.Printf("\tgsutil cp %s .\n\n", res.UploadPath)
				return nil
			}

			ep, err := transport.NewEndpoint(res.RepoUrl)
			if err != nil {
				return err
			}
			ep.User = res.Username
			ep.Password = res.Password
			cloneURL := ep.String()

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

	return cloneCmd
}
