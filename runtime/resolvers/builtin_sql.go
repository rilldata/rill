package resolvers

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime"
)

func init() {
	runtime.RegisterResolverInitializer("builtin_sql", newBuiltinSQL)
	runtime.RegisterBuiltinAPI("sql", "builtin_sql", nil)
}

type builtinSQLArgs struct {
	Connector string `mapstructure:"connector"`
	SQL       string `mapstructure:"sql"`
	Priority  int    `mapstructure:"priority"`
}

// newBuiltinSQL is the resolver for the built-in /sql API.
// It executes a SQL query provided dynamically through the args.
// It errors if the user identified by the attributes is not an admin.
func newBuiltinSQL(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Only admins and non-users (i.e. local users and service accounts) can run arbitrary SQL queries.
	if len(opts.UserAttributes) > 0 {
		admin, ok := opts.UserAttributes["admin"].(bool)
		if !ok || !admin {
			return nil, errors.New("must be an admin to run arbitrary SQL queries")
		}
	}

	// Decode the args
	args := &builtinSQLArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	// Rewrite to the regular SQL resolver
	return newSQL(ctx, &runtime.ResolverOptions{
		Runtime:    opts.Runtime,
		InstanceID: opts.InstanceID,
		Properties: map[string]any{
			"connector": args.Connector,
			"sql":       args.SQL,
		},
		Args: map[string]any{
			"priority": args.Priority,
		},
		UserAttributes: opts.UserAttributes,
		ForExport:      opts.ForExport,
	})
}
