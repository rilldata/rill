# admin

This directory contains the control-plane for multi-user, hosted deployments of Rill.

## Running in development

1. Create a `.env` file at the root of the repo containing:
```
RILL_ADMIN_DATABASE_DRIVER=postgres
RILL_ADMIN_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
RILL_ADMIN_HTTP_PORT=8080
RILL_ADMIN_GRPC_PORT=9090
RILL_ADMIN_AUTH_DOMAIN=gorillio-stage.auth0.com
RILL_ADMIN_AUTH_CALLBACK_URL=http://localhost:8080/auth/callback
# Get these from https://auth0.com/ (or ask a colleague)
RILL_ADMIN_AUTH_CLIENT_ID=
RILL_ADMIN_AUTH_CLIENT_SECRET=
# Hex-encoded comma-separated list of keys. For details: https://pkg.go.dev/github.com/gorilla/sessions#NewCookieStore
RILL_ADMIN_SESSION_KEY_PAIRS=7938b8c95ac90b3731c353076daeae8a,90c22a5a6c6b442afdb46855f95eb7d6
```
2. In a separate terminal, run Postgres in the background:
```
docker-compose -f admin/docker-compose.yml up 
```
3. Run the server:
```
go run ./cli admin start
```

You can now call the local admin server from the CLI by overriding the admin API URL. For example:
```
go run ./cli org create foo --api-url http://localhost:9090
```

## Adding endpoints

We define our APIs using gRPC and use [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to map the RPCs to a RESTful API. See `proto/README.md` for details.

To add a new endpoint:
1. Describe the endpoint in `proto/rill/admin/v1/api.proto`
2. Re-generate gRPC and OpenAPI interfaces by running `make proto.generate`
3. Copy the new handler signature from the `AdminServiceServer` interface in `proto/gen/rill/admin/v1/api_grpc_pb.go`
4. Paste the handler signature and implement it in a relevant file in `admin/server/`
