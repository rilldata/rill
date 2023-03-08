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
RILL_ADMIN_ALLOWED_ORIGINS=*
# Hex-encoded comma-separated list of keys. For details: https://pkg.go.dev/github.com/gorilla/sessions#NewCookieStore
RILL_ADMIN_SESSION_KEY_PAIRS=7938b8c95ac90b3731c353076daeae8a,90c22a5a6c6b442afdb46855f95eb7d6
# Get these from https://auth0.com/ (or ask a colleague)
RILL_ADMIN_AUTH_DOMAIN=gorillio-stage.auth0.com
RILL_ADMIN_AUTH_CLIENT_ID=
RILL_ADMIN_AUTH_CLIENT_SECRET=
```
2. In a separate terminal, run Postgres in the background:
```
docker-compose -f admin/docker-compose.yml up 
```
3. Run the server:
```
go run ./cli admin start
```
4. Ping the server:
```
go run ./cli admin ping --base-url http://localhost:9090
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


## Working with the github app 
We have setup a github app to listen to push on repositories connected with rill and do automated deployments
The app has access to read `contents` and receive webhooks on `push`.

public link for test app : https://github.com/apps/test-rill-webhooks

Compulsarily set following secrets in .env file (get data from vault / ask a colleague for values):
RILL_ADMIN_GITHUB_APP_SECRET
RILL_ADMIN_GITHUB_APP_ID
RILL_ADMIN_GITHUB_APP_PRIVATE_KEY_PATH (path to private key file in local)
RILL_ADMIN_GITHUB_APP_NAME (set to test app by default)

Also set RILL_ADMIN_GITHUB_APP_ID and RILL_ADMIN_GITHUB_APP_PRIVATE_KEY_PATH in hosted runtime

## Working with the github app in local

the test app currently sends link to a https://github.com/probot/smee.io channel. 
In order for webhooks to be received on local, install smee-client and run following command for receiving web hooks:
smee --url  <smee-channel from vault> --path /event_handler/github --port 8080
