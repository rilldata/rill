# Verified audit backlog

> **Superseded (2026-09-20):** this directory is frozen at the 2026-07-31 snapshot. The live consolidated tracker is [`../../AUDIT_TODO.md`](../../AUDIT_TODO.md) (re-verified 2026-08-30 and 2026-09-20), and the actionable work queue is [`../todo/`](../todo/README.md). Since this snapshot, the following items recorded here as pending were fixed: metrics-SQL items via #9712–#9726 (already registered), and canvas-05-class store leak (MF-33), G-4, G-5 via #9746/#9886. Use this directory only for historical per-source evidence (`parts/`).

Code snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (latest `origin/main` on 2026-07-31).

The original audit files are preserved as historical evidence. This directory is the refreshed view: every numbered source finding is rechecked against current code and merged PR history, fixed items are removed from active queues, and unresolved work is grouped by current priority.

## Active queues

- [P0 — Must fix now](./P0-NOW.md): **16**
- [P1 — High priority next](./P1-NEXT.md): **142**
- [P2 — Planned fixes](./P2-PLANNED.md): **160**
- [P3 — Later backlog](./P3-LATER.md): **42**
- [Fixed / removed](./FIXED-REMOVED.md): **33**
- [Partial / needs verification](./NEEDS-VERIFICATION.md): **32**
- [Product / PLG proposals](./IDEAS-BACKLOG.md): **84**

The updated historical 42-item ship-blocker rollup is in [MUST-FIX.md](./MUST-FIX.md). The P0/P1 files are the exhaustive current-code inventories and also include unique supplemental findings from `REVIEW_FINDINGS_TODO.md`.

## Disposition summary

| Disposition | Count |
|---|---:|
| FIXED | 29 |
| INVALID | 3 |
| NEEDS_TEST | 13 |
| PARTIAL | 43 |
| PENDING | 417 |
| REMOVE | 1 |
| **Total classified** | **506** |

## Source coverage

Covered: **37/37** substantive audit files.

| Source | Classified findings | Disposition |
|---|---:|---|
| [admin-auth-rbac.md](./parts/admin-auth-rbac.verified.md) | 13 | pending: 10 · needs_test: 3 |
| [admin-billing-provision.md](./parts/admin-billing-provision.verified.md) | 12 | pending: 9 · fixed: 1 · needs_test: 2 |
| [admin-database.md](./parts/admin-database.verified.md) | 9 | pending: 7 · needs_test: 2 |
| [admin-projects-deploy.md](./parts/admin-projects-deploy.verified.md) | 12 | pending: 12 |
| [ai.md](./parts/ai.verified.md) | 13 | pending: 9 · fixed: 1 · partial: 3 |
| [cli.md](./parts/cli.verified.md) | 14 | pending: 11 · fixed: 1 · needs_test: 2 |
| [deep-ai-performance.md](./parts/deep-ai-performance.verified.md) | 10 | pending: 9 · invalid: 1 |
| [deep-audit-canvas.md](./parts/deep-audit-canvas.verified.md) | 14 | pending: 14 |
| [deep-audit-leaderboards.md](./parts/deep-audit-leaderboards.verified.md) | 13 | pending: 13 |
| [deep-audit-pivot.md](./parts/deep-audit-pivot.verified.md) | 14 | pending: 14 |
| [deep-audit-state-management.md](./parts/deep-audit-state-management.verified.md) | 13 | pending: 11 · fixed: 2 |
| [deep-audit-time-dimension-detail.md](./parts/deep-audit-time-dimension-detail.verified.md) | 13 | pending: 13 |
| [deep-cloud-deploy.md](./parts/deep-cloud-deploy.verified.md) | 14 | pending: 14 |
| [deep-connectors.md](./parts/deep-connectors.verified.md) | 20 | pending: 20 |
| [deep-data-security.md](./parts/deep-data-security.verified.md) | 3 | pending: 1 · fixed: 1 · invalid: 1 |
| [deep-metrics-sql.md](./parts/deep-metrics-sql.verified.md) | 14 | pending: 7 · fixed: 5 · partial: 2 |
| [deep-sql-aggregation.md](./parts/deep-sql-aggregation.verified.md) | 16 | pending: 16 |
| [drivers-connectors-ai.md](./parts/drivers-connectors-ai.verified.md) | 11 | pending: 8 · fixed: 1 · needs_test: 2 |
| [drivers-object-store.md](./parts/drivers-object-store.verified.md) | 16 | pending: 12 · fixed: 2 · needs_test: 2 |
| [drivers-olap.md](./parts/drivers-olap.verified.md) | 17 | pending: 17 |
| [frontend-canvas.md](./parts/frontend-canvas.verified.md) | 10 | pending: 9 · fixed: 1 |
| [frontend-dashboard-state.md](./parts/frontend-dashboard-state.verified.md) | 12 | pending: 11 · fixed: 1 |
| [frontend-dashboard-viz.md](./parts/frontend-dashboard-viz.verified.md) | 12 | pending: 12 |
| [frontend-runtime-client.md](./parts/frontend-runtime-client.verified.md) | 11 | pending: 11 |
| [frontend-shared.md](./parts/frontend-shared.verified.md) | 12 | pending: 11 · invalid: 1 |
| [metricsview.md](./parts/metricsview.verified.md) | 8 | pending: 7 · partial: 1 |
| [parser.md](./parts/parser.verified.md) | 10 | pending: 6 · fixed: 2 · partial: 2 |
| [plg-first-user-experience.md](./parts/plg-first-user-experience.verified.md) | 22 | pending: 22 |
| [plg_improvement_ideas.md](./parts/plg_improvement_ideas.verified.md) | 19 | pending: 15 · partial: 4 |
| [product-ideas.md](./parts/product-ideas.verified.md) | 55 | pending: 35 · partial: 20 |
| [queries.md](./parts/queries.verified.md) | 12 | pending: 7 · fixed: 2 · partial: 3 |
| [reconcilers.md](./parts/reconcilers.verified.md) | 11 | pending: 7 · fixed: 3 · partial: 1 |
| [resolvers.md](./parts/resolvers.verified.md) | 8 | pending: 4 · fixed: 3 · partial: 1 |
| [runtime-core.md](./parts/runtime-core.verified.md) | 9 | pending: 5 · fixed: 1 · partial: 3 |
| [runtime-pkg.md](./parts/runtime-pkg.verified.md) | 12 | pending: 10 · fixed: 1 · partial: 1 |
| [runtime-server.md](./parts/runtime-server.verified.md) | 11 | pending: 7 · fixed: 1 · remove: 1 · partial: 2 |
| [web-admin.md](./parts/web-admin.verified.md) | 11 | pending: 11 |

## Rebuild

Run `node audit/verified/build_reports.mjs` after changing a per-source manifest. The build fails on missing sources, duplicate IDs, malformed records, or unexpected source names.
