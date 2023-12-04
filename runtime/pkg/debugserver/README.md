# `runtime/pkg/debugserver`

This package starts a HTTP server that serves [net/http/pprof](https://pkg.go.dev/net/http/pprof) on port 6060. 

## Basic usage

Visit [http://localhost:6060/debug/pprof/](http://localhost:6060/debug/pprof/) for an overview of available functionality. Note that if you're running `rill start`, you must pass the `--debug` flag.

## Advanced usage

Use the `go tool pprof` tool to sample/explore in more detail. See example invocations on [https://pkg.go.dev/net/http/pprof](https://pkg.go.dev/net/http/pprof).

Two useful flags to pass to `go tool pprof` are:
- `-seconds 10` to reduce the sampling duration
- `-http :9999` which hosts a UI for exploring the results on `localhost:9999`

## Debugging a pod on Kubernetes

Setup port forwarding for the target pod:

```bash
kubectl port-forward -n NAMESPACE POD_NAME 6060:6060
```

Then you can explore profiling on `localhost:6060` as if you were debugging locally.
