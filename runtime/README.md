# `runtime`

The runtime a data infrastructure proxy and orchestrator – our data plane. It connects to data infrastructure and is or will be responsible for transpiling queries, parsing code artifacts, reconciling infra state, implementing connectors, enforcing (row-based) access policies, scheduling tasks, triggering alerts, and much more.

It's designed as a modular component that can be embedded in local applications (as it is into Rill Developer) or deployed stand-alone in a cloud environment.

## Code structure

The base directory contains a `Runtime` type that represents the lifecycle of the runtime. It ties together the sub-directories:

- `client` contains a Go client library for connecting to a runtime server.
- `compilers` contains logic for parsing Rill projects (incomplete, currently mostly implemented in `services` instead).
- `connectors` contains connector implementations.
- `drivers` contains interfaces and drivers for external data infrastructure that the runtime interfaces with (like DuckDB and Druid).
- `pkg` contains utility libraries.
- `queries` contains pre-defined analytical queries that the runtime can serve (used for profiling and dashboards).
- `server` contains a server that implements the runtime's APIs.
- `services` contains logic for building and reconciling Rill projects.
- `testruntime` contains helper functions for initializing a test runtime with test data.

## How to test and run

You can start a stand-alone runtime server using the CLI:
```bash
go run ./cli runtime start
```

Ping the server to verify it's running:
```bash
go run ./cli runtime ping --base-url http://localhost:9090
```

You can run all tests using:
```bash
go test ./runtime/...
```

See details in `web-local/README.md` for info on how to run in development together with the local frontend.

## Configuration

The runtime server is configured using environment variables parsed in `cli/cmd/runtime/start.go`. All environment variables have reasonable defaults suitable for local development. The current defaults are:

```bash
RILL_RUNTIME_HTTP_PORT="8080"
RILL_RUNTIME_GRPC_PORT="9090"
RILL_RUNTIME_LOG_LEVEL="info"
RILL_RUNTIME_METRICS_EXPORTER="prometheus"
RILL_RUNTIME_TRACES_EXPORTER=""
RILL_RUNTIME_ALLOWED_ORIGINS="*"
RILL_RUNTIME_METASTORE_DRIVER="sqlite"
RILL_RUNTIME_METASTORE_URL=":memory:"
RILL_RUNTIME_AUTH_ENABLE="false"
RILL_RUNTIME_AUTH_ISSUER_URL=""
RILL_RUNTIME_AUTH_AUDIENCE_URL=""
RILL_RUNTIME_CONNECTION_CACHE_SIZE="100"
RILL_RUNTIME_QUERY_CACHE_SIZE="10000"
RILL_RUNTIME_ALLOW_HOST_ACCESS="false"
RILL_RUNTIME_SAFE_SOURCE_REFRESH="true"
RILL_RUNTIME_GITHUB_APP_ID=""
RILL_RUNTIME_GITHUB_APP_PRIVATE_KEY=""
```

## Adding a new endpoint

We define our APIs using gRPC and use [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to map the RPCs to a RESTful API. See `proto/README.md` for details.

To add a new endpoint:
1. Describe the endpoint in `proto/rill/runtime/v1/api.proto`
2. Re-generate gRPC and OpenAPI interfaces by running `make proto.generate`
3. Copy the new handler signature from the `RuntimeServiceServer` interface in `proto/gen/rill/runtime/v1/api_grpc_pb.go`
4. Paste the handler signature and implement it in a relevant file in `runtime/server/`

## Adding a new analytical query endpoint

1. Add a new endpoint for the query by following the steps in the section above ("Adding a new endpoint")
2. Implement the query in `runtime/queries` by following the instructions in `runtime/queries/README.md`

## Example: Creating an instance and rehydrating from code artifacts

```bash
# Start runtime
go run ./runtime/cmd/main.go

# Create instance
curl --request POST --url http://localhost:8080/v1/instances --header 'Content-Type: application/json' \
  --data '{
    "instance_id": "default",
    "olap_driver": "duckdb",
    "olap_dsn": "test.db",
    "repo_driver": "file",
    "repo_dsn": "./examples/ad_bids",
    "embed_catalog": true
}'

# Apply code artifacts
curl --request POST --url http://localhost:8080/v1/instances/default/reconcile --header 'Content-Type: application/json'

# Query data
curl --request POST --url http://localhost:8080/v1/instances/default/query --header 'Content-Type: application/json' \
  --data '{ "sql": "select * from ad_bids limit 10" }'

# Query explore API
curl --request POST --url http://localhost:8080/v1/instances/default/queries/metrics-views/ad_bids_metrics/toplist/domain --header 'Content-Type: application/json' \
  --data '{
    "measure_names": ["measure_0"],
    "limit": 10,
    "sort": [{ "name": "measure_0", "ascending": false }]
}'

# Query profiling API
curl --request GET --url http://localhost:8080/v1/instances/default/null-count/ad_bids/publisher

# Get catalog info
curl --request GET --url http://localhost:8080/v1/instances/default/catalog

# Refresh source named "ad_bids_source"
curl --request POST --url http://localhost:8080/v1/instances/default/catalog/ad_bids_source/refresh

# Get available connectors
curl --request GET   --url http://localhost:8080/v1/connectors/meta

# List files in project
curl --request GET --url http://localhost:8080/v1/instances/default/files

# Fetch file in project
curl --request GET --url http://localhost:8080/v1/instances/default/files/-/models/ad_bids.sql

# Update file in project
curl --request POST --url http://localhost:8080/v1/instances/default/files/-/models/ad_bids.sql --header 'Content-Type: application/json' \
  --data '{ "blob": "select id, timestamp, publisher, domain, bid_price from ad_bids_source" }'
```

## Using a DuckDB nightly build

The following steps apply for macOS, but a similar approach should work for Linux.

1. Download the latest DuckDB nightly from Github from the "Artifacts" section on the newest workflow run [here](https://github.com/duckdb/duckdb/actions?query=branch%3Amaster+event%3Arepository_dispatch+workflow%3AOSX))
2. Unzip the downloaded archive and copy the `libduckdb.dylib` file in the `libduckdb-osx-universal` folder to `/usr/local/lib`
  - You must use the command-line to copy the file. If you touch it using the Finder, macOS will quarantine it. To remove a quarantine, run: `xattr -d com.apple.quarantine libduckdb.dylib`.
3. DuckDB usually does not support older file formats, so delete the `stage.db` and `stage.db.wal` files in your `dev-project`
4. Add the flag `-tags=duckdb_use_lib` when running `go run` or `go build` to use the nightly build of DuckDB
  - If testing the local frontend, you need to temporarily set it in the `dev-runtime` script in `package.json`
  - For details, see [Linking DuckDB](https://github.com/marcboeker/go-duckdb#linking-duckdb)

Note: DuckDB often makes breaking changes to its APIs, so you may encounter other errors when using a dev version of DuckDB.
