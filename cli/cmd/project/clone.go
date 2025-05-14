package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CloneCmd(ch *cmdutil.Helper) *cobra.Command {
	var path string

	cloneCmd := &cobra.Command{
		Use:   "clone [<project-name>]",
		Args:  cobra.ExactArgs(1),
		Short: "Clone Project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}
			name := args[0]

			proj, err := client.GetCloneCredentials(cmd.Context(), &adminv1.GetCloneCredentialsRequest{
				Organization: ch.Org,
				Project:      name,
			})
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					return err
				}
				if st.Code() == codes.InvalidArgument {
					return fmt.Errorf("project %q not found: %w", name, err)
				}
				if st.Code() == codes.PermissionDenied {
					return fmt.Errorf("you do not have permission to clone project %q", name)
				}
				return err
			}

			return gitutil.RunGitClone(cmd.Context(), path, proj.GitProdBranch, gitutil.GitRemoteCredentials{
				Remote:   proj.GitRepoUrl,
				Username: proj.GitUsername,
				Password: proj.GitPassword,
			})
		},
	}

	cloneCmd.Flags().SortFlags = false
	cloneCmd.Flags().StringVar(&path, "path", ".", "Project path to clone to")

	return cloneCmd
}
