# Verified: drivers-connectors-ai.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **11** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| drivers-connectors-ai-02 | MEDIUM | P2 | Open handle leaks when context is already cancelled | `runtime/connection_cache.go:169-181`; a successful handle is converted to an error after cancellation but is not closed | Close any non-nil handle before returning from this cancellation branch. |
| drivers-connectors-ai-03 | MEDIUM | P1 | Archive-backed admin repo nil-dereferences `r.git` | `runtime/drivers/admin/repo.go:257-260,298-301,328-331,373-376,412-416`; nil-git guard remains inverted for write/watch methods | Reject non-Git/archive repos explicitly before dereferencing, and release the watch lock on every exit. |
| drivers-connectors-ai-04 | MEDIUM | P1 | Connector config update leaves stale handle cached | `runtime/runtime.go:211-253`; eviction still occurs only when `connector == nil` | Evict matching instance/name connections after successful upsert as well as removal. |
| drivers-connectors-ai-05 | MEDIUM | P2 | Slack batch aborts after first recipient failure | `runtime/drivers/slack/notifier.go:68-76,100-107,187-236`; fan-out helpers return immediately on first error | Continue per recipient, aggregate errors, and retry bounded transient failures. |
| drivers-connectors-ai-07 | LOW | P3 | Slack ERROR “Open” link points to edit URL | `runtime/drivers/slack/notifier.go:137-138`; both fields still use `s.EditLink` in the error branch | Populate `OpenLink` from `s.OpenLink`. |
| drivers-connectors-ai-09 | LOW | P3 | Information-schema helper hard-caps tables at 1000 | `runtime/drivers/information_schema.go:68-99`; inner table token causes an error and shadows the schema token | Implement table pagination/compound continuation and rename the token variables. |
| drivers-connectors-ai-10 | LOW | P2 | Admin repo close races on config context | `runtime/drivers/admin/repo.go:697-706,1027-1033`; `close` reads/cancels fields before taking the lock used by writers | Acquire the lock before reading/cancelling configuration context fields. |
| drivers-connectors-ai-11 | LOW | P2 | Virtual repo sync trusts traversal-capable paths | `runtime/drivers/admin/repo_virtual.go:65-84`; direct `filepath.Join` plus remove/write remains outside the new RepoStore resolver | Resolve every virtual path with `drivers.ResolveRepoPath` against the virtual root and reject escapes. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| drivers-connectors-ai-01 | FIXED — remove | `runtime/drivers/repo.go:150-160` adds `ResolveRepoPath`; all file/admin RepoStore read/write/stat/hash/rename/delete paths use it and regression tests cover traversal | PR #9723, commit `6793dbb0265ce4d126cc7206529dfef54660ea82` |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| drivers-connectors-ai-06 | LOW | P2 | NEEDS RESPONSE FIXTURE | `runtime/drivers/gemini/gemini.go:332-335` still dereferences nilable `res.UsageMetadata`; confirm a valid Gemini/proxy response can omit it, then guard it regardless as low-cost hardening. |
| drivers-connectors-ai-08 | LOW | P3 | DORMANT HARDENING | OpenAI/Claude/Gemini `Config()` methods still decode and return secret-bearing config (`runtime/drivers/openai/openai.go:268-272`, `claude/claude.go:229-233`, `gemini/gemini.go:254-258`), but no current API serializes these maps. Redact centrally and add a non-leak contract test. |

Counts: **8 pending, 1 fixed/remove, 2 partial/needs-test, 0 invalid = 11 total**.

## Machine-readable ledger

```text
FIXED|drivers-connectors-ai.md|drivers-connectors-ai-01|HIGH|-|-|RepoStore path traversal|runtime/drivers/repo.go:150-160; PR #9723 6793dbb0265ce4d126cc7206529dfef54660ea82
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-02|MEDIUM|P2|-|Cancelled open leaks handle|runtime/connection_cache.go:169-181
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-03|MEDIUM|P1|-|Archive admin repo nil dereference|runtime/drivers/admin/repo.go:257-260,298-301,328-331,373-376,412-416
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-04|MEDIUM|P1|-|Connector update retains stale handle|runtime/runtime.go:211-253
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-05|MEDIUM|P2|-|Slack fanout aborts on first failure|runtime/drivers/slack/notifier.go:68-76,100-107,187-236
NEEDS_TEST|drivers-connectors-ai.md|drivers-connectors-ai-06|LOW|P2|-|Gemini usage metadata nil dereference|runtime/drivers/gemini/gemini.go:332-335
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-07|LOW|P3|-|Slack error open link is wrong|runtime/drivers/slack/notifier.go:137-138
NEEDS_TEST|drivers-connectors-ai.md|drivers-connectors-ai-08|LOW|P3|-|AI Config methods retain API keys|runtime/drivers/openai/openai.go:268-272; runtime/drivers/claude/claude.go:229-233; runtime/drivers/gemini/gemini.go:254-258
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-09|LOW|P3|-|Information schema table cap|runtime/drivers/information_schema.go:68-99
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-10|LOW|P2|-|Admin repo close config race|runtime/drivers/admin/repo.go:697-706,1027-1033
PENDING|drivers-connectors-ai.md|drivers-connectors-ai-11|LOW|P2|-|Virtual repo path traversal|runtime/drivers/admin/repo_virtual.go:65-84
```
