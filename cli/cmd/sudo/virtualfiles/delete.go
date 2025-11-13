package virtualfiles

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete <project> <path>",
		Args:  cobra.ExactArgs(2),
		Short: "Delete a specific virtual file",
		Long: `Delete a specific virtual file by marking it as deleted.

Virtual files are stored in the virtual repository and represent runtime resources
like alerts, reports, and services. Deleting a virtual file will mark it as deleted
in the database, which will cause the runtime to remove the corresponding resource.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			project := args[0]
			path := args[1]
			org := ch.Org

			if org == "" || project == "" || path == "" {
				return fmt.Errorf("org, project, and path cannot be empty")
			}

			fileType, name := GetFileTypeAndName(path)
			if fileType == FileTypeUnknown || name == "" {
				return fmt.Errorf("unsupported file type for deletion at path %q", path)
			}

			if !force {
				ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Delete %s %q in project %q (org %q)?", fileType, name, project, org), "", false)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			switch fileType {
			case FileTypeAlert:
				_, err = client.DeleteAlert(ctx, &adminv1.DeleteAlertRequest{
					Org:     org,
					Project: project,
					Name:    name,
				})
				if err != nil {
					return fmt.Errorf("failed to delete alert: %w", err)
				}
				ch.PrintfSuccess("Successfully deleted alert %q\n", name)

			case FileTypeReport:
				_, err = client.DeleteReport(ctx, &adminv1.DeleteReportRequest{
					Org:     org,
					Project: project,
					Name:    name,
				})
				if err != nil {
					return fmt.Errorf("failed to delete report: %w", err)
				}
				ch.PrintfSuccess("Successfully deleted report %q\n", name)

			case FileTypeService:
				_, err = client.DeleteService(ctx, &adminv1.DeleteServiceRequest{
					Name: name,
					Org:  org,
				})
				if err != nil {
					return fmt.Errorf("failed to delete service: %w", err)
				}
				ch.PrintfSuccess("Successfully deleted service %q\n", name)

			default:
				return fmt.Errorf("unsupported file type: %s", fileType)
			}

			return nil
		},
	}

	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")

	return deleteCmd
}
