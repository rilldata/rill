---
note: GENERATED. DO NOT EDIT.
title: rill init
---
## rill init

Initialize a new Rill project

### Synopsis

Initialize a new Rill project. Use flags to customize the project or run interactively to be prompted for each option.

Available example projects:
  - rill-cost-monitoring (duckdb)
  - rill-github-analytics (duckdb)
  - rill-openrtb-prog-ads (duckdb)


```
rill init [<path>] [flags]
```

### Examples

```
  # Interactive initialization (prompts for all options)
  rill init

  # Create an empty DuckDB project with Claude agent instructions
  rill init my-project --olap duckdb --agent claude

  # Add Claude agent instructions to an existing Rill project
  rill init ./existing-project --agent claude
```

### Flags

```
      --agent string     Agent instructions (options: claude, cursor, agentsmd, all, none) (default "claude")
      --example string   Example project name (default: empty project)
      --olap string      OLAP engine (options: duckdb, clickhouse) (default "duckdb")
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

