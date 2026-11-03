# Verified audit: `deep-ai-performance.md`

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **10** numbered headings. Nine are actionable current findings; `deep-ai-performance-09` is explicitly a refutation and is INVALID as an issue.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| deep-ai-performance-01 | HIGH | P2 | LLM interface has no token streaming | `runtime/drivers/ai.go:10-12`; final unary completion consumed in `runtime/ai/ai.go:1282-1351` | Add provider streaming and forward deltas through a cancellation-safe event path. |
| deep-ai-performance-02 | HIGH | P1 | Large prompts are reprocessed without cache controls | Claude records cache reads but sets no cache-control blocks (`runtime/drivers/claude/claude.go:256-342`); truncation rewrites history in `runtime/ai/ai.go:1608-1664` | Add stable provider cache breakpoints and preserve a stable prefix across truncation. |
| deep-ai-performance-03 | MEDIUM | P2 | Tool calls and context seeding are serial | `runtime/ai/ai.go:1353-1397`; `runtime/ai/analyst_agent.go:111-161`; `runtime/ai/develop_file.go:96-108` | Parallelize independent calls with deterministic result ordering and bounded concurrency. |
| deep-ai-performance-04 | MEDIUM | P2 | Prompt forces 4-6 queries and repeated charts | `runtime/ai/instructions/data/analysis.md:25-40` | Make depth/charting conditional on question complexity and information gain. |
| deep-ai-performance-05 | MEDIUM | P2 | Managed AI adds a unary proxy hop and no per-call model knobs | `runtime/drivers/admin/ai.go:11-36`; defaults at `runtime/drivers/claude/claude.go:130-141` and `runtime/drivers/openai/openai.go:164-175` | Pass model/token/effort settings through managed config and default interactive loops to a latency-appropriate tier. |
| deep-ai-performance-06 | MEDIUM | P1 | Every file write waits for global reconcile plus 500 ms | `runtime/ai/file_write.go:118-156` | Wait only for resources declared by the file/dependencies; replace fixed sleep with state/event synchronization. |
| deep-ai-performance-10 | MEDIUM | P1 | `develop_file` drops its pre-invoked discovery results | `runtime/ai/develop_file.go:96-119` filters by `s.ID()`; correct parent pattern exists at `runtime/ai/developer_agent.go:79-83` | Filter by `s.ParentID`, with a mock-completion assertion that seed results are present. |
| deep-ai-performance-07 | LOW | P2 | Per-file subagents repeatedly cold-start | `runtime/ai/develop_file.go:67-119,121-138`; parent invokes it from `developer_agent.go:87-103` | Reuse shared context/discovery and batch compatible file work after fixing finding 10. |
| deep-ai-performance-08 | LOW | P1 | Agent progress is backpressured under session lock | Unbuffered subscription and locked send at `runtime/ai/ai.go:653,887-896`; client cancellation at `runtime/server/chat.go:387` | Same fix as `ai-02`: deliver outside lock through bounded cancellation-aware channels. |

## Fixed / Remove

None.

## Partial / Needs-test

None.

## Invalid / Remove

| ID | Reason | Evidence |
|---|---|---|
| deep-ai-performance-09 | INVALID — this heading is a refutation, not a defect: normal web turns explicitly select an agent and do not invoke router/feedback agents. | `runtime/ai/router_agent.go:102-142`; `runtime/ai/feedback_agent.go:68-94`; the source itself says “No action needed”. |

Counts: **9 pending, 0 fixed/remove, 0 partial/needs-test, 1 invalid = 10 total**.

## Machine-readable manifest

```text
PENDING|deep-ai-performance.md|deep-ai-performance-01|HIGH|P2|-|No token-level LLM streaming|runtime/drivers/ai.go:10-12; runtime/ai/ai.go:1282-1351
PENDING|deep-ai-performance.md|deep-ai-performance-02|HIGH|P1|-|No prompt caching controls|runtime/drivers/claude/claude.go:256-342; runtime/ai/ai.go:1608-1664
PENDING|deep-ai-performance.md|deep-ai-performance-03|MEDIUM|P2|-|Tool calls execute serially|runtime/ai/ai.go:1353-1397
PENDING|deep-ai-performance.md|deep-ai-performance-04|MEDIUM|P2|-|Prompt mandates excessive queries and charts|runtime/ai/instructions/data/analysis.md:25-40
PENDING|deep-ai-performance.md|deep-ai-performance-05|MEDIUM|P2|-|Managed AI adds proxy hop and lacks tuning|runtime/drivers/admin/ai.go:11-36
PENDING|deep-ai-performance.md|deep-ai-performance-06|MEDIUM|P1|-|File write waits for global reconcile|runtime/ai/file_write.go:118-156
PENDING|deep-ai-performance.md|deep-ai-performance-10|MEDIUM|P1|-|Develop-file seed filter cannot match|runtime/ai/develop_file.go:96-119
PENDING|deep-ai-performance.md|deep-ai-performance-07|LOW|P2|-|Develop-file subagent cold starts per file|runtime/ai/develop_file.go:67-138
PENDING|deep-ai-performance.md|deep-ai-performance-08|LOW|P1|-|Agent progress couples to client drain|runtime/ai/ai.go:653,887-896; runtime/server/chat.go:387
INVALID|deep-ai-performance.md|deep-ai-performance-09|LOW|-|-|Refutation: no extra router or feedback call on normal web turn|runtime/ai/router_agent.go:102-142; runtime/ai/feedback_agent.go:68-94
```
