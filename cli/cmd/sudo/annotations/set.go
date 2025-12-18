package annotations

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var annotations map[string]string

	setCmd := &cobra.Command{
		Use:   "set <organization> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Set annotations for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(annotations) == 0 {
				ch.PrintfWarn("Setting an empty annotation list will remove all annotations from the project.\n")
				ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", false)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			_, err = client.SudoUpdateAnnotations(ctx, &adminv1.SudoUpdateAnnotationsRequest{
				Org:         args[0],
				Project:     args[1],
				Annotations: annotations,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	setCmd.Flags().StringToStringVar(&annotations, "annotation", nil, "Annotation(s) to set on project")

	return setCmd
}
