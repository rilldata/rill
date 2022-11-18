# runtime

The runtime is our data plane. It connects to data infrastructure and is will be responsible for transpiling queries, applying migrations, implementing connectors, enforcing row-based access policies, scheduling tasks, triggering alerts, and much more.

It's designed as a stand-alone component that can be embedded in local applications (as it is into Rill Developer) or deployed in a cloud environment.

## Code structure

- `api` describes the runtime's API using Protocol Buffers (see `runtime.proto`) and generates gRPC and OpenAPI interfaces for it.
- `cmd` contains a `main.go` file that starts the runtime as a standalone server.
- `connectors` contains connector implementations.
- `drivers` contains interfaces and drivers for all the data infrastructure (and other persistant stores) we support.
- `pkg` contains utility libraries.
- `server` contains a server that implements the APIs described in `api`.
- `sql` contains bindings for the SQL native library (see the `sql` folder at the repo root for details).

## How to test and run

The runtime relies on the SQL native library being present in `runtime/sql/deps`. We don't check that into the repo, so you must manually download it by running:
```bash
go generate ./runtime/sql
```

Now, you can run and test the runtime as any other Go application. Start the server using:
```bash
go run ./runtime/cmd
```
Or run all tests using:
```bash
go test ./...
```

## Configuration

The runtime server is configured using environment variables parsed in `runtime/cmd/main.go`. All environment variables have reasonable defaults suitable for local development. The current defaults are:

```bash
RILL_RUNTIME_ENV="development"
RILL_RUNTIME_HTTP_PORT="8080"
RILL_RUNTIME_GRPC_PORT="9090"
RILL_RUNTIME_LOG_LEVEL="info"
RILL_RUNTIME_DATABASE_DRIVER="sqlite"
RILL_RUNTIME_DATABASE_URL=":memory:"
```

## Adding a new endpoint

To add a new endpoint:
1. Describe the endpoint in `runtime/api/runtime.proto`
2. Re-generate gRPC and OpenAPI interfaces by running `go generate ./runtime/api`
3. Copy the new handler signature from the `RuntimeServiceServer` interface in `runtime/api/runtime_grpc_pb.go`
4. Paste the handler signature and implement it in a file in `./runtime/server`

## Example: Creating an instance and ingesting a source

```bash
# Start runtime
go run ./runtime/cmd/main.go

# Create instance (copy the resulting instance ID into the following queries)
curl --request POST --url http://localhost:8080/v1/instances --header 'Content-Type: application/json' --data '{ "driver": "duckdb", "dsn": "test.db?access_mode=read_write", "exposed": true, "embed_catalog": true, "instance_id": "default" }'

# Create table
curl --request POST  --url http://localhost:8080/v1/instances/default/query  --header 'Content-Type: application/json'  --data '{"sql": "create table foo(x int)"}'

# Insert data into table
curl --request POST  --url http://localhost:8080/v1/instances/default/query  --header 'Content-Type: application/json'  --data '{"sql": "insert into foo(x) values (10,), (20,), (30,)"}'

# Query data
curl --request POST  --url http://localhost:8080/v1/instances/default/query  --header 'Content-Type: application/json'  --data '{"sql": "select * from foo"}'

# Get available connectors
curl --request GET   --url http://localhost:8080/v1/connectors/meta

# Create a source
curl --request POST  --url http://localhost:8080/v1/instances/default/migrate/single  --header 'Content-Type: application/json'  --data "{\"sql\": \"create source bar with connector = 'file', path = './web-local/test/data/AdBids.csv' \"}"

# Select from source
curl --request POST  --url http://localhost:8080/v1/instances/default/query  --header 'Content-Type: application/json'  --data '{"sql": "select * from bar limit 100"}'

# Get info about all sources in catalog
curl --request GET   --url http://localhost:8080/v1/instances/default/catalog

# Get info about source named "bar" in catalog
curl --request GET   --url http://localhost:8080/v1/instances/default/catalog/bar

# Refresh source named "bar"
curl --request POST --url http://localhost:8080/v1/instances/default/catalog/bar/refresh

# Delete source named "bar"
curl --request POST  --url http://localhost:8080/v1/instances/default/migrate/single/delete  --header 'Content-Type: application/json'  --data '{ "name": "bar"}'
```
