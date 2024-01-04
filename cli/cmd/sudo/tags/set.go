package tags

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var tags []string
	setCmd := &cobra.Command{
		Use:   "set <organization> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Set Tags for project in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.SudoUpdateTags(ctx, &adminv1.SudoUpdateTagsRequest{
				Organization: args[0],
				Project:      args[1],
				Tags:         tags,
			})
			if err != nil {
				return err
			}

			tags := res.Project.Tags
			fmt.Printf("Project: %s\n", res.Project.Name)
			fmt.Printf("Organization: %s\n", res.Project.OrgName)
			fmt.Printf("Tags: %v\n", tags)

			return nil
		},
	}
	setCmd.Flags().StringArrayVar(&tags, "tag", []string{}, "Tags to set on the project")

	return setCmd
}
