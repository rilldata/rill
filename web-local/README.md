# web-local

This folder contains the local frontend implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Running in development

To run the web server with an embedded runtime:

1. In `web-local/build-tools/postinstall_runtime.sh`, edit the `RUNTIME_VERSION` constant with the commit hash of the runtime version you want to use.
2. Run `npm run dev`

To run both the web server and an external development runtime:

1. Run `npm run dev -w runtime`
2. Run `RILL_EXTERNAL_RUNTIME=true npm run dev`

Note there is no hot reloading for the runtime code. You need to restart the runtime to see changes.
