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
  -e, --env strings               Set environment variables
      --environment string        Environment name (default "dev")
      --reset                     Clear and re-ingest source data
      --no-open                   Do not open browser
      --verbose                   Sets the log level to debug
      --port int                  Port for HTTP (default 9009)
      --port-grpc int             Port for gRPC (internal) (default 49009)
      --no-ui                     Serve only the backend
      --debug                     Collect additional debug info
      --log-format string         Log format (options: "console", "json") (default "console")
      --tls-cert string           Path to TLS certificate
      --tls-key string            Path to TLS key file
      --allowed-origins strings   Override allowed origins for CORS
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

