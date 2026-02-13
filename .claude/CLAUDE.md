# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Testing

- **Build CLI**: `make cli` (builds Go binary and embeds frontend)
- **Build CLI only**: `make cli-only` (faster build without frontend)
- **Full development setup**: `npm run dev` (starts both runtime and web frontend)
- **Test Go code**: `go test ./...`
- **Test frontend**: `npm run test` (runs tests in web-common and web-local)
- **Local frontend test**: `npm run local-test` (Playwright tests in web-local)
- **Lint/Format**: `npm run lint` and `npm run format` in web-common

### Protocol Buffers

- **Generate proto clients**: `make proto.generate` (run after editing .proto files)

### Development Servers

- **Start local app**: `rill devtool local`
- **Cloud dev environment**: `rill devtool start cloud --except runtime` + `go run ./cli runtime start`
- **Web only**: `npm run dev-web`
- **Runtime only**: `npm run dev-runtime`

## Architecture Overview

### High-Level Structure

Rill is a data exploration platform with three main components:

- **Runtime**: Go-based data plane (proxy/orchestrator) connecting to databases like DuckDB, ClickHouse
- **CLI**: Go-based command-line interface for project management and deployment
- **Frontend**: Svelte-based web applications (web-local, web-admin, web-common)

### Key Directories

- `runtime/`: Core data infrastructure proxy with drivers, reconcilers, queries, and API server
- `cli/`: Command-line interface and project management tools
- `web-common/`: Shared frontend components and utilities
- `web-local/`: Local development web interface
- `web-admin/`: Cloud admin web interface
- `proto/`: gRPC/Protocol Buffer API definitions
- `admin/`: Cloud backend services (auth, billing, provisioning)

### Frontend Architecture

- **Framework**: Svelte/SvelteKit with TypeScript
- **State Management**: TanStack Query for server state
- **Styling**: Tailwind CSS
- **Build**: Vite with npm workspaces
- **Testing**: Vitest (unit) + Playwright (e2e)

### Backend Architecture

- **Language**: Go 1.24
- **APIs**: gRPC with gRPC-Gateway for REST mapping
- **Databases**: DuckDB (embedded), PostgreSQL (cloud), ClickHouse, others via drivers
- **Storage**: Local filesystem + object stores (S3, GCS, Azure)

## Code Conventions

### Go Code

- Standard Go conventions and project structure
- Use `context.Context` for request handling
- Driver interface pattern for database connections
- Reconciler pattern for resource state management

### Frontend Code

See [`.claude/rules/frontend.md`](.claude/rules/frontend.md) for frontend conventions.

### API Development

1. Define endpoint in `proto/rill/runtime/v1/api.proto`
2. Run `make proto.generate`
3. Implement handler in `runtime/server/`
4. For analytical queries, add implementation in `runtime/queries/`

## Important Notes

- **Proto regeneration**: Always run `make proto.generate` after editing .proto files
- **DuckDB development**: Use `-tags=duckdb_use_lib` flag when testing nightly builds
- **Monorepo**: Uses npm workspaces for frontend packages
- **Path aliases**: Configured in tsconfig.json for `@rilldata/web-*` imports
- **Git worktrees**: Create worktrees in `.claude/worktrees/` (e.g., `git worktree add .claude/worktrees/feature-branch feature-branch`)
