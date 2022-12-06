# `cli`

## Building the CLI

In production builds, the CLI embeds the SPA in `web-local` and the examples in `examples` (from the root of the repo). To create a production build of the CLI with these embedded, run:
```bash
# Build the binary and output it to ./rill
make cli

# To output usage:
./rill

# To run start
./rill start --project dev-project
```

## Running in development

In development, the CLI will serve a dummy frontend and not embed any examples. You can run it like this:
```bash
# Optionally run this to embed the UI and examples in the cli (not needed if using `npm run dev`)
make cli.prepare

# To output usage:
go run ./cli

# To run start:
go run ./cli start --project dev-project
```
