# `proto/`

This directory contains the protocol buffer definitions for all Rill components. Instead of placing `.proto` files in their respective sub-projects, we follow the convention of having a single `proto` folder to make cross-component imports and codegen easier.

We use [Buf](https://buf.build) to lint and generate protocol buffers. The layout and style of our `.proto` files follow their [style guide](https://docs.buf.build/best-practices/style-guide).

## Defining APIs

We define APIs as gRPC services. All APIs should be defined centrally in this directory, but implemented in their respective sub-package of the monorepo. 

For all APIs, we setup [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/) to map the gRPC definitions to a RESTful API. The mapping rules are done inline in the `proto` file, using `google.api.http` annotations ([docs here](https://github.com/googleapis/googleapis/blob/master/google/api/http.proto#L44)). 

Using protocol buffers to define dual RPC and REST interfaces is a technique widely used at Google. We suggest taking a look at their excellent [API design guide](https://cloud.google.com/apis/design/resources), which describes this pattern (notice it's multiple pages).

## Generating

After changing a `.proto` file, you should re-generate the bindings:

1. Install Buf if you haven't already ([docs](https://docs.buf.build/installation))
2. From the repo root, run:
```bash
make proto.generate
```

### Typescript runtime client

We separately have a generated TypeScript client for the runtime in `web-common/src/runtime-client`. If relevant, you can re-generate it by running:

```bash
npm run generate:runtime-client -w web-common
```

(This is not automated as the frontend may currently be pinned to an older version of the runtime.)
