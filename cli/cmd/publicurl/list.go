package publicurl

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var pageSize uint32
	var pageToken string

	listCmd := &cobra.Command{
		Use:   "list [<project-name>]",
		Short: "List all public URLs",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				project = args[0]
			}
			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			res, err := client.ListMagicAuthTokens(cmd.Context(), &adminv1.ListMagicAuthTokensRequest{
				Org:       ch.Org,
				Project:   project,
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			ch.Printer.PrintMagicAuthTokens(res.Tokens)

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			cmd.Println()
			cmd.Println("NOTE: For security reasons, the actual URLs can't be displayed after creation.")

			return nil
		},
	}

	listCmd.Flags().SortFlags = false
	listCmd.Flags().StringVar(&project, "project", "", "Project name")
	listCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}
