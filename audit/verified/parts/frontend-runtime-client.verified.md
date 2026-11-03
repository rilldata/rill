# Verified audit: frontend-runtime-client

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **11 numbered findings**. Every `frontend-runtime-client-01` through `-11` appears exactly once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| frontend-runtime-client-01 | HIGH | P1 | Failed chat send loses draft and optimistic message | `features/chat/core/conversation.ts:249,378,705` still clears input/adds optimistic state without transactional rollback | Retain the draft, mark the failed message, and expose retry. |
| frontend-runtime-client-02 | MEDIUM | P1 | Tables query is not invalidated after reconcile | `runtime-client/invalidation.ts:218`; `invalidation/resource-invalidators.ts:278`; `web-local/src/features/tables/selectors.ts:30` still use mismatched/omitted keys | Centralize generated query keys and invalidate `oLAPListTables`. |
| frontend-runtime-client-03 | MEDIUM | P1 | Removed queue entries retain abort listeners | `runtime-client/v2/request-queue.ts:129,147,158`; `heap.ts:34` still remove entries without detaching listener/index state | Store cleanup callbacks and invoke them from pop/remove/clear. |
| frontend-runtime-client-04 | LOW | P2 | Priority map keys do not match RPC methods | `runtime-client/v2/request-priorities.ts:9` still names non-matching operations | Generate/type method keys and test that each configured key is reachable. |
| frontend-runtime-client-05 | MEDIUM | P1 | JWT wait ignores request abort | `runtime-client/v2/runtime-client.ts:119` still waits for JWT freshness without the request signal | Race JWT refresh with abort and release queue capacity immediately. |
| frontend-runtime-client-06 | LOW | P2 | SSE resume consumes retry budget | `runtime-client/sse/sse-connection.ts:104,162,178` still routes resume through reconnect/retry | Separate lifecycle resume from failure retries. |
| frontend-runtime-client-07 | LOW | P2 | Cross-resource request priority is ignored | `runtime-client/v2/request-queue.ts:31,176,210` still weights only active/inactive name groups | Incorporate request priority into the heap comparison. |
| frontend-runtime-client-08 | LOW | P2 | Request may send Bearer undefined | `runtime-client/v2/runtime-client.ts:63` still constructs the header from a token that may be cleared while waiting | Re-read/validate JWT after wait; fail auth cleanly if absent. |
| frontend-runtime-client-09 | LOW | P2 | stripUndefined turns array Date into empty object | `runtime-client/v2/strip-undefined.ts:15` still recursively treats Date as a plain object | Preserve Date/non-plain objects and add nested-array tests. |
| frontend-runtime-client-10 | LOW | P2 | Overlapping providers clobber global feature-flag client | `RuntimeProvider.svelte:25,37`; `features/feature-flags.ts:80-85` still expose one global slot | Use owner tokens/reference counting or provider-local context. |
| frontend-runtime-client-11 | LOW | P3 | RequestQueue columnMap is dead state | `runtime-client/v2/request-queue.ts:25,97-102,185,224-229` writes but never schedules from the map | Remove it or use it for explicit per-column fairness. |

## Fixed / remove

None.

## Partial / needs test

None.

Counts: **11 pending, 0 fixed, 0 partial = 11**.

## Machine-readable inventory

```text
PENDING|frontend-runtime-client.md|frontend-runtime-client-01|HIGH|P1|-|Failed chat send loses draft|conversation.ts:249,378,705
PENDING|frontend-runtime-client.md|frontend-runtime-client-02|MEDIUM|P1|-|Tables query not invalidated|invalidation.ts:218
PENDING|frontend-runtime-client.md|frontend-runtime-client-03|MEDIUM|P1|-|Queue abort listener survives removal|request-queue.ts:129,147,158
PENDING|frontend-runtime-client.md|frontend-runtime-client-04|LOW|P2|-|Priority keys mismatch RPC|request-priorities.ts:9
PENDING|frontend-runtime-client.md|frontend-runtime-client-05|MEDIUM|P1|-|JWT wait ignores abort|runtime-client.ts:119
PENDING|frontend-runtime-client.md|frontend-runtime-client-06|LOW|P2|-|SSE resume consumes retry|sse-connection.ts:104,162,178
PENDING|frontend-runtime-client.md|frontend-runtime-client-07|LOW|P2|-|Cross-resource priority ignored|request-queue.ts:31,176,210
PENDING|frontend-runtime-client.md|frontend-runtime-client-08|LOW|P2|-|Bearer undefined race|runtime-client.ts:63
PENDING|frontend-runtime-client.md|frontend-runtime-client-09|LOW|P2|-|Date stripped in arrays|strip-undefined.ts:15
PENDING|frontend-runtime-client.md|frontend-runtime-client-10|LOW|P2|-|Provider clobbers global client|RuntimeProvider.svelte:25,37
PENDING|frontend-runtime-client.md|frontend-runtime-client-11|LOW|P3|-|Dead columnMap|request-queue.ts:25,97-102
```
