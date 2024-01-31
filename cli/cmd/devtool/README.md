# `cli/cmd/devtool`

## Examples

Start a cloud development environment (automatically refreshes `.env` and switches the CLI to `dev`):
```bash
rill devtool start cloud
```

Authenticate and deploy the `github.com/rilldata/rill-examples/rill-openrtb-prog-ads` project in your dev environment:
```bash
rill devtool seed cloud
```

Start a clean cloud development environment:
```bash
rill devtool start cloud --reset
```

Start a cloud development environment without the runtime:
```bash
rill devtool start cloud --except runtime
```

Start a cloud development environment with only the admin server and external dependencies (Postgres, etc.):
```bash
rill devtool start cloud --only admin,deps
```

Start a local development environment:
```bash
rill devtool start local
```

Manually switch between cloud environments:
```bash
rill devtool switch-env stage
```

Capture your current `.env` file and distribute it to other users of the devtool:
```bash
rill devtool dotenv upload cloud 
```

## Services started by the devtool

### Cloud

- UI: `http://localhost:3000`
- Admin HTTP: `http://localhost:8080`
- Admin gRPC: `http://localhost:9090`
- Admin debug: `http://localhost:6060`
- Runtime HTTP: `http://localhost:8081`
- Runtime gRPC: `http://localhost:9091`
- Runtime debug: `http://localhost:6061`
- Postgres: `http://localhost:5432`
- Redis: `http://localhost:6379`
- Zipkin UI: `http://localhost:9411`
- Prometheus UI: `http://localhost:9412`

### Local

- UI: `http://localhost:3000`
- Runtime HTTP: `http://localhost:9009`
- Runtime gRPC: `http://localhost:49009`
- Runtime debug: `http://localhost:6060`

## How it works

The devtool is simply a convenience wrapper that:

1. Starts the services that make up our "local" and "cloud" experiences in the correct order/configuration for a development environment
2. Uses Docker compose to start cloud dependencies
3. Uses the `gs://rill-devtool` GCS bucket to share `.env` files for local development
4. Manipulates the `~/.rill` configuration files to re-direct CLI commands to dev/staging backends
