# admin

This directory contains the control-plane for multi-user, hosted deployments of Rill.

## Running in development

1. Create a `.env` file at the root of the repo containing:
```
RILL_ADMIN_ENV=development
RILL_ADMIN_DATABASE_DRIVER=postgres
RILL_ADMIN_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
RILL_ADMIN_PORT=8080
RILL_ADMIN_SESSIONS_SECRET=secret
RILL_ADMIN_AUTH_DOMAIN=gorillio-stage.auth0.com
RILL_ADMIN_AUTH_CALLBACK_URL=http://localhost:8080/auth/callback
# Get these from https://auth0.com/ (or ask a colleague)
RILL_ADMIN_AUTH_CLIENT_ID=
RILL_ADMIN_AUTH_CLIENT_SECRET=
```
2. In a separate terminal, run Postgres in the background:
```
docker-compose -f admin/docker-compose.yml up 
```
3. Run the server:
```
go run admin/cmd/main.go
```

## Adding endpoints

We define endpoints using OpenAPI and generate Go handlers and types using [oapi-codegen](https://github.com/deepmap/oapi-codegen). To add a new endpoint:

1. Describe the new endpoint in `admin/api/openapi.yaml`
2. Make sure you have `oapi-codegen` installed by running `go mod tidy`
3. Run: `go generate ./admin/api`
4. Copy the new handler(s) from `admin/api/server.gen.go` into `admin/server/handlers.go` and implement it
