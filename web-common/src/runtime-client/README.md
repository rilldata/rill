# runtime-client

This folder contains a `svelte-query` client for interfacing with the runtime. It's auto-generated based on the runtime's OpenAPI spec.

## Generating the client

We use [Orval](https://orval.dev) to generate the client. The client is based on the OpenAPI schema in `runtime/api/runtime.swagger.json`. Orval is configured to generate a client that uses [@tanstack/svelte-query](https://tanstack.com/query).

You have to manually re-generate the client when the OpenAPI spec changes. We could automate this step, but for now, we're going to avoid cross-language magic.

To re-generate the client (from the repo root), run:

```script
npm run generate:runtime-client -w web-common
```
