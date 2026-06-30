# Rill Development

## Cursor Cloud specific instructions

### Overview

Rill is a monorepo containing a Go backend (runtime + admin + CLI) and TypeScript/Svelte frontends (web-local, web-admin, web-common). The primary development mode for local work is **Rill Developer** which runs a Go runtime backend (port 9009) alongside the `web-local` SvelteKit frontend (port 3001).

### Quick reference

Commands are documented in `.claude/CLAUDE.md` under "Common Commands". Key ones:

| Task | Command |
|------|---------|
| Frontend lint + type check | `npm run quality` |
| Frontend unit tests (web-common) | `npm run test -w web-common` |
| Frontend unit tests (web-admin) | Run from `web-admin/`: `npx vitest run src/path/to/spec.ts` |
| Go tests | `go test ./...` (or `-short` for faster runs) |
| Go lint | `golangci-lint run ./path/to/package/` |
| Start local dev (runtime + frontend) | `npm run dev` from root (or start separately below) |
| Start runtime only | `go run cli/main.go start dev-project --no-ui --debug --allowed-origins http://localhost:3001` |
| Start frontend only | `npm run dev -w web-local -- --port 3001` |

### Startup caveats

- The runtime creates a `dev-project/` directory at the repo root on first run. This is gitignored.
- Go module download on first `go run` or `go test` takes significant time (~60s) in fresh environments due to the large dependency tree.
- `npm run quality` uses `npm ci` internally; if you already ran `npm install`, dependencies are cached and it proceeds quickly.
- The `web-common` vitest suite uses `forks` pool by default. In resource-constrained environments some worker processes may exit unexpectedly (showing "Worker exited unexpectedly" errors) even though all actual tests pass. This is a known environment limitation, not a code bug.
- `svelte-kit sync` must run before lint/type-check in each frontend workspace. The `quality` script handles this automatically; if running eslint or svelte-check manually, run `npx svelte-kit sync` first in the relevant workspace directory.
- Go 1.25 is required (see `go.mod`). Node.js 22 is required (see `.nvmrc`).
