# Verified audit: plg_improvement_ideas

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **19 numbered prioritized ideas** (table rows 1-19). Narrative journey steps and experiment IDs are supporting analysis, not additional findings. Every prioritized idea appears exactly once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| PLG-1 | M / Now | P1 | Make sample generation resilient | `sample-data/generate-sample-data.ts:59-70` still has chat-only indefinite/silent paths | Capability gate, bounded retry, visible failure, offline DuckDB fallback. |
| PLG-2 | L / Now | P2 | Remove fixed five-second splash delay | `welcome/ProjectCards.svelte:62-67` still waits 5000ms and hard-reloads | Navigate on reconcile/entry artifact readiness. |
| PLG-3 | M / Now | P1 | Surface every onboarding failure | `ProjectCards.svelte:76-78`; `actions.ts:162-179`; `ConnectYourDataWidget.svelte:55-82` still hide or wait indefinitely | Shared recoverable error UI plus timeouts/abort. |
| PLG-4 | L / Now | P2 | Create one default dashboard per source | `add-data/manager/steps/import.ts:60-65,90-114` still emits separate Explore and Canvas | Choose one default and offer the second after activation. |
| PLG-5 | L / Now | P1 | Fix `/deploy?org=` route | `web-local/src/routes/(misc)/deploy/+page.ts:4-9` still targets missing `fresh-deploy` | Route to org selection and add deep-link test. |
| PLG-6 | L / Now | P2 | Repair and complete funnel telemetry | `cli/pkg/local/app.go:544-556` still drops AppStart on connector errors and no complete UI funnel is defined | Fix error return and emit stable activation events. |
| PLG-8 | M / Now | P3 | Persistent onboarding checklist | Current welcome/home code has no resource/deploy-backed checklist | Add persisted activation checklist with contextual next action. |
| PLG-9 | M / Now | P2 | Real cloud zero-org empty state | `web-admin/src/routes/+page.svelte:27` still renders `features/home/WelcomeMessage.svelte` for the empty state | Offer in-product org create, live demo, and working local-deploy instructions. |
| PLG-10 | L / Next | P3 | Dashboard-first post-deploy sharing | Deploy routes still treat invite/build as separate terminal surfaces and lack one contextual share success step | Land on live dashboard, then prompt copy/invite. |
| PLG-12 | L / Next | P2 | Explain dimension-less generated dashboards | `generate_metrics_view.go:272-292,456-469`; `generate_canvas_dashboard.go:818-927` still generate KPI-only results silently | Detect zero dimensions/time and show guided remediation. |
| PLG-13 | L / Next | P3 | Intent-aware splash | `web-local/src/routes/(misc)/welcome/+page.svelte` still presents one static set/order of paths | Segment by intent/context and recommend a path. |
| PLG-15 | M / Next | P3 | Activation-state lifecycle messaging | Billing/trial UI exists under `web-admin/src/features/billing`, but no code-backed activation email sequence spans deploy/invite/stall state | Define owned event-driven lifecycle jobs and suppression rules. |
| PLG-16 | L / Next | P3 | CLI deploy nudge | `cli/pkg/local/app.go` has no healthy-dashboard engagement trigger for a deploy hint | Emit one non-intrusive clickable nudge after activation. |
| PLG-17 | M / Later | P3 | Explain-this-dashboard narrative | Generated dashboard navigation has no first-view narrative/fallback | Generate a non-blocking summary with deterministic fallback. |
| PLG-18 | M / Later | P3 | Progressive trial expiry | Current billing/hibernation behavior has no read-only grace policy | Add bounded grace, clear upgrade state, cost controls, and experiments. |

## Fixed / remove

None.

## Partial / needs test

| ID | Original severity | Priority | Existing implementation | Remaining work |
|---|---|---|---|---|
| PLG-7 | M / Now | P2 | Connector reconciliation/validation and database exploration primitives exist (`add-data/manager/steps/connector.ts:97-166`; `connectors/explorer/DatabaseExplorer.svelte`) | Test credentials and preview/select tables before committing connector/secrets. |
| PLG-11 | M / Next | P2 | Deploy/import routes expose basic loading/build states (`web-local .../deploy/+page.svelte:112`; `features/projects/DashboardBuilding.svelte`) | Add reconcile-derived stages, elapsed time, per-source detail, background notification. |
| PLG-14 | M / Next | P2 | Backend `ListExamples` exists (`runtime/server/examples.go:19-37`) | Frontend still uses `welcome/constants.ts:3-28`; extend metadata and add reachability badges/preflight. |
| PLG-19 | M / Later | P3 | Explore/Canvas bookmark creation is implemented (`web-admin/src/features/bookmarks/BookmarksFormDialog.svelte:184-243`) | Add viewer orientation and an explicit bookmark-to-creator path; validate viewer permissions. |

Counts: **15 pending, 0 fixed, 4 partial = 19**.

## Machine-readable inventory

```text
PENDING|plg_improvement_ideas.md|PLG-1|M-Now|P1|-|Resilient sample generation|generate-sample-data.ts:59-70
PENDING|plg_improvement_ideas.md|PLG-2|L-Now|P2|-|Remove five-second delay|ProjectCards.svelte:62-67
PENDING|plg_improvement_ideas.md|PLG-3|M-Now|P1|-|Surface onboarding failures|actions.ts:162-179
PENDING|plg_improvement_ideas.md|PLG-4|L-Now|P2|-|One dashboard per source|import.ts:60-65,90-114
PENDING|plg_improvement_ideas.md|PLG-5|L-Now|P1|-|Fix deploy org route|deploy/+page.ts:4-9
PENDING|plg_improvement_ideas.md|PLG-6|L-Now|P2|-|Funnel telemetry|cli/pkg/local/app.go:544-556
PARTIAL|plg_improvement_ideas.md|PLG-7|M-Now|P2|-|Connection preview before commit|connector.ts:97-166
PENDING|plg_improvement_ideas.md|PLG-8|M-Now|P3|-|Onboarding checklist|features/welcome
PENDING|plg_improvement_ideas.md|PLG-9|M-Now|P2|-|Cloud zero-org empty state|web-admin/src/routes/+page.svelte:27
PENDING|plg_improvement_ideas.md|PLG-10|L-Next|P3|-|Dashboard-first sharing|web-local deploy routes
PARTIAL|plg_improvement_ideas.md|PLG-11|M-Next|P2|-|Legible build progress|DashboardBuilding.svelte
PENDING|plg_improvement_ideas.md|PLG-12|L-Next|P2|-|Dimension-less explainer|generate_metrics_view.go:272-292
PENDING|plg_improvement_ideas.md|PLG-13|L-Next|P3|-|Intent-aware splash|welcome/+page.svelte
PARTIAL|plg_improvement_ideas.md|PLG-14|M-Next|P2|-|Backend-driven example gallery|runtime/server/examples.go:19-37
PENDING|plg_improvement_ideas.md|PLG-15|M-Next|P3|-|Activation lifecycle messaging|web-admin/src/features/billing
PENDING|plg_improvement_ideas.md|PLG-16|L-Next|P3|-|CLI deploy nudge|cli/pkg/local/app.go
PENDING|plg_improvement_ideas.md|PLG-17|M-Later|P3|-|Dashboard narrative|generate_canvas_dashboard.go:818-927
PENDING|plg_improvement_ideas.md|PLG-18|M-Later|P3|-|Progressive trial expiry|web-admin/src/features/billing
PARTIAL|plg_improvement_ideas.md|PLG-19|M-Later|P3|-|Viewer creator on-ramp|BookmarksFormDialog.svelte:184-243
```
