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
      --no-open             Do not open browser
      --port int            Port for HTTP (default 9009)
      --port-grpc int       Port for gRPC (internal) (default 49009)
      --readonly            Show only dashboards in UI
      --no-ui               Serve only the backend
      --verbose             Sets the log level to debug
      --debug               Collect additional debug info
      --reset               Clear and re-ingest source data
      --log-format string   Log format (options: "console", "json") (default "console")
      --tls-cert string     Path to TLS certificate
      --tls-key string      Path to TLS key file
  -e, --env strings         Environment name (default "dev")
  -v, --var strings         Set project variables
```

### Global flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

### SEE ALSO

* [rill](cli.md)	 - Rill CLI

