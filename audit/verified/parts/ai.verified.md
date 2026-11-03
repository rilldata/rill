# Verified audit: `ai.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **13** numbered findings (`ai-01` through `ai-13`, with `ai-11`/`ai-12` under Product Improvements). Each is classified exactly once below. Priority reflects current impact, not the original audit label.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| ai-02 | HIGH | P1 | Subscriber send can hang under session lock | `runtime/ai/ai.go:653,887-896`; `runtime/server/chat.go:387` | Snapshot subscribers under the lock, then use buffered/cancellation-aware delivery outside it. |
| ai-03 | MEDIUM | P1 | Read-only SQL tool can execute mutations | `runtime/ai/query_sql.go:38-41`; `runtime/resolvers/sql.go:163-165` | Enforce a read-only connection/parser allowlist and correct MCP annotations. |
| ai-04 | MEDIUM | P2 | AI metrics limit parsing and lower bound are unsafe | `runtime/ai/metrics_view_query.go:241-246` | Decode numeric types directly and clamp to `[1, AIMaxQueryLimit]`. |
| ai-06 | LOW | P2 | Stack traces enter persisted LLM messages | `runtime/ai/ai.go:991-995,1337` | Log stacks server-side; return a bounded, stack-free error. |
| ai-07 | LOW | P2 | Final response retains only the last text block | `runtime/ai/ai.go:1405-1414` | Concatenate final text blocks in order. |
| ai-08 | LOW | P3 | Title truncation splits UTF-8 | `runtime/ai/router_agent.go:237-240` | Truncate by runes or with a UTF-8-aware helper. |
| ai-10 | LOW | P3 | Unused `Session.RootID` contains panic paths | `runtime/ai/ai.go:901-917` | Remove it, or return errors and add callers/tests. |
| ai-11 | MEDIUM | P2 | Nested agents lack an aggregate call/token budget | `runtime/ai/ai.go:1132-1151`; `runtime/ai/developer_agent.go:102`; `runtime/ai/develop_file.go:136` | Track and enforce a turn/session-wide token and iteration budget. |
| ai-12 | MEDIUM | P2 | Analyst context becomes stale after navigation | `runtime/ai/analyst_agent.go:106-161` | Seed/dedupe context by resource identity rather than first call count. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| ai-01 | FIXED | `runtime/ai/util.go:66-80` rejects `..`; `runtime/drivers/repo.go:70-82` resolves and verifies containment; regression tests cover read/write/repo traversal. | PR #9723, `6793dbb0265ce4d126cc7206529dfef54660ea82` |

## Partial / Needs-test

| ID | Original severity | Priority | Why not cleanly pending/fixed | Current evidence / next test |
|---|---:|---:|---|---|
| ai-05 | LOW | P3 | Fail-open branch remains, but current share-boundary invariants make it unreachable. | `runtime/ai/ai.go:135-152`; add a missing-boundary catalog test and fail closed. |
| ai-09 | LOW | P3 | Locking inconsistency remains latent because the session currently has one driving goroutine. | `runtime/ai/ai.go:535-599,682-738,880-896`; run a concurrent accessor/flush race test. |
| ai-13 | LOW | P3 | Dirty flags are not reset, but current callers flush a fresh session once. | `runtime/ai/ai.go:558-594,886-893`; add a two-flush test and reset per-message dirty state after insert. |

Counts: **9 pending, 1 fixed/remove, 3 partial/needs-test, 0 invalid = 13 total**.

## Machine-readable manifest

```text
FIXED|ai.md|ai-01|HIGH|-|-|File tools permit project-root traversal|runtime/ai/util.go:66-80; runtime/drivers/repo.go:70-82; PR#9723
PENDING|ai.md|ai-02|HIGH|P1|-|Subscriber send can hang under session lock|runtime/ai/ai.go:653,887-896; runtime/server/chat.go:387
PENDING|ai.md|ai-03|MEDIUM|P1|-|Read-only SQL tool can execute mutations|runtime/ai/query_sql.go:38-41; runtime/resolvers/sql.go:163-165
PENDING|ai.md|ai-04|MEDIUM|P2|-|AI metrics limit parsing and lower bound are unsafe|runtime/ai/metrics_view_query.go:241-246
PARTIAL|ai.md|ai-05|LOW|P3|-|Shared-session boundary fails open only if an invariant breaks|runtime/ai/ai.go:135-152
PENDING|ai.md|ai-06|LOW|P2|-|Stack traces enter persisted LLM messages|runtime/ai/ai.go:991-995,1337
PENDING|ai.md|ai-07|LOW|P2|-|Final response retains only last text block|runtime/ai/ai.go:1405-1414
PENDING|ai.md|ai-08|LOW|P3|-|Title truncation splits UTF-8|runtime/ai/router_agent.go:237-240
PARTIAL|ai.md|ai-09|LOW|P3|-|Session locking is inconsistent but currently latent|runtime/ai/ai.go:535-599,682-738,880-896
PENDING|ai.md|ai-10|LOW|P3|-|Unused Session.RootID contains panic paths|runtime/ai/ai.go:901-917
PENDING|ai.md|ai-11|MEDIUM|P2|-|Nested agents lack aggregate budget|runtime/ai/ai.go:1132-1151
PENDING|ai.md|ai-12|MEDIUM|P2|-|Analyst context becomes stale after navigation|runtime/ai/analyst_agent.go:106-161
PARTIAL|ai.md|ai-13|LOW|P3|-|Message dirty flag is never reset|runtime/ai/ai.go:558-594,886-893
```
