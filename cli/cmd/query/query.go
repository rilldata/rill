package query

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

// Command Long Description
var long = `Query a resolver within a project.

You can query a resolver by providing a SQL query, a resolver name, or a connector name.

Example Usage:

Query a resolver by providing a SQL query:
rill query my-project --sql "SELECT * FROM my-table"
rill query --sql "SELECT * FROM my-table" --limit 10
`

func QueryCmd(ch *cmdutil.Helper) *cobra.Command {
	var sql, connector, resolver, project, path string
	var local bool
	var limit int
	var properties, args map[string]string

	queryCmd := &cobra.Command{
		Use:   "query [<project>]",
		Short: "Query a resolver within a project",
		Long:  long,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			// Validate all inputs
			if err := validateQueryFlags(resolver, sql, properties, args); err != nil {
				return err
			}
			// If the limit is negative, use the default limit and print a warning
			if limit < 0 {
				limit = 100
				fmt.Printf("WARNING: limit is negative, using default limit of 100\n")
			}

			// If the resolver is not provided, use the sql
			if resolver == "" {
				resolver = "sql"
			}

			// Initialize if nil
			if properties == nil {
				properties = make(map[string]string)
			}
			if args == nil {
				args = make(map[string]string)
			}

			// Determine project name
			if len(cmdArgs) > 0 {
				project = cmdArgs[0]
			}
			if !local && !cmd.Flags().Changed("project") && len(cmdArgs) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Set properties
			if sql != "" {
				properties["sql"] = sql
			}
			if connector != "" {
				properties["connector"] = connector
			}
			if limit > 0 {
				args["limit"] = fmt.Sprintf("%d", limit)
			}

			// Build the properties and args
			resolverProperties, resolverArgs, err := buildStructs(properties, args)
			if err != nil {
				return err
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return err
			}
			defer rt.Close()

			// Execute the query
			res, err := rt.RuntimeServiceClient.QueryResolver(cmd.Context(), &runtimev1.QueryResolverRequest{
				InstanceId:         instanceID,         // The instance ID to query
				Resolver:           resolver,           // This is the type of resolver to use (e.g. sql, metrics_view, etc.)
				ResolverProperties: resolverProperties, // These are resolver-specific properties
				ResolverArgs:       resolverArgs,       // These are resolver-specific arguments
			})
			if err != nil {
				return err
			}

			// Print the data in the requested format (default: human)
			ch.PrintQueryResponse(res)

			return nil
		},
	}

	queryCmd.Flags().StringVar(&project, "project", "", "Project name")
	queryCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	queryCmd.Flags().BoolVar(&local, "local", false, "Target local runtime instead of Rill Cloud")

	// Query flags
	queryCmd.Flags().StringVar(&sql, "sql", "", "A SELECT query to execute")
	queryCmd.Flags().StringVar(&connector, "connector", "", "Connector to execute against. Defaults to the OLAP connector.")
	queryCmd.Flags().StringVar(&resolver, "resolver", "", "Explicit resolver (cannot be combined with --sql)")
	queryCmd.Flags().StringToStringVar(&properties, "properties", nil, "Explicit resolver properties (only with --resolver)")
	queryCmd.Flags().StringToStringVar(&args, "args", nil, "Explicit resolver args (only with --resolver)")
	queryCmd.Flags().IntVar(&limit, "limit", 100, "The maximum number of rows to print (default: 100)")

	return queryCmd
}

func validateQueryFlags(resolver, sql string, properties, args map[string]string) error {
	// If the users provides a resolver via a flag, they cannot provide a sql
	if resolver != "" && sql != "" {
		return fmt.Errorf("cannot combine --resolver and --sql")
	}

	// If the user provides args or properties, they must provide a resolver
	if (len(args) > 0 || len(properties) > 0) && resolver == "" {
		return fmt.Errorf("must provide --resolver when using --args or --properties")
	}

	return nil
}

func buildStruct(m map[string]string) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}

	anyMap := make(map[string]interface{}, len(m))
	for k, v := range m {
		anyMap[k] = v
	}
	return structpb.NewStruct(anyMap)
}

// returns both the properties and args as structs
func buildStructs(properties, args map[string]string) (*structpb.Struct, *structpb.Struct, error) {
	propertiesStruct, err := buildStruct(properties)
	if err != nil {
		return nil, nil, err
	}

	argsStruct, err := buildStruct(args)
	if err != nil {
		return nil, nil, err
	}

	return propertiesStruct, argsStruct, nil
}
