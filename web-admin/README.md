# web-admin

This folder contains the control plane frontend for a hosted, multi-user Rill (not launched yet). It's implemented with TypeScript and [SvelteKit](https://kit.svelte.dev). 

## Running in development

1. Run the Go backend in `admin` (see its `README` for instructions)
2. Run `npm install -w web-admin`
3. Run `npm run dev -w web-admin`

There's currently no UI to create orgs or projects. You can add some directly using `curl`:
```
# Add an organization
curl -X POST http://localhost:8080/v1/organizations -H 'Content-Type: application/json' -d '{"name":"foo", "description":"org foo"}'

# Add a project
curl -X POST http://localhost:8080/v1/organizations/foo/projects -H 'Content-Type: application/json' -d '{"name":"bar", "description":"project bar"}'
```

## Generating the client

We use [Orval](https://orval.dev) to generate a client for interacting with the backend server (in `admin`). The client is generated in `web-admin/src/client/gen/` and based on the OpenAPI schema in `admin/api/openapi.yaml`. Orval is configured to generate a client that uses [@sveltestack/svelte-query](https://sveltequery.vercel.app).

You have to manually re-generate the client when the OpenAPI spec changes. We could automate this step, but for now, we're going to avoid cross-language magic.

To re-generate the client, run:

```script
npm run generate:client -w web-admin
```

## Building for production

1. Set the `VITE_RILL_ADMIN_URL` environment variable to the URL of the control plane server (e.g. `https://admin.rilldata.com`)