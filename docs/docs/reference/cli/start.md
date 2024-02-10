---
note: GENERATED. DO NOT EDIT.
title: rill start
---
## rill start

Build project and start web app

```
rill start [<path>] [flags]
```

### Flags

```
      --no-open              Do not open browser
      --db string            Database DSN (default "main.db")
      --db-driver string     Database driver (default "duckdb")
      --port int             Port for HTTP (default 9009)
      --port-grpc int        Port for gRPC (internal) (default 49009)
      --readonly             Show only dashboards in UI
      --no-ui                Serve only the backend
      --verbose              Sets the log level to debug
      --debug                Collect additional debug info
      --reset                Clear and re-ingest source data
      --log-format string    Log format (options: "console", "json") (default "console")
      --environment string   Environment name (default "development")
  -v, --variable strings     Set project variables
```

### Global flags

```
  -h, --help          Print usage
      --interactive   Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](cli.md)	 - Rill CLI

