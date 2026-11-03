# Verified audit: plg-first-user-experience

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive actionable count: **22 numbered items**: 12 finding headings (`plg-first-run-01..08`, `10..13`) plus 10 explicitly numbered New Feature Ideas (`NFI-1..10`). Every numbered item is represented exactly once.

## Pending findings

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| plg-first-run-01 | HIGH | P1 | Sample-data CTA can silently dead-end | `welcome/+page.svelte:41`; `OnboardingGenerateSampleData.svelte:32`; `sample-data/generate-sample-data.ts:59-70` still expose an ungated fire-and-forget chat flow with unbounded streaming wait | Gate by capability, bound waits, show failure, and provide deterministic fallback. |
| plg-first-run-02 | MEDIUM | P1 | Standalone AI dashboard generation can spin forever | `entity-management/actions/actions.ts:162-179`; `metrics-views/ai-generation/generateMetricsView.ts:226-251,405-425` still poll for a file after dropping backend rejection | Await the request, terminate polling on error/timeout, retain retry/cancel. |
| plg-first-run-03 | MEDIUM | P1 | `/deploy?org=` redirects to missing route | `web-local/src/routes/(misc)/deploy/+page.ts:4-9` still redirects to `/deploy/fresh-deploy`, for which no route exists | Redirect to the real org select/create flow and add a route test. |
| plg-first-run-04 | MEDIUM | P2 | Example unpack failures are hidden | `welcome/ProjectCards.svelte:76-78`; `onboarding/OnboardingWorkspace.svelte:52-54` still swallow/log only | Show actionable error and retry; preserve selected choice. |
| plg-first-run-05 | MEDIUM | P2 | Example selection always waits five seconds/reloads | `welcome/ProjectCards.svelte:62-67` still uses 5000ms `setTimeout` and `window.location.assign` | Navigate when reconcile/entry artifact is ready, with timeout/error. |
| plg-first-run-06 | MEDIUM | P2 | Add-data still creates model+metrics+explore+canvas | `add-data/manager/steps/import.ts:60-65,90-114,314-379`; `steps/utils.ts:95-120` still creates both dashboards | Pick one default dashboard and offer the other explicitly later. |
| plg-first-run-07 | LOW | P2 | AppStart telemetry is dropped on connector error | `cli/pkg/local/app.go:544-549` still returns the prior nil `err` instead of `c.Err` | Return/log `c.Err` and emit a failure-qualified start event. |
| plg-first-run-08 | LOW | P1 | Reconcile wait compares version strings | `entity-management/actions/actions.ts:95-111` still evaluates `newStateVersion > prevStateVersion` on strings in an unbounded loop | Compare BigInt/numeric versions and add timeout/abort. |
| plg-first-run-10 | LOW | P3 | Welcome examples remain frontend constants | `welcome/constants.ts:3-28`; `runtime/server/examples.go:19-37` remain separate contracts | Extend/use `ListExamples`, including entrypoint/connector metadata. |
| plg-first-run-11 | LOW | P2 | Connector quick picks have no error state | `add-data/ConnectYourDataWidget.svelte:55-82` still has an explicit error-state TODO and no fallback branch | Render retry/error while retaining the full connector action. |
| plg-first-run-12 | LOW | P2 | Numeric-only source yields unexplained KPI-only dashboard | `runtime/server/generate_metrics_view.go:272-292,456-469`; `generate_canvas_dashboard.go:818-927` still emit no warning when dimensions/time are absent | Detect this outcome and guide the user to add dimensions/time. |
| plg-first-run-13 | LOW | P1 | Failed example unpack leaves partial project | `runtime/server/examples.go:97-110` still writes files sequentially without rollback | Stage and atomically commit, or delete writes on failure. |

