# Migrating an API from Orval `V1*` types to proto messages

Quick reference for moving RPCs off the legacy JSON bridge.
Reference implementations:
[#9663](https://github.com/rilldata/rill/pull/9663) (profiling queries: oneofs, enums, `Value`, `Timestamp`),
the `RuntimeService` instance APIs (`getInstance`, `listInstances`: `Struct` configs, shared nested types, spec mocks),
and the `RuntimeService` connector APIs (`analyzeConnectors`, `listNotifierConnectors`, `listConnectorDrivers`, `getConnectorOAuthConfig`:
locally built descriptor types, helpers typed with deleted `V1*` types).

## Background

The generated hooks in `v2/gen/*.ts` call ConnectRPC.
By default each method is on a **JSON bridge**: the request is parsed with `Request.fromJson(stripUndefined(...))`,
and the response is converted with `toJson({ emitDefaultValues: true })` and cast to the Orval `V1*` type from `gen/index.schemas.ts`.

A **migrated** method skips both conversions:
it takes `PartialMessage<Request>` (passed straight to the client) and returns the `Response` message instance.
Migration is opt-in per method via `protoMessageMethods` in `config.ts`;
deleting a `V1*` type alone does not migrate anything.

## Steps

1. **Pick a cohesive group** of methods (e.g. all `Column*` queries, all instance APIs).
   Only methods that are actually generated can be migrated:
   check `methodOverrides` in `config.ts` and drop any classified `skip` (e.g. `createInstance`, `editInstance`, `deleteInstance`).
   Include generated methods in the group that have no frontend callers (e.g. `listConnectorDrivers`); they cost nothing to move
   and let their `V1*` response types be deleted.
2. **List the call sites**:
   ```sh
   rg -n "RuntimeServiceGetInstance|runtimeServiceGetInstance\b|V1GetInstanceResponse|\bV1Instance\b" web-common/src web-local/src web-admin/src \
     --glob '!**/runtime-client/v2/gen/**' --glob '!**/runtime-client/gen/index.schemas.ts'
   ```
   Cover raw functions (`<service><Method>`), hooks (`create<Service><Method>`), query option/key builders,
   `getQueryData<V1...>` / `setQueryData` on the method's query key, spec mocks of the raw function,
   and every `V1*` type reachable from the response (including map aliases like `V1InstanceFeatureFlags`).
   Paths contain `[...]` route segments, so pipe with `rg -l0 ... | xargs -0 grep ...` rather than unquoted `$(...)`.
3. **Triage which fields matter.** Grep the call sites for field reads.
   Scalars, strings, repeated scalars, and `map<string, scalar>` fields are unchanged (maps are plain objects in protobuf-es),
   so most call sites need no edits.
   Focus on the rows in the conversion table below: oneofs, enums, `Timestamp`, `int64`, `Struct`/`Value`, `bytes`.
   Also grep for casts on those fields (e.g. `connector.config as Record<string, string>`); a cast can hide the shape change from the compiler.
4. **Register the methods** in `protoMessageMethods` in `config.ts` (camelCase method key, keyed by service name).
5. **Regenerate**: `npm run generate:runtime-client -w web-common` (or `make proto.generate`).
   `validateProtoMessageMethods` fails the run if a listed method is not generated (typo, rename, or `skip` classification).
   Diff `v2/gen/<service>.ts`: migrated methods should have no `fromJson`/`toJson` and should return `Promise<Response>`.
   Paginated methods (e.g. `listInstances`) also get infinite-query hooks;
   type-check the generated file itself, since a template that assumes JSON-shaped responses
   (such as casting to `Record<string, unknown>`) fails on message classes, which have no index signature.
6. **Update call sites** (see the conversion table below). Remove any `as any` casts that were working around the bridge.
7. **Delete dead `V1*` types** from `gen/index.schemas.ts`.
   Check each candidate with `rg -lw <Type> web-common/src web-local/src web-admin/src`.
   - Delete the response, its nested types, and map aliases once nothing else references them.
   - Also delete `V1*` types for sibling `skip` methods that only exist to reference the removed types
     (e.g. `V1CreateInstanceResponse`, `RuntimeServiceEditInstanceBody`); nothing generated uses them.
   - **Keep** nested types still reachable from non-migrated responses (e.g. `V1AIPrompt` via `V1CanvasSpec`).
     Revisit them when their last source migrates: `V1Connector` was kept for `V1ListNotifierConnectorsResponse`
     and deleted when that method moved.
   - A `V1*` type that is also used as a frontend type (built locally, used in spec fixtures) can still be deleted;
     see "Locally built descriptor types" and "Helpers typed with a deleted `V1*` type" below.
   - Request `V1*` types can go once nothing imports them; many requests never had one (already `PartialMessage`).
8. **Update tests**: wrap mocked responses in real messages (see Gotchas);
   add codegen test cases only for new generator behavior.
9. **Verify**:
   ```sh
   (cd web-common && npx vitest run src/runtime-client/v2/codegen <touched specs>)
   npm run quality   # eslint + svelte-check across web-common, web-local, web-admin
   ```
   `npm run quality` runs svelte-check with `--no-tsconfig`, which only checks `.svelte` files.
   To also check `.ts` files, run it from each workspace with its tsconfig:
   `(cd web-common && NODE_OPTIONS=--max-old-space-size=12288 npx svelte-check --threshold error --output machine)`.
   That configuration reports over a thousand errors that already exist on `main`, so compare against a baseline instead of reading the output.
   Filtering to touched files is not enough: it misses errors in untouched consumers.
   Run it for all three workspaces with and without your changes (`git stash push -u`, run, `git stash pop`),
   strip `line:col` from the `ERROR` lines, and `comm` the sorted lists; the "new" side must be empty.
   A `git worktree` baseline is not comparable, because generated files (e.g. i18n messages) are not committed.
   Then exercise the affected UI; most bugs here are runtime shape differences that type-check fine through `any`.

## Conversion table

| JSON bridge (`V1*`)                                               | Proto message                                        | What to do                                                                                                                                                    |
| ----------------------------------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| oneof as sibling optional fields: `x.topK`, `x.cardinality`       | `x.case = { case: "topK", value }`                   | `getOneofValue(x, "case", "topK")` from `lib/proto-utils.ts`; the oneof key is the oneof's name in the proto.                                                 |
| enum as string: `"TIME_GRAIN_DAY"`                                | enum as number: `TimeGrain.DAY`                      | Import the enum from `proto/gen/.../*_pb`. Convert at boundaries with maps in `features/dashboards/proto-state/enum-maps.ts` (`FromProtoTimeGrainMap`, etc.). |
| enum truthiness check (`if (x.grain)`)                            | `UNSPECIFIED` is `0`, which is falsy                 | Check `!== undefined`, or compare against `UNSPECIFIED` explicitly.                                                                                           |
| request enum `QueryService...HistogramMethod.HISTOGRAM_METHOD_FD` | `HistogramMethod.FD`                                 | Use the proto enum in request args.                                                                                                                           |
| `Timestamp` as ISO string                                         | `Timestamp` message                                  | `ts?.toDate()`; build with `Timestamp.fromDate(d)`.                                                                                                           |
| `int64`/`uint64` as string                                        | `bigint`                                             | `Number(x)` where precision is safe; never compare to a number with `===`.                                                                                    |
| `google.protobuf.Struct` as plain object (`c.config?.dsn`)        | `Struct` message (`.fields` map of `Value`)          | `structToJson(c.config)?.dsn` from `lib/proto-utils.ts`; or read `s.fields.foo` and unwrap with `valueAsNumber` / `getOneofValue(v, "kind", ...)`.            |
| `google.protobuf.Value` as plain JSON                             | `Value` message (`kind` oneof)                       | `v.kind.case ? v.toJson() : null`.                                                                                                                            |
| `map<string, T>` alias (`V1InstanceFeatureFlags`)                 | plain object `{ [k: string]: T }`                    | No runtime change; replace the alias with `Record<string, T>`.                                                                                                |
| unset fields omitted or defaulted                                 | always present with proto defaults (`""`, `0`, `[]`) | Messages are never sparse; `?.` on scalar fields is redundant but harmless.                                                                                   |
| unset message fields                                              | `undefined`                                          | Keep `?.` on nested messages.                                                                                                                                 |
| `bytes` as base64 string                                          | `Uint8Array`                                         | Decode explicitly if a string is needed.                                                                                                                      |

## Gotchas

- **Typed through `any` stays broken silently.** Grep for `as any` near the call sites and for components receiving the data as props with `V1*` types; update the prop types too, so the compiler finds the remaining shape mismatches.
- **Props passed down the tree.** Child components often import nested `V1*` types (e.g. `TopKEntry`). Switch them to the `_pb` types (nested messages are named `Parent_Child`, e.g. `TopK_Entry`), or normalize to a plain local type at the boundary when the component needs unwrapped values.
- **Shared helpers typed with a `V1*` type that is kept.** A message is assignable to the `V1*` type when its fields are scalars or strings (e.g. `AIPrompt` to `V1AIPrompt`), so those need no change.
  A message containing a `Struct`, `Timestamp`, enum, or oneof is not (e.g. `Connector` to `V1Connector`).
  Convert at the call site rather than widening the helper, especially when the helper's specs pass plain objects:
  `getOlapEngineLabel({ ...c, config: structToJson(c.config) })`.
  Spread the message and override only the converted fields;
  listing fields by hand silently drops any the helper reads that you missed
  (the instance migration first dropped `provision`, losing the "(Rill-managed)" label),
  and the compiler cannot catch it because the helper's fields are optional.
- **Locally built descriptor types.** Some `V1*` types double as frontend types:
  `V1ConnectorDriver` was built from JSON schemas (`toConnectorDriver`) and in dozens of spec fixtures (`{ name: "s3" }`),
  while also receiving `AnalyzedConnector.driver` messages.
  Replace such a type with `PartialMessage<ConnectorDriver>`: both plain partial objects and the message itself are assignable to it,
  so neither the builders nor the fixtures change beyond the type name.
  This works as long as the code does not read fields whose shape differs (enums, `Struct`, etc.) from locally built values.
- **Helpers typed with a deleted `V1*` type.** If a helper only reads plain fields and its callers or specs pass plain objects,
  give it a small structural parameter type instead of a `_pb` type,
  e.g. `getOlapEngineLabel(connector: { name?: string; type?: string; config?: Record<string, unknown> })`.
  If a parameter is a union of the `V1*` type and an equivalent plain shape (e.g. `ConnectorDriverProperty | { key?: string; type?: string }`),
  just drop the `V1*` branch.
- **Spec mocks and cache seeding must be messages.** `vi.mocked(fn).mockResolvedValue({...})` and `queryClient.setQueryData(key, {...})` with plain objects no longer type-check, and would break any code calling message methods.
  Wrap them: `new GetInstanceResponse({ instance: { olapConnector: "duckdb" } })` (constructor takes a `PartialMessage`).
- **Typed cache reads.** Update `queryClient.getQueryData<V1XResponse>(...)` generics to the `_pb` response type.
- **Svelte `{#each ... (key)}`** keyed on a message field: make sure the key is still a primitive (e.g. unwrap a `Value` first).
- **Responses are class instances.** TanStack's structural sharing only applies to plain objects, so `select` results are not deep-memoized unless the select returns plain data; avoid spreading a message (`{ ...msg }`) and expecting message methods to survive.
- **Shared code that also receives `V1*` data** (e.g. from a non-migrated method or URL state): convert at the boundary rather than widening the type to accept both.
- **Request side is usually free.** `PartialMessage<Request>` accepts plain objects; only enums, int64, timestamps, and oneofs need changes.
- **Codegen specs use fake services named after real ones** (the `mockService` fixture is `RuntimeService` so it gets prefix-based classification).
  Keep migration-list validation in `validateProtoMessageMethods` (run against the real services by `run.ts`), not in `extractMethods`, so fixtures stay independent of which real methods are migrated.

## Checklist

- [ ] Methods are generated (not `skip`) and added to `protoMessageMethods` in `config.ts`, including ones without callers
- [ ] Runtime client regenerated; gen diff shows no JSON bridge for those methods
- [ ] All call sites, cache reads, and child component prop types moved to `_pb` types
- [ ] `Struct` / `Timestamp` / enum / oneof / `int64` reads converted
- [ ] `as any` workarounds removed
- [ ] Spec mocks and `setQueryData` seeds wrapped in `new XResponse(...)`
- [ ] Dead `V1*` types removed from `gen/index.schemas.ts`; shared nested types kept; previously kept types rechecked
- [ ] Frontend-only uses of deleted types switched to `PartialMessage<X>` or a structural type
- [ ] Codegen tests, touched specs, and `svelte-check` pass; affected UI exercised manually
