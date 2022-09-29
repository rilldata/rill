# server-cloud

This directory contains the control-plane backend server for cloud deployments.

## Running in development

1. Create a `.env` file at the root of the repo containing:
```
RILL_CLOUD_ENV=development
RILL_CLOUD_DATABASE_DRIVER=postgres
RILL_CLOUD_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
RILL_CLOUD_PORT=8080
```
2. In a separate terminal, run Postgres in the background:
```
cd server-cloud
docker-compose up
```
3. Run the server:
```
go run server-cloud/cmd/main.go
```

## Adding endpoints

We define endpoints using OpenAPI and generate Go handlers and types using [oapi-codegen](https://github.com/deepmap/oapi-codegen). To add a new endpoint:

1. Describe the new endpoint in `server-cloud/api/openapi.yaml`
2. Make sure you have `oapi-codegen` installed by running `go mod tidy`
3. Run: `go generate ./server-cloud/api`
4. Copy the new handler(s) from `server-cloud/api/server.gen.go` into `server-cloud/server/handlers.go` and implement it
