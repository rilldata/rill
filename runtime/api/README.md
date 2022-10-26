# `runtime/api/`

This directory contains gRPC (protocol buffer) definitions for the Runtime's external APIs. They're defined in `runtime.proto`, which generates the other files in this directory. The actual handlers are implemented in `runtime/server/`.

We use [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to map the gRPC definitions to a RESTful API. The mappings are done inline in the `proto` file, using `google.api.http` annotations ([docs here](https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L44)). 

Using protocol buffers to define dual RPC and REST interfaces is a technique widely used at Google. We suggest taking a look at their excellent [API design guide](https://cloud.google.com/apis/design/resources), which describes this pattern (notice it's multiple pages).

## Generating

After changing the `.proto` file, you can re-generate the bindings by running (from the repo root):

```bash
go generate ./runtime/api
```

We also have a generated TypeScript client for the runtime in `web-common/src/runtime-client`. If relevant, you can re-generate it by running:

```bash
npm run generate:runtime-client -w web-common
```

(This is not automated as the frontend may be pinned to an older version of the runtime.)
