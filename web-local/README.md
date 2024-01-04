# web-local

This folder contains the local frontend implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Running in development

The following command starts a local development environment with hot reloading for the frontend code (restarts are required for backend changes):
```bash
rill devtool start local
```

Running in development creates a (gitignored) empty project in `dev-project`. Pass `--reset` to the command above to clear its state.

## Testing

1. Build the application and the rill cli for E2E tests: `make cli`
2. Run all the tests `npm run test`

## Production builds

In production, the frontend is built into a static site and embedded in the CLI. See `cli/README.md` for details. 
