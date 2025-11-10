package virtual_files

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/rilldata/rill/cli/pkg/adminenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var pageSize int

	listCmd := &cobra.Command{
		Use:   "list <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "List all virtual files in a project's virtual repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			org := args[0]
			project := args[1]
			environment, err := adminenv.Infer(ch.AdminURL())
			if err != nil {
				return err
			}

			if pageSize <= 0 {
				return fmt.Errorf("page-size must be greater than 0")
			}

			files, err := pullVirtualFiles(ctx, client, project, environment, uint32(pageSize))
			if err != nil {
				return err
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
				fileType := GetFileType(file.Path)
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", file.UpdatedOn, file.Path, fileType, size, deleted)
			}
			w.Flush()

			return nil
		},
	}

	listCmd.Flags().IntVar(&pageSize, "page-size", 100, "Number of files per page")

	return listCmd
}
