# `runtime`

The runtime a data infrastructure proxy and orchestrator – our data plane. It connects to data infrastructure and is responsible for dashboard queries, parsing code files, reconciling infra state, implementing connectors, enforcing (row-based) access policies, scheduling tasks, triggering reports, and much more.

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
- `testruntime` contains helper functions for initializing a test runtime with test data.

## Development

### Developing the local application

Run `rill devtool local`. You need to stop and restart it using ctrl+C when you make code changes.

### Developing for cloud

In one terminal, start a full cloud development environment except the runtime:
```bash
rill devtool start cloud --except runtime
```

In a separate terminal, start a runtime server:
```bash
go run ./cli runtime start
```

Optionally, deploy a seed project:
```bash
rill devtool seed cloud
```

### Running tests

You can run all tests using:
```bash
go test ./runtime/...
```

## Configuration

The runtime server is configured using environment variables parsed in `cli/cmd/runtime/start.go`.

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
