## What is Rill

Rill is a business intelligence platform built around the following principles:

- Code-first: configure projects using versioned and reproducible source code in the form of YAML and SQL files.
- Full stack: go from raw data sources to user-friendly dashboards powered by clean data with a single tool.
- Declarative: describe your business logic and Rill automatically runs the infrastructure, migrations and services necessary to make it real.
- OLAP databases: you can easily provision a fast analytical database and load data into it to build dashboards that stay interactive at scale.

## Architecture

Users define projects as YAML and SQL files that describe _resources_ — connectors, models, metrics views, dashboards, and more — organized in a DAG. The runtime **parses** project files into resources and **reconciles** each resource to its desired state (e.g., materializing a model into DuckDB, validating a connector). On the frontend, metrics views power two dashboard types: **explore dashboards** (drill-down, slice-and-dice) and **canvas dashboards** (free-form charts and tables). The platform also supports alerts, scheduled reports, custom APIs, and a built-in AI assistant.

Two deployment modes share the same codebase:

- **Rill Developer** — local application for data engineers. A single Go binary that embeds the CLI, runtime, and `web-local` frontend. Code-first, version-controlled workflow.
- **Rill Cloud** — hosted platform for teams. Runs the `admin` service, runtime(s), and `web-admin` frontend as separate services. Adds auth, billing, multi-tenancy, and collaboration.

### Key Directories

- `runtime/` — data plane: orchestration, queries, connectors, access policies, reconcilers
- `admin/` — cloud control plane: auth, billing, provisioning, project management
- `cli/` — CLI and local application server
- `web-common/` — shared frontend library consumed by both `web-local` and `web-admin`
- `web-local/` — local frontend (Rill Developer)
- `web-admin/` — cloud frontend (Rill Cloud)
- `proto/` — gRPC/protobuf API definitions (source of truth for all APIs)

## Development

### Common Commands

- **Build CLI**: `make cli` (Go binary + embedded frontend)
- **Build CLI only**: `make cli-only` (skip frontend, faster)
- **Local dev**: `rill devtool start local`
- **Cloud dev**: `rill devtool start cloud`
- **Test Go**: `go test ./...`
- **Test frontend (unit)**: `npm run test -w web-common` (fast, use for tight feedback loops)
- **Test frontend (e2e)**: `npm run test -w web-local` or `npm run test -w web-admin` (Playwright, slow)
- **Lint/format frontend**: `npm run quality`

### Adding or Changing APIs

APIs are defined in `.proto` files and mapped to REST via gRPC-Gateway. See `proto/README.md` for conventions.

1. Define endpoint in the relevant `.proto` file under `proto/rill/`
2. Run `make proto.generate`
3. Implement handler in `runtime/server/` (or `admin/server/`)

See `runtime/README.md` for details and analytical query patterns.

Frontend API clients are auto-generated from proto definitions using **Orval**. Do not hand-edit files under `web-common/src/runtime-client/` — regenerate them instead.

## Code Conventions

### Go

Follow the conventions in `CONTRIBUTING.md`. Key points:

- Use standard library `errors` (not `github.com/pkg/errors`)
- `golangci-lint` enforces style — integrate it in your editor
- Non-trivial directories should have a `README.md`
- Cloud deployments require backwards compatibility (see "Services" in `CONTRIBUTING.md`)

For runtime-specific patterns (drivers, reconcilers, queries), see `runtime/README.md`.

### Frontend

**Tech stack**: Svelte 4 (migrating to Svelte 5), TypeScript, TanStack Query, Tailwind CSS, Orval (API client generation)

Frontend conventions are being formalized in `.claude/rules/frontend.md` (coming soon).

## Tips

- **Monorepo**: Uses npm workspaces (frontend) and Go modules (backend)
- **Path aliases**: `@rilldata/web-*` imports configured in tsconfig.json
- **Embedded dashboards**: Explore and Canvas dashboards can be embedded in customer apps via iframe. When changing dashboard components, consider whether the change also affects the embed surface.
