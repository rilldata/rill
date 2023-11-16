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
- When the resource is deleted (`.Meta.DeletedOn` will be non-nil)
- When the resource is renamed (`.Meta.RenamedFrom` will be non-nil)
- When all the resource's refs finish running `Reconcile` (even if they return an error)
- When `Controller.Reconcile` is called for the resource
- When the retrigger timestamp returned from an earlier invocation of `Reconcile` falls due

## How the controller schedules reconcilers

- The controller may run `Reconcile` for different resources at the same time.
- The controller will only run one `Reconcile` *per resource name* at a time.
- The controller will not reconcile resources that reference each other at the same time – i.e. it only runs one non-cancelled `Reconcile` call between any given root and leaf node of the resource DAG. This means `Reconcile` does not need to worry about refs changing their state while it's running.
- The controller cancels a reconcile if the resource being reconciled was updated/deleted by another agent than the reconciler itself (self-updates do not cause cancellations).
- The controller cancels reconciles if DAG ancestors need to be reconciled. It will wait for the cancelled reconcile to complete before starting reconcile for the ancestor.
- The controller does not call `Reconcile` for resources with cyclic refs. Instead, it immediately sets an error on them. 
- The controller schedules reconciles of deleted and renamed resources, and waits for them to finish, before scheduling new regular reconciles. It does so in two phases, first letting all deletes finish (resources with `deleted_on != nil`), second letting all renames finish (resources with `renamed_from != nil`).

## Principles for reconciler development

- The implementation of `Reconcile` should be idempotent
- Assume `Reconcile` may be invoked at any time
- The `ctx` may be cancelled at any time. When `ctx` is cancelled, the reconciler should return as soon as possible.
- After the `ctx` has been cancelled, the reconciler can still update its state before returning (by using the `ctx` to call `UpdateState`).
- If the `ctx` is cancelled, you can assume that `Reconcile` will be invoked again shortly.
- Calls to `Reconcile` can run for a long time (as long as they respond quickly to a cancelled `ctx`).
- `Reconcile` should strive to keep a resource's `.State` correct at all times because it may be accessed while `Reconcile` is running to resolve API requests (such as dashboard queries).
- The `Reconciler` struct is shared for all resources of the registered kind for a given instance ID. This enables it to cache (ephemeral) state in-between invocations for optimization.
- The resource's meta and spec (but not state) may be updated concurrently. Calls to `Get` return a clone of the resource, but if the reconciler update's the resource's meta or spec, it must use a lock to read and update it.
