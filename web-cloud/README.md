# web-cloud

This folder contains the cloud frontend implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Generating the client

We use [Orval](https://orval.dev) to generate a client for interacting with cloud backend server (in `server-cloud`). The client is generated in `web-cloud/src/client/gen/` and based on the OpenAPI schema in `server-cloud/api/openapi.yaml`. Orval is configured to generate a client that uses [@sveltestack/svelte-query](https://sveltequery.vercel.app).

You have to manually re-generate the client when the OpenAPI spec changes. We could automate this step, but for now, we're going to avoid cross-language magic.

To re-generate the client, run:

```script
npm run generate:client -w web-cloud
```
