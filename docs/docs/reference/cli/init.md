---
note: GENERATED. DO NOT EDIT.
title: rill init
---
## rill init

Add Rill project files from a template

### Synopsis

Initialize a new Rill project or add files to an existing project from a template.

The available templates are:
- empty-duckdb: Create a new empty Rill project with DuckDB.
- empty-clickhouse: Create a new empty Rill project with ClickHouse.
- cursor: Add Cursor rules to an existing Rill project.
- claude: Add Claude Code instructions to an existing Rill project.


```
rill init [<path>] [flags]
```

### Flags

```
      --force             Overwrite existing files when unpacking a template
      --template string   Project template to use (default: prompt to select)
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

