# CONTRIBUTING

[![PkgGoDev](https://pkg.go.dev/badge/github.com/rilldata/rill)](https://pkg.go.dev/github.com/rilldata/rill)
[![Go Report Card](https://goreportcard.com/badge/github.com/rilldata/rill)](https://goreportcard.com/report/github.com/rilldata/rill)
[![codecov](https://codecov.io/gh/rilldata/rill/branch/main/graph/badge.svg?token=RQA182JGA5)](https://codecov.io/gh/rilldata/rill)

This file should serve as an entrypoint for learning about and contributing to Rill Developer.

## Development environment

If you're a Rill team member, you can run `rill devtool start` from the project root to start a full local development environment. If you select the cloud preset, you can fill it with seed data using `rill devtool seed cloud`. See `cli/cmd/devtool/README.md` for more details.

### Development dependencies

This is a full list of development dependencies:

- [Docker](https://www.docker.com)
- [Node.js 18](https://nodejs.org/en/) (we recommend installing it with [nvm](https://github.com/nvm-sh/nvm))
- [Go 1.22](https://go.dev) (on macOS, install with `brew install go`)
- [Buf](https://buf.build) (Protocol Buffers) (on macOS, install with `brew install bufbuild/buf/buf`)

### Editor setup

- Integrate `golangci-lint` ([instructions](https://golangci-lint.run/usage/integrations))

## Build the application

Running `make cli` will build a production-ready binary and output it to `./rill` (see `cli/README.md` for details).

For detailed instructions on how to run or test the application in development, see the `README.md` file in the individual components' directories (e.g. `web-local/README.md` for the local web app).

## Release a new version

To release a new version of Rill, simply add a Git tag with the new version number on `main` and push it to Github:

```bash
git tag -a v0.16.0 -m "v0.16.0 release"
git push origin v0.16.0
```

This will trigger the `cli-release.yml` Github Action, which will:

- Build binaries for macOS (arm64, amd64) and Linux (amd64)
- Upload the binaries to `https://cdn.rilldata.com/rill/$VERSION/$NAME`
- Upload the newest version of the install script (in `scripts/install.sh`) to `https://cdn.rilldata.com/install.sh`
- Create a Github release containing an auto-generated changelog and the new binaries
- Publish the new version to our brew tap `rilldata/tap/rill`

You can follow the progress of the release action from the ["Actions" tab](https://github.com/rilldata/rill/actions). It usually completes in about 10 minutes. See our internal [release run book](https://www.notion.so/rilldata/Release-Run-Book-20a4afb8f2f64d06814a0c89d51bfdcf) for more details.

## Technologies

Here's a high-level overview of the technologies we use for different parts of the project:

- Typescript and SvelteKit for all frontend code
- Go for the CLI and runtime (backend)
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

- `.github` contain CI/CD workflows.
- `admin` contains the backend control plane for the managed, multi-user version of Rill.
- `cli` contains the CLI and a server for the local frontend (used only in production).
- `docs` contains the user-facing documentation that we deploy to [docs.rilldata.com](https://docs.rilldata.com).
- `proto` contains protocol buffer definitions for all Rill components, which notably includes our API interfaces.
- `runtime` contains the engine (data plane) responsible for orchestrating and serving data.
- `scripts` contains various scripts and other resources used in development.
- `web-admin` contains the frontend control plane for the managed, multi-user version of Rill.
- `web-auth` contains the frontend code for `auth.rilldata.com` (managed with Auth0).
- `web-common` contains common functionality shared across the local and admin frontend applications.
- `web-local` contains the local Rill application, notably the data modeller.
