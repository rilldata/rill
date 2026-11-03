# Rill Deep Code Audit

**Date:** 2026-07-19
**Scope:** whole repository — `runtime/`, `admin/`, `cli/` (Go) and `web-common/`, `web-admin/`, `web-local/` (TypeScript/Svelte).
**Method:** 24 area-scoped review agents run one at a time, each followed by an independent **adversarial verifier** that re-opened every cited line, tried to refute the finding, removed false positives, and corrected severities/locations. A final synthesis pass produced the cross-cutting product-idea catalog. Static review only — no builds or tests were executed.

## Start here

- **[../AUDIT_TODO.md](../AUDIT_TODO.md)** — the current consolidated pending list (110 findings: 40 must-fix, 64 medium, 6 low), re-verified item by item against `main` on 2026-08-30. It supersedes the old `MUST-FIX.md` and root `REVIEW_FINDINGS_TODO.md` (both removed; MF ids and the merged-PR register are carried over). Start here if you only read one file.
- **[verified/](./verified/README.md)** — the exhaustive per-source classification of all 506 findings (2026-07-31 snapshot), including P2/P3 long-tail queues and product-idea backlogs.
- **[00-summary.md](./00-summary.md)** — executive roll-up: findings **by severity**, **by priority (P0–P3)**, the cross-cutting themes, the consolidated feature list, PLG highlights, and a suggested Now/Next/Later roadmap.
- **[product-ideas.md](./product-ideas.md)** — 55 synthesized product/feature ideas across 10 themes, with effort ratings, quick wins, and bigger bets.
- **[plg-first-user-experience.md](./plg-first-user-experience.md)** — the first-run / product-led-growth review (funnel friction + activation features). Supplementary brainstorm in `plg_improvement_ideas.md`.

## Later focused reviews

- **[user-management.md](./user-management.md)** (2026-08-10) — user, usergroup, org/project membership and invite management across `admin/`, `cli/` and the `web-admin` user-management UI. 43 findings (6 High · 20 Medium · 17 Low); top risks are member attributes overriding the reserved runtime JWT keys, an unguarded nil dereference in `SetProjectMemberUsergroupRole`, account deletion orphaning organizations, and cache invalidation that never matches the infinite queries the UI renders. Remediation sequencing in **[user-management-fix-plan.md](./user-management-fix-plan.md)**.

## Remediation status (2026-08-30)

- 25 PRs merged 2026-07-11 through 2026-07-31 exactly resolved audit findings (register preserved in [../AUDIT_TODO.md](../AUDIT_TODO.md)).
- The 42 PRs merged 2026-07-31 through 2026-08-30 resolved **zero** of the pending findings; every pending item was individually re-verified against current code on 2026-08-30 and line numbers refreshed. See [../AUDIT_TODO.md](../AUDIT_TODO.md) for the deduplicated inventory.
- The totals below remain the original 2026-07-19 audit snapshot; they are not reduced when a finding is fixed. Individual source findings now carry explicit merged-PR status where applicable.
- Supplemental deep dives included in the all-source pass: canvas, pivot, state management, time-dimension detail, connectors, data security, Metrics SQL, SQL aggregation, AI performance, cloud deploy, and leaderboards. Findings explicitly tagged `Must-fix: no` remain in those reports but are excluded from the ship-blocker count.

## Totals

**278 findings** — **29 High · 90 Medium · 159 Low**. Priority: **P0 28 · P1 44 · P2 47 · P3 159**.
Categories: correctness 89 · bug 40 · security 39 · product-improvement 34 · concurrency 30 · performance 16 · tech-debt 13 · readability 10 · data-loss 5 · simplification 2.

## Area reports

Counts are **High / Medium / Low**. "Health" is the reviewing agent's overall rating after verification.

### Runtime (data plane)

| Area | Report | Health | H/M/L | Top risk |
|---|---|---|---|---|
| Runtime Parser | [parser.md](./parser.md) | fair | 1 / 2 / 7 | Component `value:` panics the whole project parse |
| Reconcilers | [reconcilers.md](./reconcilers.md) | fair | 2 / 5 / 4 | Alert via non-email notifier nil-derefs `adminMeta` |
| Runtime Core (controller/caches/security) | [runtime-core.md](./runtime-core.md) | needs-attention | 2 / 1 / 7 | Resolver cache key omits `AdditionalRules` → data leak |
| Metrics View Engine | [metricsview.md](./metricsview.md) | fair | 0 / 6 / 2 | LIKE→regex doesn't escape metacharacters |
| Resolvers | [resolvers.md](./resolvers.md) | needs-attention | 3 / 2 / 3 | `metrics_annotations` skips `CanAccess()` |
| Legacy Queries | [queries.md](./queries.md) | fair | 1 / 4 / 7 | `MetricsViewAggregation.Export` panics with no time range |
| Runtime Server + Auth | [runtime-server.md](./runtime-server.md) | needs-attention | 2 / 4 / 5 | `ConnectorService` handlers do no authorization |
| AI Assistant / Agents | [ai.md](./ai.md) | needs-attention | 2 / 4 / 7 | File tools allow path traversal outside project root |
| Runtime Shared Utilities | [runtime-pkg.md](./runtime-pkg.md) | needs-attention | 3 / 3 / 6 | `ResolveLocalPath` sandbox bypassable |

