# web-local

This folder contains the local frontend implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Running in development

In development, we run the frontend and backend separately:

1. Start the frontend: `npm run dev-web`
2. Start the backend: `npm run dev-runtime`

You can also run the two together using `npm run dev`. Make sure to wait for both servers to start.
Note there is no hot reloading for the runtime code. You need to restart the runtime and the web server (in that order) to see changes.

Running in development creates a (gitignored) empty project in `dev-runtime`. You can clear it with `npm run clean`.

In production, the frontend is built into a static site and embedded in the CLI. See `cli/README.md` for details. 

More resources:
- [Contributor's guide](../CONTRIBUTING.md) for installing any missing dependencies (like Go)
- [Runtime's README](../runtime/README.md) for how to contribute to the runtime code

## Testing

1. Build the application and the rill cli for E2E tests: `make cli`
2. Run all the tests `npm run test`
