package shareurl

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var ttlMinutes int
	var filter string
	var fields []string

	createCmd := &cobra.Command{
		Use:   "create [<project-name>] <metrics view>",
		Short: "Create a shareable URL",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) == 2 {
				project = args[0]
			}
			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			metricsView := args[len(args)-1]

			var filterExpr *runtimev1.Expression
			if filter != "" {
				filterExpr = &runtimev1.Expression{}
				err := protojson.Unmarshal([]byte(filter), filterExpr)
				if err != nil {
					return fmt.Errorf("failed to parse filter expression: %w", err)
				}
			}

			res, err := client.IssueMagicAuthToken(cmd.Context(), &adminv1.IssueMagicAuthTokenRequest{
				Organization:      ch.Org,
				Project:           project,
				TtlMinutes:        int64(ttlMinutes),
				MetricsView:       metricsView,
				MetricsViewFilter: filterExpr,
				MetricsViewFields: fields,
			})
			if err != nil {
				return err
			}

			ch.Printf("%s\n", res.Url)
			return nil
		},
	}

	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&project, "project", "", "Project name")
	createCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	createCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 0, "Duration until the token expires (use 0 for no expiry)")
	createCmd.Flags().StringVar(&filter, "filter", "", "Limit access to the provided filter (json)")
	createCmd.Flags().StringSliceVar(&fields, "fields", nil, "Limit access to the provided fields")

	return createCmd
}
