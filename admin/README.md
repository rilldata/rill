# admin

This directory contains the control-plane for the managed, multi-user version of Rill (currently available on `ui.rilldata.com`).

## Running in development

Run the following command from the repository root to start a full development environment except the admin service:
```bash
rill devtool start cloud --except admin # optional: --reset 
```

For as long as the devtool is running, `rill` commands will target your local development environment instead of `rilldata.com` (you can manually switch environments using `rill devtool switch-env`.)

Then separately start the admin service (and start/stop it when you make code changes):
```bash
go run ./cli admin start
```

### Using Github webhooks in development

The local development environment is not capable of receiving Github webhooks. In most cases, you can just run `rill project reconcile` to manually trigger a reconcile after pushing changes to Github.

Continue reading only if you are making changes to the Github webhooks code and need to these changes specifically.

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

## Adding endpoints

We define our APIs using gRPC and use [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to map the RPCs to a RESTful API. See `proto/README.md` for details.

To add a new endpoint:
1. Describe the endpoint in `proto/rill/admin/v1/api.proto`
2. Re-generate gRPC and OpenAPI interfaces by running `make proto.generate`
3. Copy the new handler signature from the `AdminServiceServer` interface in `proto/gen/rill/admin/v1/api_grpc_pb.go`
4. Paste the handler signature and implement it in a relevant file in `admin/server/`

## Adding a new user preferences field

To add a new preference field for the user, follow these steps:

1. Include a new column named `preference_<name>` in the `users` table. This can be accomplished by appending an appropriate `ALTER TABLE` query to a newly created `.sql` file located within the `postgres/migrations` folder. 
2. In the admin `api.proto` file, incorporate the optional preference field within the `message UserPreferences` definition. 
3. Revise the method definition for UpdateUserPreferences to encompass the handling of the new preference in the respective service. 
4. Adjust the `UpdateUser` SQL query to encompass the new preference field, ensuring that it is included during the update operation.
5. Identify all instances where the `UpdateUser` method is called and update them to include the new preference value.

By meticulously following these steps, the new preference field can be successfully incorporated for the user. Remember to update the database schema, proto file, service method, SQL query, and method invocations to properly accommodate the new preference field.