## Pending new feature ideas

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| NFI-1 | M idea | P1 | Offline deterministic sample-data fallback | `sample-data/generate-sample-data.ts:59-70` still depends exclusively on developer chat | Add bundled DuckDB templates and use them when AI is unavailable/fails. |
| NFI-2 | M idea | P2 | One-click seeded datasets | `welcome/constants.ts:3-28` contains example projects but no local seeded-data chooser | Bundle small Parquet seeds and create project resources directly. |
| NFI-3 | M idea | P2 | Reconcile-driven first-dashboard progress | `welcome/ProjectCards.svelte:62-67` remains fixed-delay; generate/import surfaces remain opaque | Drive progress from resource reconcile state with skip/background actions. |
| NFI-4 | M idea | P3 | Guided onboarding checklist | Current welcome/home flow (`web-common/src/features/welcome`) has no persisted activation checklist | Add resource/deploy-derived checklist and dismissal state. |
| NFI-5 | M idea | P2 | Test connection and preview before import | Connector steps still write/reconcile configuration before full success (`steps/connector.ts:112-166`) | Add credential validation and table preview before commit. |
| NFI-6 | S idea | P3 | Post-deploy share funnel | Deploy flow routes through invite/build surfaces without one consolidated success/share action | Add copy link, invite, open dashboard on successful deploy. |
| NFI-7 | M idea | P2 | Backend-driven example gallery with reachability | `welcome/constants.ts:3-28` remains hardcoded while `runtime/server/examples.go:19-37` exposes partial metadata | Extend API, preflight connectors, and badge requirements. |
| NFI-8 | M idea | P1 | Consistent generation/import recovery | `ProjectCards.svelte:76-78`, `actions.ts:162-179`, `ConnectYourDataWidget.svelte:55-82` retain distinct silent/indefinite failure modes | Standard Retry / no-AI / edit-YAML recovery component. |
| NFI-9 | S idea | P2 | First-run activation funnel events | `cli/pkg/local/app.go:544-556` still has the broken start event and no complete UI funnel contract | Define and emit stable funnel events with path/failure dimensions. |
| NFI-10 | S idea | P3 | CLI deploy-to-share nudge | `cli/pkg/local/app.go` starts the local app but emits no dashboard-health-triggered deploy hint | Print a one-time clickable deploy suggestion after healthy engagement. |

## Fixed / remove

None.

## Partial / needs test

None. PR #9680 introduced inline explores for metrics-view editing, but the add-data import path still explicitly writes separate explore and canvas files, so it does not close finding 06.

Counts: **22 pending, 0 fixed, 0 partial = 22**.

## Machine-readable inventory

```text
PENDING|plg-first-user-experience.md|plg-first-run-01|HIGH|P1|-|Sample-data silent dead-end|generate-sample-data.ts:59-70
PENDING|plg-first-user-experience.md|plg-first-run-02|MEDIUM|P1|-|AI generation infinite poll|actions.ts:162-179
PENDING|plg-first-user-experience.md|plg-first-run-03|MEDIUM|P1|-|Deploy org route 404|deploy/+page.ts:4-9
PENDING|plg-first-user-experience.md|plg-first-run-04|MEDIUM|P2|-|Unpack failure hidden|ProjectCards.svelte:76-78
PENDING|plg-first-user-experience.md|plg-first-run-05|MEDIUM|P2|-|Fixed five-second reload|ProjectCards.svelte:62-67
PENDING|plg-first-user-experience.md|plg-first-run-06|MEDIUM|P2|-|Four resources per source|import.ts:60-65,90-114
PENDING|plg-first-user-experience.md|plg-first-run-07|LOW|P2|-|AppStart dropped|cli/pkg/local/app.go:544-549
PENDING|plg-first-user-experience.md|plg-first-run-08|LOW|P1|-|String reconcile versions|actions.ts:95-111
PENDING|plg-first-user-experience.md|plg-first-run-10|LOW|P3|-|Hardcoded examples|welcome/constants.ts:3-28
PENDING|plg-first-user-experience.md|plg-first-run-11|LOW|P2|-|No connector error state|ConnectYourDataWidget.svelte:55-82
PENDING|plg-first-user-experience.md|plg-first-run-12|LOW|P2|-|Dimensionless dashboard unexplained|generate_metrics_view.go:272-292
PENDING|plg-first-user-experience.md|plg-first-run-13|LOW|P1|-|Partial example unpack|runtime/server/examples.go:97-110
PENDING|plg-first-user-experience.md|NFI-1|M-idea|P1|-|Offline sample generator|generate-sample-data.ts:59-70
PENDING|plg-first-user-experience.md|NFI-2|M-idea|P2|-|Seeded datasets|welcome/constants.ts:3-28
PENDING|plg-first-user-experience.md|NFI-3|M-idea|P2|-|Reconcile progress|ProjectCards.svelte:62-67
PENDING|plg-first-user-experience.md|NFI-4|M-idea|P3|-|Onboarding checklist|features/welcome
PENDING|plg-first-user-experience.md|NFI-5|M-idea|P2|-|Connection preview|steps/connector.ts:112-166
PENDING|plg-first-user-experience.md|NFI-6|S-idea|P3|-|Post-deploy share funnel|web-local deploy routes
PENDING|plg-first-user-experience.md|NFI-7|M-idea|P2|-|Backend example gallery|runtime/server/examples.go:19-37
PENDING|plg-first-user-experience.md|NFI-8|M-idea|P1|-|Standard error recovery|actions.ts:162-179
PENDING|plg-first-user-experience.md|NFI-9|S-idea|P2|-|Activation telemetry|cli/pkg/local/app.go:544-556
PENDING|plg-first-user-experience.md|NFI-10|S-idea|P3|-|CLI deploy nudge|cli/pkg/local/app.go
```
