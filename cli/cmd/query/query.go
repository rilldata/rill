package query

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

var long = `Query data in a project.

You can query data by providing a SQL query and optional connector name.
As an advanced option, you can also query other resolvers such as metrics_sql.

Note that large results are automatically truncated (use --limit to override).
`

func QueryCmd(ch *cmdutil.Helper) *cobra.Command {
	var sql, connector, resolver, project, path, branch string
	var properties, args map[string]string
	var limit int
	var local bool

	queryCmd := &cobra.Command{
		Use:   "query [<project>]",
		Short: "Query data in a project",
		Long:  long,
		Example: `  # SQL query against a Rill Cloud project
  rill query my-project --sql "SELECT * FROM my-table"

  # SQL query against a local Rill project running with 'rill start'
  rill query --local --sql "SELECT * FROM my-table"`,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			// Validate the inputs
			if resolver == "" && sql == "" {
				return fmt.Errorf("must provide --sql or --resolver")
			}
			if resolver != "" && (sql != "" || connector != "") {
				return fmt.Errorf("cannot combine --resolver with --sql or --connector")
			}
			if sql != "" && len(properties) > 0 {
				return fmt.Errorf("cannot combine --sql with --properties")
			}

			// Rewrite --sql to resolver
			if sql != "" {
				resolver = "sql"
				properties = map[string]string{"sql": sql}
				if connector != "" {
					properties["connector"] = connector
				}
			}

			// Determine project name
			if len(cmdArgs) > 0 {
				project = cmdArgs[0]
			}
			if !local && project == "" {
				if !ch.Interactive {
					return fmt.Errorf("set --project to target a Rill Cloud project, or use --local to target a locally running Rill project")
				}
				// Check if a local Rill project is running; if so, target it automatically.
				if cmdutil.IsLocalRillRunning(cmd.Context()) {
					local = true
				} else {
					var err error
					project, err = ch.InferProjectName(cmd.Context(), path, "set --project to target a Rill Cloud project, or use --local to target a locally running Rill project")
					if err != nil {
						return err
					}
				}
			}

			// If targeting a local runtime, verify that rill start is running.
			if local && !cmdutil.IsLocalRillRunning(cmd.Context()) {
				return fmt.Errorf("could not connect to a local Rill project on http://localhost:9009 (run `rill start` to start the project, then retry this query)")
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, branch, local)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}
			defer rt.Close()

			// Execute the query
			res, err := rt.RuntimeServiceClient.QueryResolver(cmd.Context(), &runtimev1.QueryResolverRequest{
				InstanceId:         instanceID,
				Resolver:           resolver,
				ResolverProperties: buildStruct(properties),
				ResolverArgs:       buildStruct(args),
				Limit:              int32(limit),
			})
			if err != nil {
				return fmt.Errorf("failed to execute query: %w", err)
			}

			// Print rows with warning if the default limit was reached.
			ch.PrintQueryResponse(res)
			if !cmd.Flags().Changed("limit") && len(res.Data) == limit {
				ch.PrintfWarn("Warning: The result was truncated to %d rows (use --limit to override)\n", limit)
			}
			return nil
		},
	}

	// Project flags
	queryCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	queryCmd.Flags().StringVar(&project, "project", "", "Project name")
	queryCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	queryCmd.Flags().StringVar(&branch, "branch", "", "Target deployment by Git branch (default: primary deployment)")
	queryCmd.Flags().BoolVar(&local, "local", false, "Target local runtime instead of Rill Cloud")

	// Query flags
	queryCmd.Flags().StringVar(&sql, "sql", "", "A SELECT query to execute")
	queryCmd.Flags().StringVar(&connector, "connector", "", "Connector to execute against. Defaults to the OLAP connector.")
	queryCmd.Flags().StringVar(&resolver, "resolver", "", "Explicit resolver (cannot be combined with --sql)")
	queryCmd.Flags().StringToStringVar(&properties, "properties", nil, "Explicit resolver properties (only with --resolver)")
	queryCmd.Flags().StringToStringVar(&args, "args", nil, "Explicit resolver args (only with --resolver)")
	queryCmd.Flags().IntVar(&limit, "limit", 100, "The maximum number of rows to print")

	return queryCmd
}

func buildStruct(m map[string]string) *structpb.Struct {
	if m == nil {
		return nil
	}
	anyMap := make(map[string]any, len(m))
	for k, v := range m {
		anyMap[k] = v
	}
	pb, err := structpb.NewStruct(anyMap)
	if err != nil {
		// Acceptable to panic because there are no unknown types, so this should never happen.
		panic(fmt.Errorf("failed to build struct: %w", err))
	}
	return pb
}
