package virtualfiles

import (
	"fmt"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize int

	listCmd := &cobra.Command{
		Use:   "list <project>",
		Args:  cobra.ExactArgs(1),
		Short: "List all virtual files in a project's virtual repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			project := args[0]

			org := ch.Org
			if org == "" {
				return fmt.Errorf("org cannot be empty")
			}

			projResp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				Org:                  org,
				Project:              project,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}
			projectID := projResp.Project.Id

			if pageSize <= 0 {
				return fmt.Errorf("page-size must be greater than 0")
			}

			// Retrieve all virtual files with pagination
			var files []*adminv1.VirtualFile
			pageToken := ""
			ps := uint32(pageSize)
			if ps > 1000 {
				ps = 1000
			}
			for {
				res, err := client.PullVirtualRepo(ctx, &adminv1.PullVirtualRepoRequest{
					ProjectId:            projectID,
					Environment:          "prod",
					PageSize:             ps,
					PageToken:            pageToken,
					SuperuserForceAccess: true,
				})
				if err != nil {
					return fmt.Errorf("failed to pull virtual repo: %w", err)
				}

				if res.Files == nil {
					break
				}

				files = append(files, res.Files...)

				if res.NextPageToken == "" {
					break
				}
				pageToken = res.NextPageToken
			}

			if len(files) == 0 {
				ch.PrintfWarn("No virtual files found for project %q in org %q\n", project, org)
				return nil
			}

			sort.Slice(files, func(i, j int) bool {
				return files[i].Path < files[j].Path
			})

			ch.PrintfSuccess("Virtual files for project %q in org %q (%d total):\n", project, org, len(files))

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Updated On\tPath\tType\tSize (bytes)\tDeleted")
			for _, file := range files {
				size := len(file.Data)
				deleted := "No"
				if file.Deleted {
					deleted = "Yes"
				}
				updatedOn := file.UpdatedOn.AsTime().Local().Format(time.DateTime)
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", updatedOn, file.Path, size, deleted)
			}
			w.Flush()

			return nil
		},
	}

	listCmd.Flags().IntVar(&pageSize, "page-size", 100, "Number of files per page")
	listCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")

	return listCmd
}
