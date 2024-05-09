# web-admin

This folder contains the control plane frontend for the managed, multi-user Rill (available on `ui.rillcloud.com`). It's implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Running in development

The following command starts a development environment with hot reloading for the frontend code (restarts are required for backend changes):
```bash
rill devtool start cloud
```

Press ctrl+C to gracefully stop the development environment. While the development environment is running, any `rill` command you run will target your local development environment instead of the one on `rilldata.com`. (You can manually switch environments using `rill devtool switch-env`.)

All application state is persisted in the (gitignored) `dev-cloud-state` directory. Pass `--reset` to the command above to clear state and start a clean environment.

## Generating the API client

We use [Orval](https://orval.dev) to generate a client for interacting with the backend server (in `admin`). The client is generated in `web-admin/src/client/gen/` and based on the OpenAPI schema in `admin/api/openapi.yaml`. Orval is configured to generate a client that uses [@sveltestack/svelte-query](https://sveltequery.vercel.app).

You have to manually re-generate the client when the OpenAPI spec changes. We could automate this step, but for now, we're going to avoid cross-language magic.

To re-generate the client, run:

```script
npm run generate:client -w web-admin
```

## Building for production

1. Set the `RILL_UI_PUBLIC_RILL_ADMIN_URL` environment variable to the URL of the control plane server (e.g. `https://admin.rilldata.com`)
