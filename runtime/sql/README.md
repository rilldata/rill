# `runtime/sql`

This folder contains Go bindings for the SQL native library. 

It relies on a platform-specific native library being present in `runtime/sql/deps`. For example, for macOS ARM, the library should be at `runtime/sql/deps/darwin_arm64/librillsql.dylib`. The library is not checked into Git, but you can generate (download) it by running (from the repo root):
```
go generate ./runtime/sql
```

The `pbast` package contains bindings for the native library's protobuf-based SQL AST (found in `sql/src/main/java/com/rilldata/protobuf/SqlNodeProto.proto`). You can re-generate these by running (from the repo root)
```
go generate ./runtime/sql/pbast
```
