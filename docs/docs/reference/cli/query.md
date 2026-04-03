---
note: GENERATED. DO NOT EDIT.
title: rill query
---
## rill query

Query data in a project

### Synopsis

Query data in a project.

You can query data by providing a SQL query and optional connector name.
As an advanced option, you can also query other resolvers such as metrics_sql.

Note that large results are automatically truncated (use --limit to override).


```
rill query [<project>] [flags]
```

### Examples

```
  # SQL query against a Rill Cloud project
  rill query my-project --sql "SELECT * FROM my-table"

  # SQL query against a local Rill project running with 'rill start'
  rill query --local --sql "SELECT * FROM my-table"
```

### Flags

```
      --args stringToString         Explicit resolver args (only with --resolver) (default [])
      --branch string               Target deployment by Git branch (default: primary deployment)
      --connector string            Connector to execute against. Defaults to the OLAP connector.
      --limit int                   The maximum number of rows to print (default 100)
      --local                       Target local runtime instead of Rill Cloud
      --org string                  Organization Name
      --path string                 Project directory (default ".")
      --project string              Project name
      --properties stringToString   Explicit resolver properties (only with --resolver) (default [])
      --resolver string             Explicit resolver (cannot be combined with --sql)
      --sql string                  A SELECT query to execute
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](cli.md)	 - A CLI for Rill

