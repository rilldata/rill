# admin

This directory contains the control-plane for multi-user, hosted deployments of Rill.

## Running in development

1. Create a `.env` file at the root of the repo containing:
```
RILL_ADMIN_DATABASE_DRIVER=postgres
RILL_ADMIN_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
RILL_ADMIN_HTTP_PORT=8080
RILL_ADMIN_GRPC_PORT=9090
RILL_ADMIN_EXTERNAL_URL=http://localhost:8080
RILL_ADMIN_FRONTEND_URL=http://localhost:3000
RILL_ADMIN_ALLOWED_ORIGINS=*
# Hex-encoded comma-separated list of keys. For details: https://pkg.go.dev/github.com/gorilla/sessions#NewCookieStore
RILL_ADMIN_SESSION_KEY_PAIRS=7938b8c95ac90b3731c353076daeae8a,90c22a5a6c6b442afdb46855f95eb7d6
# Get these from https://auth0.com/ (or ask a team member)
RILL_ADMIN_AUTH_DOMAIN=gorillio-stage.auth0.com
RILL_ADMIN_AUTH_CLIENT_ID=
RILL_ADMIN_AUTH_CLIENT_SECRET=
# Get these from https://github.com/ (or ask a team member)
RILL_ADMIN_GITHUB_APP_ID=302634
RILL_ADMIN_GITHUB_APP_NAME=rill-cloud-dev
RILL_ADMIN_GITHUB_APP_PRIVATE_KEY=
RILL_ADMIN_GITHUB_APP_WEBHOOK_SECRET=
RILL_ADMIN_GITHUB_CLIENT_ID=
RILL_ADMIN_GIHUB_CLIENT_SECRET=
```
2. In a separate terminal, run Postgres in the background:
```bash
docker-compose -f admin/docker-compose.yml up 
# Data is persisted. To clear, run: docker-compose -f admin/docker-compose.yml down --volumes
```
3. Run the server:
```bash
go run ./cli admin start
```
4. Ping the server:
```bash
go run ./cli admin ping --base-url http://localhost:9090
```

You can now call the local admin server from the CLI by overriding the admin API URL. For example:
```bash
go run ./cli org create foo --api-url http://localhost:9090
```

## Adding endpoints

We define our APIs using gRPC and use [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to map the RPCs to a RESTful API. See `proto/README.md` for details.

To add a new endpoint:
1. Describe the endpoint in `proto/rill/admin/v1/api.proto`
2. Re-generate gRPC and OpenAPI interfaces by running `make proto.generate`
3. Copy the new handler signature from the `AdminServiceServer` interface in `proto/gen/rill/admin/v1/api_grpc_pb.go`
4. Paste the handler signature and implement it in a relevant file in `admin/server/`

## Using the Github App in development

We use a Github App to listen to pushes on repositories connected to Rill to do automated deployments. The app has access to read `contents` and receives webhooks on `git push`.

Github relies on webhooks to deliver information about new connections, pushes, etc. In development, in order for webhooks to be received on `localhost`, we use this proxy service: https://github.com/probot/smee.io.

Setup instructions:

1. Install Smee
```bash
npm install --global smee-client
```
2. Run it (get `IDENTIFIER` from the Github App info or a team member):
```bash
smee --port 8080 --path /github/webhook --url https://smee.io/IDENTIFIER
```

## CLI login/logout

For trying out CLI login add api-url parameter to point to local admin HTTP server like this:
```
go run ./cli auth login --api-url http://localhost:8080/
```
For trying out CLI logout add api-url parameter to point to local admin gRPC server like this:
```
go run ./cli auth logout --api-url http://localhost:9090/
```
