# `runtime/reconcilers/`

## Adding a new resource type

1. Define spec and state schemas in `proto/rill/runtime/v1/resource.proto`
2. Define a reconciler for it in `runtime/reconcilers/`
3. If the resource should be defined in code files,
    - Define parser logic for it in `runtime/compilers/rillv1`
    - Define mapping from parser output to catalog in `runtime/reconcilers/project_parser.go`

## When reconcile is invoked

- When the runtime is restarted
- When the resource's refs or spec is updated
- When the resource is deleted (`.Meta.Deleted` will be true)
- When the resource is renamed (`.Meta.RenamedFrom` will be non-nil)
- When all the resource's refs finish running `Reconcile` (even if they return an error)
- When `Controller.Retrigger` is called for the resource
- When the retrigger timestamp returned from an earlier invocation of `Reconcile` falls due

## Principles for reconciler development

- The controller will run only one `Reconcile` *per resource name* at a time
- The implementation of `Reconcile` should be idempotent
- Assume `Reconcile` may be invoked at any time
- Assume the `ctx` may be cancelled at any time
- If the `ctx` is cancelled, you can assume that `Reconcile` will be invoked again shortly
- Calls to `Reconcile` may run for a long time
- `Reconcile` should strive to keep the resource's state valid at all times. The resource may be accessed while `Reconcile` is running.
