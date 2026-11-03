# Audit work queue — Now / Next / Later

Generated 2026-09-20 after merging `origin/main` (HEAD `008315ff3`) and re-verifying every pending audit finding (changed-file intersection over the 53 PRs merged since the 2026-08-30 verification; each overlapping finding re-checked at its cited location).

**Source of truth for finding detail:** [`../../AUDIT_TODO.md`](../../AUDIT_TODO.md) (107 pending: 39 must-fix · 62 medium · 6 low). Items below reference stable ids (MF-x, A-x…L-x) — full failure scenarios live there and in the linked `audit/*.md` reports. GitHub-issue backlog context: [`../../OPEN_ISSUES_REPORT.md`](../../OPEN_ISSUES_REPORT.md). Competitive roadmap items: [`../COMPETITORS.md`](../COMPETITORS.md).

## Status after the 2026-09-20 re-verification

- **Fixed since 2026-08-30:** MF-33 (canvas store subscription leak), G-4 (falsy filter literals), G-5 (in-list filter marker) — all fell out of the unified expression filter refactor (#9746/#9886). MF-38 partially mitigated.
- **Zero targeted fixes:** as in the previous window, no pending finding was deliberately addressed; the three fixes were side effects of a refactor. The queue only shrinks when items are scheduled — hence this folder.

| File | Contents | Rough scope |
|---|---|---|
| [`NOW.md`](NOW.md) | Exploitable security boundaries, process-wide crash/deadlock risks, irrecoverable data loss, plus one-line correctness fixes | 24 items, ~1–2 sprints |
| [`NEXT.md`](NEXT.md) | Remaining must-fix correctness/races/leaks, security-adjacent mediums, user-visible wrong-result bugs, GitHub backlog hygiene | ~35 items |
| [`LATER.md`](LATER.md) | Remaining mediums by area, robustness lows, needs-verification backlog, product/PLG ideas, competitive-roadmap engineering items | Everything else |

**Working the queue:** take items top-down within a file; check the box here *and* in `AUDIT_TODO.md` when merged; re-run the changed-file intersection against `AUDIT_TODO.md` after each merge window and move anything newly fixed to its resolved register.