### Drivers

| Area | Report | Health | H/M/L | Top risk |
|---|---|---|---|---|
| OLAP Drivers | [drivers-olap.md](./drivers-olap.md) | fair | 1 / 6 / 10 | Pinot `All()` panics on nil table |
| Object Store & Warehouse Drivers | [drivers-object-store.md](./drivers-object-store.md) | needs-attention | 2 / 11 / 3 | Redshift `UNLOAD` breaks on a single quote; `path_prefixes` bypass |
| Connector Framework & AI/Notifier Drivers | [drivers-connectors-ai.md](./drivers-connectors-ai.md) | needs-attention | 1 / 4 / 6 | `RepoStore` file APIs allow path traversal |

### Admin (control plane)

| Area | Report | Health | H/M/L | Top risk |
|---|---|---|---|---|
| Admin Auth & RBAC | [admin-auth-rbac.md](./admin-auth-rbac.md) | fair | 0 / 1 / 12 | Magic tokens replay creator's `admin` flag |
| Admin Projects & Deployments | [admin-projects-deploy.md](./admin-projects-deploy.md) | fair | 0 / 7 / 5 | Editable-deploy delete discards uncommitted work on push fail |
| Admin Billing, Provisioners & Jobs | [admin-billing-provision.md](./admin-billing-provision.md) | needs-attention | 2 / 2 / 8 | clickhouse-static DB-name collision (cross-tenant) |
| Admin Database Layer | [admin-database.md](./admin-database.md) | fair | 0 / 2 / 7 | Billing reads bypass the transaction context |

### CLI

| Area | Report | Health | H/M/L | Top risk |
|---|---|---|---|---|
| CLI & Local App | [cli.md](./cli.md) | fair | 1 / 5 / 8 | `DeployProject` pushes `.env` to the wrong project on name collision |

### Frontend

| Area | Report | Health | H/M/L | Top risk |
|---|---|---|---|---|
| Dashboard State Management | [frontend-dashboard-state.md](./frontend-dashboard-state.md) | needs-attention | 1 / 1 / 10 | `updating` lock never released → state sync frozen |
| Dashboard Visualizations | [frontend-dashboard-viz.md](./frontend-dashboard-viz.md) | fair | 0 / 3 / 9 | Big number hides deltas when comparison value is 0 |
| Canvas Dashboards | [frontend-canvas.md](./frontend-canvas.md) | fair | 1 / 4 / 5 | PDF export omits all tab-group content |
| Runtime Client, SSE & Chat | [frontend-runtime-client.md](./frontend-runtime-client.md) | fair | 1 / 3 / 7 | Failed chat send loses the user's draft |
| Web Admin (Cloud Frontend) | [web-admin.md](./web-admin.md) | fair | 1 / 2 / 8 | Open redirect via `redirect` query param |
| Shared Frontend Features | [frontend-shared.md](./frontend-shared.md) | fair | 1 / 3 / 7 | File rename/delete swallow errors (callers treat as success) |

### Product

| Area | Report | Health | H/M/L | Top risk |
|---|---|---|---|---|
| PLG & First-User Experience | [plg-first-user-experience.md](./plg-first-user-experience.md) | fair | 1 / 5 / 6 | Splash "Generate sample data" ungated, silently dead-ends |

## How to read a finding

Each area report lists findings most-severe-first with: **Category · Confidence · Location (`file:line`)**, the problem, a concrete failure scenario, and a recommended fix. A **Verification Notes** section at the end of each file records what the adversarial verifier removed or corrected (e.g. `frontend-shared` had one theme-CSS-injection finding removed as unreachable dead code).

**Severity** = impact if it triggers. **Confidence** = how sure the (verified) finding is real. **Priority** in the summary is derived as: P0 = High+High-confidence; P1 = other High, or High-confidence Medium in security/data-loss/concurrency/correctness; P2 = remaining Medium; P3 = Low.
