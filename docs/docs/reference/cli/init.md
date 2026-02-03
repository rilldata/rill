---
note: GENERATED. DO NOT EDIT.
title: rill init
---
## rill init

Add Rill project files from a template

### Synopsis

Initialize a new Rill project or add files to an existing project from a template.

The available templates are:
- duckdb: Creates an empty Rill project configured to use DuckDB as the OLAP database.
- clickhouse: Creates an empty Rill project configured to use ClickHouse as the OLAP database.
- cursor: Adds Cursor rules in .cursor to an existing Rill project.
- claude: Adds Claude Code instruction in .claude to an existing Rill project.


```
rill init [<path>] [flags]
```

### Flags

```
      --force             Overwrite existing files when unpacking a template
      --template string   Project template to use (options: duckdb, clickhouse, cursor) (default "duckdb")
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

