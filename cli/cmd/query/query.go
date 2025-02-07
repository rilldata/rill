package query

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
)

func QueryCmd(ch *cmdutil.Helper) *cobra.Command {
	var sql, connector, resolver, project, path string
	var local bool
	var limit int
	var properties, args map[string]string

	queryCmd := &cobra.Command{
		Use:   "query [<project>]",
		Short: "Query a resolver within a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project name
			if len(args) > 0 {
				project = args[0]
			}
			if !local && !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// TODO: Validate flag combinations

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return err
			}
			defer rt.Close()

			// Execute the query
			res, err := rt.RuntimeServiceClient.QueryResolver(cmd.Context(), &runtimev1.QueryResolverRequest{
				InstanceId: instanceID,
				Resolver:   resolver,
			})
			if err != nil {
				return err
			}

			// Print the data in the requested format (default: human)
			ch.PrintData(res.Data)

			return nil
		},
	}

	queryCmd.Flags().StringVar(&project, "project", "", "Project name")
	queryCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	queryCmd.Flags().StringVar(&sql, "sql", "", "A SELECT query to execute")
	queryCmd.Flags().StringVar(&connector, "connector", "", "Connector to execute against. Defaults to the OLAP.")
	queryCmd.Flags().StringVar(&resolver, "resolver", "", "Explicit resolver (cannot be combined with --sql)")
	queryCmd.Flags().StringToStringVar(&properties, "properties", nil, "Explicit resolver properties (only with --resolver)")
	queryCmd.Flags().StringToStringVar(&args, "args", nil, "Explicit resolver args (only with --resolver)")
	queryCmd.Flags().BoolVar(&local, "local", false, "Target localhost instead of Rill Cloud")
	queryCmd.Flags().IntVar(&limit, "limit", 100, "The maximum number of rows to print (default: 100)")

	return queryCmd
}
