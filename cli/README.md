# `cli`

## Building the CLI

In production builds, the CLI embeds the SPA in `web-local` and the examples in `examples` (from the root of the repo). To create a production build of the CLI with these embedded, run:
```bash
make cli
./rill --help
```

## Running in development

In development, the CLI will serve a dummy frontend and not embed any examples. You can run it like this:
```bash
# To output usage:
go run ./cli

# To run start:
go run ./cli start --dir dev-project
```
