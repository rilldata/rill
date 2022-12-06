# CONTRIBUTING

This file should serve as an entrypoint for learning about and contributing to Rill Developer.

## Development environment

This is a full list of development dependencies:

- [Docker](https://www.docker.com)
- [Node.js 18](https://nodejs.org/en/) (we recommend installing it with [nvm](https://github.com/nvm-sh/nvm))
- [Go 1.19](https://go.dev) (on macOS, install with `brew install go`)
- [GraalVM](https://www.graalvm.org) and [Maven](https://maven.apache.org) for Java (we recommend installing both through [sdkman](https://sdkman.io))
- [Buf](https://buf.build) (Protocol Buffers) (on macOS, install with `brew install bufbuild/buf/buf`)

## Build the application

Running `make cli` will build a production-ready binary and output it to `./rill` (see `cli/README.md` for details).

For detailed instructions on how to run or test the application in development, see the `README.md` file in the individual components' directories (e.g. `web-local/README.md` for the local web app).

## Technologies

Here's a high-level overview of the technologies we use for different parts of the project:

- Typescript and SvelteKit for all frontend code
- Go for the CLI and runtime (backend)
- Java and Apache Calcite for the SQL engine, which gets compiled to a native library using GraalVM
- DuckDB for OLAP on small data
- Apache Druid for OLAP on big data
- Postgres for handling metadata in hosted deployments
- OpenAPI and/or gRPC for most APIs
- Docker for running Postgres and Druid in local development

## Monorepo

Rill uses a monorepo and you can expect to find all its code in this repository. This allows us to move faster as we can coordinate changes across multiple components in a single PR. It also gives people a single place to learn about the project and follow its development.

We want the codebase to be easy to understand and contribute to. To achieve that, every directory that contains code of non-trivial complexity should include a `README.md` file that provides details about the module, such as its purpose, how to run and test it, links to relevant tutorials or docs, etc. Only the root `README.md` file should be considered user-facing.

The project uses NPM for Node.js (specifically, NPM [workspaces](https://docs.npmjs.com/cli/v7/using-npm/workspaces)), Go modules for Go, and Maven for Java. They function well alongside each other and we have not yet found a need for a cross-language build tool.

## Project structure

Here's a guide to the top-level structure of the repository:

- `.github` and `.travis.yml` contain CI/CD workflows. We allow both, but the goal is to move fully to Github Actions.
- `admin` contains the backend control plane for a multi-user, hosted version of Rill (in progress, not launched yet).
- `cli` contains the CLI and a server for the local frontend (used only in production).
- `docs` contains the user-facing documentation that we deploy to [docs.rilldata.com](https://docs.rilldata.com).
- `proto` contains protocol buffer definitions for all Rill components, which notably includes our API interfaces.
- `runtime` is our data plane, responsible for querying and orchestrating data infra. It currently supports DuckDB and Druid.
- `sql` contains our SQL parser and transpiler. It's based on Apache Calcite.
- `web-admin` contains the frontend control plane for a multi-user, hosted version of Rill (in progress, not launched yet).
- `web-common` contains common functionality shared across the local and cloud frontends.
- `web-local` contains the local Rill Developer application, including the data modeller and current CLI.
