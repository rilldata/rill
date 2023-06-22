---
title: rill start
---
## rill start

Build project and start web app

```
rill start [<path>] [flags]
```

### Flags

```
      --no-open             Do not open browser
      --db string           Database DSN (default "stage.db")
      --db-driver string    Database driver (default "duckdb")
      --port int            Port for HTTP (default 9009)
      --port-grpc int       Port for gRPC (default 9010)
      --readonly            Show only dashboards in UI
      --no-ui               Serve only the backend
      --verbose             Sets the log level to debug
      --strict              Exit if project has build errors
      --log-format string   Log format (options: "console", "json") (default "console")
  -e, --env strings         Set project variables
```

### Global flags

```
  -h, --help          Print usage
      --interactive   Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](cli.md)	 - Rill CLI

