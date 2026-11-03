# Rill open issues prioritization report

- **Snapshot:** 2026-07-31
- **Scope:** All 122 open, non-pull-request issues in [`rilldata/rill`](https://github.com/rilldata/rill/issues)
- **Repository reference:** `origin/main` at `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)
- **Evidence reviewed:** Issue titles, bodies, comments, labels, assignees, linked timelines and pull requests, plus a targeted implementation and Git-history audit of `origin/main`.

## Executive summary

The open issue list needs a triage pass before it can function as an engineering queue:

- 122 open issues: 108 opened in 2024, 6 in 2025, and 8 in 2026.
- 110 have had no activity in the last 12 months.
- 43 have no label, 71 have no assignee, and all 122 have no milestone.
- None uses the repository's existing `Priority:*` or `Size:*` labels.
- The type taxonomy is split between `Type:Feature`, `enhancement`, `feature request`, `Type:Bug`, and `Bug`.
- Several issues remain open despite direct completion evidence on `main`; several others point to closed, unmerged PRs and must not be treated as complete.

No public issue has enough evidence for `Priority:P0`. The immediate engineering focus should be the two current correctness/access-policy defects, [#9395](https://github.com/rilldata/rill/issues/9395) and [#8850](https://github.com/rilldata/rill/issues/8850). Four old but potentially severe defects should receive time-boxed reproduction before being admitted to the sprint: [#5318](https://github.com/rilldata/rill/issues/5318), [#3835](https://github.com/rilldata/rill/issues/3835), [#5061](https://github.com/rilldata/rill/issues/5061), and [#5430](https://github.com/rilldata/rill/issues/5430).

Recommended disposition across the complete backlog:

| Disposition | Count | Meaning |
|---|---:|---|
| Now — high impact | 7 | Fix now, or reproduce within 48 hours and promote/close |
| Now — small | 12 | Bounded quick wins; pull only after P1 work is staffed |
| Next | 23 | Good candidates for the following 1–2 cycles |
| Later | 28 | Strategic, broad, or low-signal work requiring product commitment |
| Reproduce/refresh now | 14 | Do not estimate or implement until current behavior is confirmed |
| Close/consolidate now | 38 | Completed, superseded, duplicate, answered, or too stale/unscoped |
| **Total** | **122** | Every open issue appears exactly once in this report |

## Code verification results

The second pass checked the implementation at `origin/main`, not merely issue links. It changed the disposition of 12 issues: eight moved into the close queue; two false completion candidates (#3958 and #4405) were restored to Later; code-confirmed #4871 moved into Now-small; and #5624 moved from generic verification into a defined Later item. The 2026-07-31 refresh added newly filed #9780 to Now-small after confirming its localhost URL construction in current code.

### Implemented or superseded in code

| Issue | Code / merged-PR evidence | Recommendation |
|---|---|---|
| [#5326 — Remember File Explorer directories](https://github.com/rilldata/rill/issues/5326) | [PR #9682](https://github.com/rilldata/rill/pull/9682), merged 2026-07-15, persists expand/collapse state per project in `directory-store.ts`. | Close `completed`. This is a direct recent implementation. |
| [#5493 — Improve post-deploy navigation](https://github.com/rilldata/rill/issues/5493) | [PR #7826](https://github.com/rilldata/rill/pull/7826) routes deploys through a loading page and then to the first/originating dashboard; [PR #9435](https://github.com/rilldata/rill/pull/9435) fixes the remaining reconciliation race. | Close `completed`; the delivered flow supersedes the proposed status-page toast. |
| [#5985 — Search OLAP information schema](https://github.com/rilldata/rill/issues/5985) | `LocalProjectTables.svelte` now renders a debounced Search input and passes `%term%` as `searchPattern` to the paginated tables query; the UI arrived with [PR #8864](https://github.com/rilldata/rill/pull/8864) on top of platform [PR #5980](https://github.com/rilldata/rill/pull/5980). | Close `completed` after one connector smoke test. |
| [#5269 — Filter directly from user input](https://github.com/rilldata/rill/issues/5269) | [PR #6773](https://github.com/rilldata/rill/pull/6773) added tested `Contains` and arbitrary `In List` modes, allowing typed input to be applied without waiting for suggestions. | Close `completed/superseded`; retain #5081/#5082 for case and prefix-performance behavior only. |
| [#5469 — Configurable ClickHouse database](https://github.com/rilldata/rill/issues/5469) | [PR #5829](https://github.com/rilldata/rill/pull/5829) added the ClickHouse `database` connector property; current `clickhouse.go` exposes it and applies it to `opts.Auth.Database`. | Close `completed`. |
| [#4877 — Organize Cloud content](https://github.com/rilldata/rill/issues/4877) | [PR #9283](https://github.com/rilldata/rill/pull/9283) added tags to resources and folder-like dashboard organization/filtering; recent [PR #9672](https://github.com/rilldata/rill/pull/9672) added tag authoring in both dashboard visual editors. | Close `completed/superseded` if tags are the intended folder replacement; open a new issue only if true hierarchical folders are still required. |
| [#5645 — Storage usage on Status](https://github.com/rilldata/rill/issues/5645) | [PR #8604](https://github.com/rilldata/rill/pull/8604) added a Status > Tables page with per-model/table database size; [PR #9203](https://github.com/rilldata/rill/pull/9203) added per-connector size estimation and [PR #9080](https://github.com/rilldata/rill/pull/9080) exposed project data usage. | Close `completed` after confirming the original screenshot's “component” means model/table. |
| [#5544 — ClickHouse `unnest` compilation error](https://github.com/rilldata/rill/issues/5544) | The issue has no repro, while ClickHouse unnest handling was implemented/fixed in [#5899](https://github.com/rilldata/rill/pull/5899), [#6184](https://github.com/rilldata/rill/pull/6184), [#7531](https://github.com/rilldata/rill/pull/7531), and [#9582](https://github.com/rilldata/rill/pull/9582). Current dialect code emits `arrayJoin`, with ClickHouse unnest filter tests in `astexpr_test.go`. | Run one array-dimension smoke test, then close `completed/superseded`; reopen only with a current SQL/schema repro. |

### Confirmed still pending

| Issue | Evidence on `origin/main` | Disposition |
|---|---|---|
| [#3958 — Grey out unsupported time grains](https://github.com/rilldata/rill/issues/3958) | `TimeGrainSelector.svelte` and Pivot's `time-pill-store.ts` still filter invalid grains out instead of rendering disabled options. | Retain Later; it is not complete. |
| [#4405 — Previous leaderboard value as a column](https://github.com/rilldata/rill/issues/4405) | `LeaderboardRow.svelte` renders `previousValueString` only while `hovered`; persistent columns remain delta and percent delta. | Retain Later; the exact request is not complete. |
| [#4871 — Comparison colors run out/shift](https://github.com/rilldata/rill/issues/4871) | Both selector components still turn selection index 7+ gray, and colors are derived from the current selection index, so deselection shifts identity even though the palette now has 24 colors. | Promote to Now-small; this is code-confirmed and bounded. |
| [#5428 — Unset column sorting](https://github.com/rilldata/rill/issues/5428) | Pivot and shared tables explicitly set `enableSortingRemoval: false`. | Keep Now-small. |
| [#5601 — Reopen Rows Viewer from Pivot](https://github.com/rilldata/rill/issues/5601) | `RowsViewerAccordion.svelte` sets `manualClose = true` and refuses to reopen on later cell selection while that flag is set—the reported behavior is explicit in current code. | Keep Now-small. |
| [#5713 — Preserve folder context for new resources](https://github.com/rilldata/rill/issues/5713) | Blank files/folders honor `currentDirectory`, but `new-files.ts` still creates typed resources in fixed top-level folders such as `components/` and `models/`. | Keep Now-small. |
| [#5624 — Show comparison dates in Pivot](https://github.com/rilldata/rill/issues/5624) | Pivot carries `comparisonTime` into queries, but `PivotDeltaCell.svelte` has no date/range label or tooltip. | Retain Later with a current UX acceptance criterion. |
| [#5699 — Aliases in `metrics_sql`](https://github.com/rilldata/rill/issues/5699) | `runtime/metricsview/metricssql/parser.go` still returns “aliasing a dimension or measure is not supported.” | Keep Next. |
| [#5755 — Alert criteria on dimensions](https://github.com/rilldata/rill/issues/5755) | Current criteria options are still built from the selected measure; `splitByDimension` is grouping context, not dimension-as-criteria. | Keep Next. |
| [#4838 — Extension-specific File Explorer icons](https://github.com/rilldata/rill/issues/4838) | `getIconComponent` special-cases environment files and Parquet, but `.py`, `.json`, and `.md` still fall through to the generic File icon. | Keep Later. |
| [#9780 — Self-hosted buttons redirect to localhost](https://github.com/rilldata/rill/issues/9780) | `cli/cmd/start/start.go:125-140` hardcodes `LocalURL` as `localhost`; `cli/pkg/local/server.go:108` publishes it as `LoginUrl`, consumed by both the avatar login link and deploy flow. | Add to Now-small; derive the browser-reachable origin safely and cover remote-host/proxy deployments. |

### Partial or adjacent work that does not close the issue

- [#5647](https://github.com/rilldata/rill/issues/5647): the organization thumbnail API/storage and a `ThumbnailSettings.svelte` component exist, but the component is not referenced by the current UI, `app.html` still contains a static `og:image`, and the root layout no longer binds `thumbnailUrl` after [PR #7492](https://github.com/rilldata/rill/pull/7492) reverted that wiring. Retain the issue and verify the unfurl delivery path before estimating.
- [#4821](https://github.com/rilldata/rill/issues/4821): recent [PR #9674](https://github.com/rilldata/rill/pull/9674) fixes a related transient “Metrics view not found” flash for Canvas table/Pivot, but it does not prove the original Dashboard Workspace model-ingestion state is fixed.
- [#6350](https://github.com/rilldata/rill/issues/6350): later hibernation fixes reload config and repair a stuck wake CTA, but current wake state is still local to the CTA rather than visibly shared across Project, Settings, and Users routes.
- [#5081](https://github.com/rilldata/rill/issues/5081)/[#5082](https://github.com/rilldata/rill/issues/5082): direct input now exists, but current dimension search still constructs `%term%` and the backend maps it to case-insensitive matching; case-sensitive and prefix-only modes remain open.

## Ranking rules

- `Priority:P0`: active outage, security exposure, or data loss with no workaround. None is confirmed in the public backlog.
- `Priority:P1`: incorrect analytical results, access-policy breakage, or broad reliability failure.
- `Priority:P2`: important product work with a workaround, narrower impact, or a small quality improvement.
- Later work gets `Needs:Prioritization` instead of a priority label until it is actually committed.
- Size estimates use the repository's definitions: `Size:S` under 100 changed lines, `Size:M` 100–499, `Size:L` 500–1,999, and `Size:XL` 2,000+. They are planning estimates, not promises.

Labels marked **new** in the label strategy section do not exist today. Target labels below are the desired final set, not merely additions to the current labels.

## Now — high-impact queue (7)

| Rank | Issue | Why now | Estimate | Target labels | Gate / acceptance signal |
|---:|---|---|---|---|---|
| 1 | [#9395 — Pivot/Table absolute change ignores new dimension members](https://github.com/rilldata/rill/issues/9395) | Produces materially wrong comparison deltas and rollups; recently reconfirmed; contributor is waiting. | `Size:M` | `Type:Bug`, `Priority:P1`, `Size:M`, `Area:Pivot`, `Area:Time`, `community` | Add an outer-join regression test: missing prior member is treated as zero for absolute delta, while ratio remains null/undefined. |
| 2 | [#8850 — Filter KPI/Table based on metrics policy](https://github.com/rilldata/rill/issues/8850) | A policy-excluded field leaves a broken/loading Canvas component. This is access-policy correctness, not an enhancement. | `Size:M` | `Type:Bug`, `Priority:P1`, `Size:M`, `Area:Dashboard`, `Area:Security` (new), `customer`, `from-slack` | Resolve the Canvas spec against permitted fields and hide invalid components/columns; cover KPI and table with policy tests. |
| 3 | [#5318 — ClickHouse query cancellation can hang](https://github.com/rilldata/rill/issues/5318) | Reported 10-minute reconciler hangs and continued S3 downloads can waste capacity and block deploys. | `Size:L` | `Type:Bug`, `Priority:P1`, `Size:L`, `Area:ClickHouse`, `Area:Reconciler`, `Area:Performance`, `Needs:Reproduction` (new) | Reproduce against current `clickhouse-go`; trace cancellation from controller context through driver and server query. Close if current driver behavior has removed it. |
| 4 | [#3835 — `include` access policy breaks dashboard](https://github.com/rilldata/rill/issues/3835) | If current, a valid access policy makes the dashboard unusable. The issue is old, so confirmation comes before implementation. | `Size:M` | `Type:Bug`, `Priority:P1`, `Size:M`, `Area:Dashboard`, `Area:Security` (new), `Needs:Reproduction` (new) | Turn the supplied YAML into an integration test using `View As`; verify both admin and non-admin field sets. |
| 5 | [#5061 — Last X weeks/months show future dates](https://github.com/rilldata/rill/issues/5061) | Misleading time windows undermine trust in every result. Rill Time has changed substantially since the report. | `Size:M` | `Type:Bug`, `Priority:P1`, `Size:M`, `Area:Time`, `Needs:Reproduction` (new) | Test trailing and calendar ranges across timezone/DST boundaries on current controls. Close if current Rill Time behavior passes. |
| 6 | [#5430 — LIST/VARCHAR dimension filtering fails](https://github.com/rilldata/rill/issues/5430) | A supported-looking dimension can neither filter nor aggregate in the viewer, while export succeeds. | `Size:M` | `Type:Bug`, `Priority:P2`, `Size:M`, `Area:Dashboard`, `Needs:Reproduction` (new) | Reproduce on current DuckDB and define whether list-valued dimensions are supported, rejected with validation, or unnested. |
| 7 | [#6350 — Project wake UI resets when navigating to settings](https://github.com/rilldata/rill/issues/6350) | It is one of the few recently maintained issues, has an owner, and is a bounded Cloud state-consistency defect. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Team:Applications`, `Area:Cloud` (new) | Keep one project-level wake state across Project, Settings, and Users routes; add a navigation test. |

### Immediate maintainer actions

1. Assign [#9395](https://github.com/rilldata/rill/issues/9395) and [#8850](https://github.com/rilldata/rill/issues/8850), respond to the waiting contributors, and attach explicit acceptance tests.
2. Give the five conditional items above a two-day reproduction budget. A failed reproduction should result in closure with tested version/environment, not indefinite return to the backlog.
3. Pair the policy fixes with [#5988](https://github.com/rilldata/rill/issues/5988), the missing security-policy editing end-to-end test.

## Now — small, bounded queue (12)

These are suitable for a quick-win batch after the P1 owners are assigned. Reconfirm the UI location before coding because most reports predate the current application architecture.

| Issue | Bounded outcome | Estimate | Target labels |
|---|---|---|---|
| [#3784 — Warn when a source URI has no file format](https://github.com/rilldata/rill/issues/3784) | Add pre-submit validation/warning and a test for ambiguous file URIs. | `Size:S` | `Type:Improvement`, `Priority:P2`, `Size:S`, `Area:Connectors` |
| [#4821 — Show that a dashboard model is ingesting](https://github.com/rilldata/rill/issues/4821) | Replace the false “does not exist” state with an ingesting state. | `Size:S` | `Type:Improvement`, `Priority:P2`, `Size:S`, `Area:Developer` (new) |
| [#4871 — Comparison selector colors run out and shift](https://github.com/rilldata/rill/issues/4871) | Use the available palette beyond seven selections and key color identity by selected value so deselection does not recolor the survivors; add both regressions. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Dashboard`, `nit` |
| [#4917 — SQL autocomplete fails for quoted table names](https://github.com/rilldata/rill/issues/4917) | Normalize quoted identifiers in completion lookup and add parser cases. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Developer` (new) |
| [#5428 — Allow column sorting to return to unset](https://github.com/rilldata/rill/issues/5428) | Enable the documented third sort state and add an interaction test. | `Size:S` | `Type:Feature`, `Priority:P2`, `Size:S`, `Area:Pivot` |
| [#5601 — Pivot cell no longer reopens model viewer](https://github.com/rilldata/rill/issues/5601) | Restore expansion after the viewer has been minimized; add click-state coverage. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Pivot` |
| [#5647 — Improve shared-dashboard thumbnail](https://github.com/rilldata/rill/issues/5647) | Trace the crawler/unfurl path, expose or remove the dormant thumbnail setting component, and use the stored organization thumbnail for `og:image`; current app HTML remains static after the earlier wiring was reverted. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Team:Applications`, `Area:Cloud` (new), `Needs:Verification` (new) |
| [#5713 — New File Explorer item loses folder context](https://github.com/rilldata/rill/issues/5713) | Seed the create path from the current directory and test nested folders. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Developer` (new) |
| [#5799 — `metrics_sql` day truncation is formatted incorrectly](https://github.com/rilldata/rill/issues/5799) | Preserve day precision in display formatting; results are already correct. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Time`, `Area:Metrics` (new) |
| [#5970 — Generated Metrics View YAML order is inconsistent](https://github.com/rilldata/rill/issues/5970) | Use one serializer/order for generation and visual editing. | `Size:S` | `Type:Improvement`, `Priority:P2`, `Size:S`, `Area:Developer` (new), `Area:Metrics` (new) |
| [#5988 — Add E2E coverage for editing security policies](https://github.com/rilldata/rill/issues/5988) | Add edit/save/view-as regression coverage and include policy-excluded fields. | `Size:S` | `Type:Improvement`, `Priority:P2`, `Size:S`, `tests`, `Area:Security` (new) |
| [#9780 — Self-hosted buttons redirect to localhost](https://github.com/rilldata/rill/issues/9780) | Stop publishing the CLI's hardcoded localhost URL to browser login/deploy links; resolve a trusted request/public origin and test non-localhost plus reverse-proxy deployments. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Developer` (new), `Area:Auth` (new) |

## Next — following 1–2 cycles (23)

These have a reasonably clear user or engineering outcome, but should follow the Now queue. `Priority:P2` here means “next after committed P1 work,” not “start all 25.”

| Issue | Recommendation | Estimate | Target labels |
|---|---|---|---|
| [#3896 — Add dependency validation to CI](https://github.com/rilldata/rill/issues/3896) | Define allowed package-layer imports and introduce the checker in warning mode before making CI blocking. | `Size:M` | `Type:Improvement`, `Priority:P2`, `Size:M`, `tests`, `tech debt` |
| [#4260 — Support a source `limit` property](https://github.com/rilldata/rill/issues/4260) | Scope a development-only limit with explicit connector semantics and project-level defaults. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Connectors`, `customer` |
| [#4366 — Custom time-grain formatting](https://github.com/rilldata/rill/issues/4366) | Unify with #8768 under one temporal-formatting design shared by Explore, Pivot, and Canvas. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Time`, `Area:Dashboard` |
| [#4898 — Show dashboard 4xx errors inline](https://github.com/rilldata/rill/issues/4898) | Audit current surfaces, retain the work already delivered by PR #5171, and finish only missing Pivot/detail states. | `Size:M` | `Type:Improvement`, `Priority:P2`, `Size:M`, `Area:Dashboard` |
| [#4933 — Duplicate SQL columns missing from rows response](https://github.com/rilldata/rill/issues/4933) | Decide a stable duplicate-column naming contract before changing the API; then align preview and downstream references. | `Size:L` | `Type:Bug`, `Priority:P2`, `Size:L`, `Area:Developer` (new), `Team:Platform`, `Needs:Scoping` |
| [#5026 — Auto-populate all dimensions](https://github.com/rilldata/rill/issues/5026) | Add an explicit “all columns as dimensions” generation option with collision/exclusion rules. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Developer` (new), `Area:Metrics` (new), `customer` |
| [#5081 — Allow case-sensitive `LIKE`](https://github.com/rilldata/rill/issues/5081) | Design together with #5082 as case and match-position settings for the advanced dimension filter delivered after these issues were filed. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:ClickHouse`, `Area:Performance` |
| [#5082 — Allow prefix rather than contains search](https://github.com/rilldata/rill/issues/5082) | Implement a per-dimension search mode so indexed ClickHouse prefixes avoid `%term%` scans. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:ClickHouse`, `Area:Performance` |
| [#5199 — Better ClickHouse connection errors](https://github.com/rilldata/rill/issues/5199) | Re-test after the connection-cache retry fix; keep only actionable connectivity guidance that can be detected reliably. | `Size:S` | `Type:Improvement`, `Priority:P2`, `Size:S`, `Area:ClickHouse`, `Area:Connectors` |
| [#5284 — Fix pagination in GitHub list APIs](https://github.com/rilldata/rill/issues/5284) | Inventory every GitHub list call, add pagination helpers, and cover a result set over one page. | `Size:M` | `Type:Improvement`, `Priority:P2`, `Size:M`, `Team:Platform`, `tech debt` |
| [#5455 — Case-only model rename collision](https://github.com/rilldata/rill/issues/5455) | Define case normalization across files, catalog resources, and ClickHouse tables; add rename coverage. | `Size:M` | `Type:Bug`, `Priority:P2`, `Size:M`, `Area:ClickHouse`, `Area:Reconciler` |
| [#5547 — Pivot export should respect selected granularity](https://github.com/rilldata/rill/issues/5547) | Carry grain metadata into export formatting without changing the underlying timestamp semantics. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Pivot`, `Area:Time`, `customer` |
| [#5588 — Preview underlying OLAP data types](https://github.com/rilldata/rill/issues/5588) | Return both logical/runtime and OLAP-native types, then select the appropriate UI display. | `Size:M` | `Type:Improvement`, `Priority:P2`, `Size:M`, `Area:Connectors`, `Area:Developer` (new) |
| [#5633 — Preserve chart brush after refresh](https://github.com/rilldata/rill/issues/5633) | Rehydrate the brush from the selected scrub range and add refresh/navigation coverage. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Dashboard`, `Area:Time` |
| [#5634 — Update stacked-area opacity for brush selection](https://github.com/rilldata/rill/issues/5634) | Apply the established grouped-bar selection styling to stacked area. | `Size:S` | `Type:Improvement`, `Priority:P2`, `Size:S`, `Area:Dashboard`, `Area:Time` |
| [#5681 — File header jumps after timestamp-chart zoom](https://github.com/rilldata/rill/issues/5681) | Reproduce on the current profile chart, then isolate scroll/focus cleanup during route change. | `Size:S` | `Type:Bug`, `Priority:P2`, `Size:S`, `Area:Developer` (new), `Needs:Reproduction` (new) |
| [#5699 — Dimension aliases in `metrics_sql`](https://github.com/rilldata/rill/issues/5699) | Add alias semantics to the parser/AST rather than overloading the current dimension struct. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Metrics` (new), `Team:Platform` |
| [#5755 — Alert criteria on dimensions](https://github.com/rilldata/rill/issues/5755) | Implement the written text-operator acceptance criteria and validate cardinality/query cost. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Alerts` |
| [#5975 — Keep two recent Developer tabs active](https://github.com/rilldata/rill/issues/5975) | Coordinate tab recency via `BroadcastChannel` with a local-storage fallback and cleanup tests. | `Size:M` | `Type:Improvement`, `Priority:P2`, `Size:M`, `Area:Developer` (new), `Area:Performance` |
| [#6399 — ClickHouse-specific source forms](https://github.com/rilldata/rill/issues/6399) | Re-scope from the closed, unmerged draft PR #6400 against the current Add Data flow. | `Size:M` | `Type:Improvement`, `Priority:P2`, `Size:M`, `Area:ClickHouse`, `Area:Connectors`, `Area:Developer` (new) |
| [#8085 — Template model names in Metrics Views](https://github.com/rilldata/rill/issues/8085) | Define which Metrics View properties are templatable and validate references after template resolution. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Metrics` (new), `Team:Platform` |
| [#8768 — ISO temporal identifiers without losing temporal type](https://github.com/rilldata/rill/issues/8768) | Fold into the time-formatting design with #4366; support ISO week/year labels while retaining temporal encoding. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Time`, `Area:Dashboard`, `from-slack` |
| [#9130 — Heatmap row/column normalization](https://github.com/rilldata/rill/issues/9130) | Add a documented `normalize` option with filtered row/column denominator tests and export behavior. | `Size:M` | `Type:Feature`, `Priority:P2`, `Size:M`, `Area:Dashboard` |

## Later — retain only with product sponsorship (28)

These are legitimate ideas, but they are broad, strategic, low-signal, or dependent on a product decision. Do not apply a `Priority:*` label until an owner commits to discovery or delivery. A small estimate alone does not make an old cosmetic request urgent.

| Issue | Why later | Estimate | Target labels |
|---|---|---|---|
| [#3760 — Local SSD serving with persistent GCS backup](https://github.com/rilldata/rill/issues/3760) | Storage architecture with no public requirements or acceptance criteria. | `Size:XL` | `Type:Feature`, `Size:XL`, `Team:Infra`, `Team:Platform`, `infrastructure`, `Needs:Scoping` |
| [#3854 — Persistent storage for distributed workers](https://github.com/rilldata/rill/issues/3854) | Depends on the ETL/serving topology, isolation, quotas, and system connector design. | `Size:XL` | `Type:Feature`, `Size:XL`, `Team:Infra`, `Team:Platform`, `infrastructure`, `Needs:Scoping` |
| [#3958 — Grey out rather than hide time grains](https://github.com/rilldata/rill/issues/3958) | Current time controls and Pivot still filter invalid grains out, so the feature is not delivered; retain only if disabled options remain a desired UX. | `Size:S` | `Type:Feature`, `Size:S`, `Area:Time`, `Needs:Prioritization` |
| [#4007 — Reusable dimension/measure sets](https://github.com/rilldata/rill/issues/4007) | Valuable reuse primitive, but it changes Metrics View composition and override semantics. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Metrics` (new), `customer`, `Needs:Scoping` |
| [#4235 — Local LLM support](https://github.com/rilldata/rill/issues/4235) | Requires a supported provider/configuration model, security posture, and ongoing compatibility. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Developer` (new), `Team:Platform`, `Needs:Prioritization` |
| [#4405 — Show previous value as leaderboard column](https://github.com/rilldata/rill/issues/4405) | Current code still displays the previous value only on hover; decide whether a persistent column is worth the additional table width before implementing it. | `Size:M` | `Type:Feature`, `Size:M`, `Area:Dashboard`, `Needs:Prioritization` |
| [#4407 — Distribution into locked-down environments](https://github.com/rilldata/rill/issues/4407) | Enterprise packaging/distribution strategy, not a bounded code change. | `Size:L` | `Type:Feature`, `Size:L`, `Area:CLI`, `infrastructure`, `community`, `Needs:Scoping` |
| [#4421 — Cross-dashboard filters](https://github.com/rilldata/rill/issues/4421) | Needs navigation, state ownership, compatibility, and URL/bookmark semantics. Keep this as the canonical issue over #4044. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Dashboard`, `customer`, `Needs:Scoping` |
| [#4429 — Global asset search](https://github.com/rilldata/rill/issues/4429) | Broad indexing and navigation feature without current product signal. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Cloud` (new), `Area:Developer` (new), `customer`, `Needs:Prioritization` |
| [#4445 — Custom cross-platform scrollbar](https://github.com/rilldata/rill/issues/4445) | Cosmetic platform work with accessibility and maintenance tradeoffs; no recent demand. | `Size:M` | `Type:Improvement`, `Size:M`, `Team:Applications`, `design`, `Needs:Prioritization` |
| [#4577 — Custom API management UI](https://github.com/rilldata/rill/issues/4577) | Spans local editing, validation, execution, Cloud inventory, and results preview. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Developer` (new), `Area:Cloud` (new), `Needs:Scoping` |
| [#4828 — Close Rill from the Developer UI](https://github.com/rilldata/rill/issues/4828) | Requires safe local-process ownership semantics; CLI stop remains a workaround. | `Size:M` | `Type:Feature`, `Size:M`, `Area:CLI`, `Area:Developer` (new), `community`, `Needs:Prioritization` |
| [#4838 — Extension-specific File Explorer icons](https://github.com/rilldata/rill/issues/4838) | Small, but purely cosmetic and unsupported by recent demand. | `Size:S` | `Type:Improvement`, `Size:S`, `Area:Developer` (new), `nit`, `Needs:Prioritization` |
| [#4863 — Forecasting in time-series charts](https://github.com/rilldata/rill/issues/4863) | Requires an explicit statistical/product model, evaluation, confidence display, and alert semantics. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Dashboard`, `Area:Time`, `customer`, `Needs:Scoping` |
| [#5040 — Include and exclude the same dimension](https://github.com/rilldata/rill/issues/5040) | Compound filter semantics touch query generation, URL state, and UI; current workaround is multiple filters/models. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Dashboard`, `customer`, `community`, `Needs:Scoping` |
| [#5101 — MongoDB connector](https://github.com/rilldata/rill/issues/5101) | A new connector needs a decision between ingestion and direct query plus schema/type behavior. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Connectors`, `customer`, `Needs:Scoping` |
| [#5330 — Use Pivot in Dimension Detail](https://github.com/rilldata/rill/issues/5330) | Cross-surface component reuse needs current UX and performance requirements. | `Size:L` | `Type:Improvement`, `Size:L`, `Area:Pivot`, `Area:Dashboard`, `Needs:Scoping` |
| [#5495 — Create mapping tables in Developer](https://github.com/rilldata/rill/issues/5495) | Requires lifecycle, persistence, editing, and deployment semantics for user-authored data. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Developer` (new), `Area:Security` (new), `Needs:Scoping` |
| [#5612 — Raw data viewer for Time Dimension Detail](https://github.com/rilldata/rill/issues/5612) | Potentially expensive drill-through requiring limits, policy handling, and a current design. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Dashboard`, `Area:Pivot`, `Needs:Scoping` |
| [#5624 — Show comparison dates in Pivot](https://github.com/rilldata/rill/issues/5624) | Comparison queries carry the range, but Pivot still shows no comparison date/range label. Add a current tooltip/header design before scheduling. | `Size:M` | `Type:Feature`, `Size:M`, `Area:Pivot`, `Area:Time`, `Needs:Prioritization` |
| [#5648 — Pin values on the Time Dimension graph](https://github.com/rilldata/rill/issues/5648) | Interaction design is clear at a concept level but lacks current demand and accessibility behavior. | `Size:M` | `Type:Feature`, `Size:M`, `Area:Dashboard`, `Area:Time`, `design`, `Needs:Prioritization` |
| [#5664 — Group-scoped dashboard bookmarks](https://github.com/rilldata/rill/issues/5664) | Requires ownership, permissions, precedence, and migration rules for user/group/global bookmarks. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Dashboard`, `Area:Cloud` (new), `Needs:Scoping` |
| [#5775 — Vertically resize Time Dimension Detail](https://github.com/rilldata/rill/issues/5775) | Useful but low impact; needs a current responsive-layout design. | `Size:M` | `Type:Feature`, `Size:M`, `Area:Dashboard`, `Area:Time`, `customer`, `design`, `Needs:Prioritization` |
| [#5798 — Trino connector](https://github.com/rilldata/rill/issues/5798) | Direct-query performance conflicts with Rill's query-heavy workload; requires a connector strategy decision. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Connectors`, `community`, `Needs:Scoping` |
| [#6357 — Labeled custom time ranges in YAML](https://github.com/rilldata/rill/issues/6357) | Rill Time covers expressions, but named range configuration and compatibility still need design. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Time`, `Needs:Scoping` |
| [#8711 — Exportable metrics dictionary](https://github.com/rilldata/rill/issues/8711) | Needs source-of-truth, editable-vs-generated, cross-project, and export format decisions. | `Size:L` | `Type:Feature`, `Size:L`, `Area:Metrics` (new), `Area:Cloud` (new), `Needs:Scoping` |
| [#8756 — Custom role creation](https://github.com/rilldata/rill/issues/8756) | A major authorization model and migration project; should start with a security/product design, not UI implementation. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Security` (new), `Area:Cloud` (new), `from-slack`, `Needs:Scoping` |
| [#8911 — Arrow Flight SQL for StarRocks](https://github.com/rilldata/rill/issues/8911) | Strong benchmark, but PR #8910 closed unmerged and maintainers raised generic-driver/dialect questions. Resolve architecture first. | `Size:XL` | `Type:Feature`, `Size:XL`, `Area:Connectors`, `Area:Performance`, `community`, `Needs:Scoping` |

## Reproduce or refresh now (14)

These issues are too old or incomplete to enter an engineering queue. Give each a current-version check and a 14-day response window. If it cannot be reproduced and no owner supplies updated acceptance criteria, close it with the tested commit, browser/OS, and dataset shape.

| Issue | Required triage | Resolution rule | Target labels |
|---|---|---|---|
| [#3792 — Flush catalog when adding a trigger](https://github.com/rilldata/rill/issues/3792) | Confirm ad-hoc triggers and the stated crash window still exist in the current catalog architecture. | Keep only with a persistence test that fails on `main`; otherwise close superseded. | `Type:Improvement`, `Area:Reconciler`, `Needs:Verification` (new) |
| [#3807 — Model results missing when typing slowly](https://github.com/rilldata/rill/issues/3807) | Request a current repro and test editor debounce/query cancellation. | Promote to Next if reproducible; otherwise close. | `Type:Bug`, `Area:Developer` (new), `Needs:Reproduction` (new) |
| [#3820 — Do not report obviously invalid queries](https://github.com/rilldata/rill/issues/3820) | Inspect current reporting filters and quantify noisy invalid-query events. | Rewrite with measurable telemetry acceptance criteria or close. | `Type:Improvement`, `Area:Telemetry`, `Needs:Verification` (new) |
| [#3949 — Model row total does not update after switching](https://github.com/rilldata/rill/issues/3949) | Exercise model switching with different row counts in current Developer. | Quick-win if reproducible; otherwise close. | `Type:Bug`, `Area:Developer` (new), `Needs:Reproduction` (new) |
| [#3960 — Hiding a field cancels outstanding queries](https://github.com/rilldata/rill/issues/3960) | Re-test after state-management and Svelte changes; measure whether cancellation causes a user-visible failure. | Merge with #4213 if the same cause remains; close if cancellation is harmless/currently resolved. | `Type:Bug`, `Area:Dashboard`, `Area:Performance`, `Needs:Reproduction` (new) |
| [#4001 — Missing date column in Time Dimension table](https://github.com/rilldata/rill/issues/4001) | Recreate the stated 24-hour/day-grain boundary in multiple timezones. | Promote as P1 only if data is omitted, not merely labeled differently. | `Type:Bug`, `Area:Dashboard`, `Area:Time`, `Needs:Reproduction` (new) |
| [#4213 — Leaderboard filtering shows cancelled query](https://github.com/rilldata/rill/issues/4213) | Re-test and compare with #3960. | Consolidate into one current state-management issue or close both. | `Type:Bug`, `Area:Dashboard`, `Area:Performance`, `Needs:Reproduction` (new) |
| [#4258 — Misaligned measures without time-series data](https://github.com/rilldata/rill/issues/4258) | Compare current responsive layout with the issue and linked merged PR #4229. | Close if current layout is acceptable; otherwise replace screenshots/acceptance criteria. | `Type:Bug`, `Area:Dashboard`, `Needs:Verification` (new) |
| [#4579 — Cast numbers to strings during schema relaxation](https://github.com/rilldata/rill/issues/4579) | Test mixed-schema ingestion against the current DuckDB version and relaxation path. | Keep only with failing fixtures and explicit coercion rules. | `Type:Improvement`, `Area:Connectors`, `Team:Platform`, `Needs:Verification` (new) |
| [#4638 — One row and one column dimension yields empty Pivot](https://github.com/rilldata/rill/issues/4638) | Reproduce on the current Pivot with a checked-in fixture. | If confirmed, promote to P1/`Size:M`; otherwise close. | `Type:Bug`, `Area:Pivot`, `Needs:Reproduction` (new) |
| [#4895 — Snowflake DBUS warning on WSL](https://github.com/rilldata/rill/issues/4895) | Test current CLI and Snowflake dependency on WSL with no DBUS variable. | Close if dependency upgrades removed it; otherwise suppress/configure narrowly. | `Type:Bug`, `Area:CLI`, `Area:Connectors`, `Needs:Verification` (new) |
| [#5379 — Drag-and-drop briefly shows 404](https://github.com/rilldata/rill/issues/5379) | Re-test the current upload/copy/reconcile sequence. | Quick-win if reproducible; otherwise close. | `Type:Bug`, `Area:Developer` (new), `Needs:Reproduction` (new) |
| [#5958 — Example project sometimes hangs](https://github.com/rilldata/rill/issues/5958) | Identify the example, environment, network state, and current loading phase. | Promote if deterministic; close if no current repro can be obtained. | `Type:Bug`, `Area:Developer` (new), `Needs:Reproduction` (new) |
| [#8573 — Time interval always says invalid UTC](https://github.com/rilldata/rill/issues/8573) | Reporter did not answer whether the Metrics View has a time-series dimension; request project YAML/version and a minimal repro. | Close as `not planned` after 14 days without information; reopen on a usable repro. | `Type:Bug`, `Area:Time`, `Needs:Reproduction` (new) |

## Close or consolidate now (38)

Closing these is part of the Now plan. Do not leave completed or context-free issues open merely because a final manual check was never performed. For completion candidates, run the stated smoke test, link the delivering PR/current behavior, then close.

| Issue | Evidence and recommended action | Interim labels / close reason |
|---|---|---|
| [#3764 — Launch browser in background](https://github.com/rilldata/rill/issues/3764) | The issue comments identify the existing `rill start --no-open` behavior and no unmet requirement was supplied. Close as answered; open a new preference-setting issue only with fresh demand. | `Type:Question`, `Area:CLI`; close `completed` |
| [#3803 — Remove profiling queries from dashboard](https://github.com/rilldata/rill/issues/3803) | Direct fixes [PR #3813](https://github.com/rilldata/rill/pull/3813) and [PR #3814](https://github.com/rilldata/rill/pull/3814) merged. Confirm no dashboard caller remains, then close. | `Type:Improvement`, `Area:Dashboard`, `Needs:Verification` (new); close `completed` |
| [#3906 — Welcome-page line animation](https://github.com/rilldata/rill/issues/3906) | Cosmetic 2024 request with no product signal. Close as not planned; a future visual refresh can carry its own current design. | `Type:Improvement`, `nit`, `design`, `wontfix`; close `not planned` |
| [#4044 — Apply filter states across dashboards](https://github.com/rilldata/rill/issues/4044) | Same outcome as #4421. Move any unique context to #4421 and close this duplicate. | `Type:Feature`, `Area:Dashboard`, `customer`; close `duplicate` of #4421 |
| [#4308 — Revamp filter state management](https://github.com/rilldata/rill/issues/4308) | Internal task tied to the 2024 state model and closed prototype #4023. Current implementation has moved on; close superseded and file only concrete residual defects. | `Type:Improvement`, `Area:Dashboard`, `tech debt`; close `not planned/superseded` |
| [#4309 — Revamp time-control state management](https://github.com/rilldata/rill/issues/4309) | Same obsolete implementation-plan problem as #4308; current time controls and Rill Time have been substantially rewritten. | `Type:Improvement`, `Area:Time`, `tech debt`; close `not planned/superseded` |
| [#4310 — Move measure filters to backend](https://github.com/rilldata/rill/issues/4310) | References old issue #3980 and has no standalone acceptance criteria. Close the implementation prescription; retain any current user-facing failure as a new issue. | `Type:Improvement`, `Area:Dashboard`, `Team:Platform`, `tech debt`; close `not planned/superseded` |
| [#4404 — Plot two metrics on one chart](https://github.com/rilldata/rill/issues/4404) | Current Canvas chart code supports multi-metric chart variants. Smoke-test the originally requested two-series case and close completed. | `Type:Feature`, `Area:Dashboard`, `Needs:Verification` (new); close `completed` |
| [#4439 — Finish menu replacement](https://github.com/rilldata/rill/issues/4439) | Broad internal migration task from 2024 with no body or remaining inventory. Close; file component-specific gaps discovered in current code. | `Type:Improvement`, `Team:Applications`, `tech debt`; close `not planned/superseded` |
| [#4440 — Refactor MeasureChart](https://github.com/rilldata/rill/issues/4440) | The exact refactor landed in [PR #8752](https://github.com/rilldata/rill/pull/8752). Close completed. | `Type:Improvement`, `Area:Dashboard`, `tech debt`; close `completed` |
| [#4441 — Component loading states / remove global spinner](https://github.com/rilldata/rill/issues/4441) | Unbounded implementation task with no body. Current components own loading/error states; close the umbrella and open concrete missing-state bugs only. | `Type:Improvement`, `Area:Dashboard`, `tech debt`; close `not planned/superseded` |
| [#4446 — Multiple canonical time fields](https://github.com/rilldata/rill/issues/4446) | The issue comments describe one selectable active time field; alternate time-dimension selection landed in [PR #7848](https://github.com/rilldata/rill/pull/7848). Verify the stated order/shipping-date case and close. | `Type:Feature`, `Area:Time`, `Area:Dashboard`, `Needs:Verification` (new); close `completed` |
| [#4675 — Show invalid theme-color parser error](https://github.com/rilldata/rill/issues/4675) | It was explicitly deferred to common editor tracker #4688, which is closed. Test one invalid color in the current editor and close or rewrite with a current screenshot. | `Type:Bug`, `Area:Developer` (new), `Needs:Verification` (new); close stale tracker |
| [#4817 — Derisk Postgres wire protocol](https://github.com/rilldata/rill/issues/4817) | The linked exploration/implementation [PR #4881](https://github.com/rilldata/rill/pull/4881) merged. The vague “derisk” outcome is complete; any productionization work needs specific issues. | `Type:Improvement`, `Team:Platform`; close `completed` |
| [#4858 — Panning for alternate charts](https://github.com/rilldata/rill/issues/4858) | Direct implementation PRs [#5635](https://github.com/rilldata/rill/pull/5635) and [#5708](https://github.com/rilldata/rill/pull/5708) merged. Close completed. | `Type:Feature`, `Area:Dashboard`; close `completed` |
| [#4877 — Organize Rill Cloud content](https://github.com/rilldata/rill/issues/4877) | Tags now provide the folder-like organization outcome: [PR #9283](https://github.com/rilldata/rill/pull/9283) added tags, dashboard tag rows/sidebar/filtering, and [PR #9672](https://github.com/rilldata/rill/pull/9672) added visual-editor authoring. Confirm tags are the intended replacement and close; file a new hierarchical-folder issue only if that residual is required. | `Type:Feature`, `Area:Cloud` (new), `customer`, `Needs:Verification` (new); close `completed/superseded` |
| [#4887 — Create dev partition during source generation](https://github.com/rilldata/rill/issues/4887) | Two-year-old Notion-only issue with no public acceptance criteria. Close and re-open only as a current, self-contained development-scale proposal. | `Type:Feature`, `Area:Developer` (new), `Needs:Scoping`; close `not planned` |
| [#4944 — Source Inspector reconciling spinner](https://github.com/rilldata/rill/issues/4944) | Direct fix [PR #4953](https://github.com/rilldata/rill/pull/4953) merged. Smoke-test once, then close completed. | `Type:Bug`, `Area:Developer` (new), `Needs:Verification` (new); close `completed` |
| [#5055 — Research form-generation libraries](https://github.com/rilldata/rill/issues/5055) | A research spike is not an enduring backlog item; current JSON-schema and Superforms choices already exist. Close with any useful conclusion captured in docs. | `Type:Question`, `Area:Developer` (new), `tech debt`; close `completed/not planned` |
| [#5060 — Order filters by addition rather than alphabetically](https://github.com/rilldata/rill/issues/5060) | No body, acceptance criteria, or activity since 2024. Close; reopen with a current UX example if users still encounter it. | `Type:Improvement`, `Area:Dashboard`, `Needs:Scoping`; close `not planned` |
| [#5076 — Do not cache connection errors indefinitely](https://github.com/rilldata/rill/issues/5076) | Fixed by [PR #6587](https://github.com/rilldata/rill/pull/6587) (10-second error TTL) and [PR #6595](https://github.com/rilldata/rill/pull/6595) (safe reopen). Current `conncache` config sets `ErrTTL: 10s`. | `Type:Bug`, `Area:Connectors`, `Team:Platform`; close `completed` |
| [#5173 — Extensionless Parquet files fail ingestion](https://github.com/rilldata/rill/issues/5173) | Maintainer supplied a working explicit `read_parquet` source. Close as answered; if auto-detection/hints are desired, rewrite as a feature linked to #3784. | `Type:Question`, `Area:Connectors`; close `completed/answered` |
| [#5202 — Allow `now` as time-range endpoint](https://github.com/rilldata/rill/issues/5202) | Current parser/tests accept expressions such as `-90d to now`, `-1Y to now`, and fixed dates to `now`. Close completed. | `Type:Feature`, `Area:Time`; close `completed` |
| [#5229 — Migrate to Superforms](https://github.com/rilldata/rill/issues/5229) | Current application code uses `sveltekit-superforms`; the old `svelte-forms-lib` dependency is absent. Close completed. | `Type:Improvement`, `Team:Applications`, `tech debt`; close `completed` |
| [#5269 — Filter directly from user input](https://github.com/rilldata/rill/issues/5269) | [PR #6773](https://github.com/rilldata/rill/pull/6773) added tested `Contains` and arbitrary `In List` modes, which apply typed values without waiting for suggestion completion. Close the requested outcome; #5081/#5082 retain the distinct case/prefix semantics. | `Type:Feature`, `Area:Dashboard`, `Area:Performance`; close `completed/superseded` |
| [#5326 — Remember open File Explorer directories](https://github.com/rilldata/rill/issues/5326) | [PR #9682](https://github.com/rilldata/rill/pull/9682), merged within the last two weeks, persists expand/collapse state in local storage with a project-scoped key. Close completed. | `Type:Improvement`, `Area:Developer` (new); close `completed` |
| [#5469 — Use a configurable ClickHouse database for models](https://github.com/rilldata/rill/issues/5469) | [PR #5829](https://github.com/rilldata/rill/pull/5829) added the `database` connector property, and current ClickHouse configuration applies it to the connection auth/database. Close completed. | `Type:Feature`, `Area:ClickHouse`, `Area:Connectors`; close `completed` |
| [#5493 — Improve post-deploy navigation](https://github.com/rilldata/rill/issues/5493) | [PR #7826](https://github.com/rilldata/rill/pull/7826) now lands deploys on a loading page and redirects to the first or originating dashboard; [PR #9435](https://github.com/rilldata/rill/pull/9435) fixed the remaining parser-reconciliation race. Close completed. | `Type:Improvement`, `Team:Applications`, `Area:Cloud` (new); close `completed` |
| [#5544 — ClickHouse `unnest` SQL compilation error](https://github.com/rilldata/rill/issues/5544) | The empty 2024 issue is superseded by repeated ClickHouse unnest implementations/fixes in PRs [#5899](https://github.com/rilldata/rill/pull/5899), [#6184](https://github.com/rilldata/rill/pull/6184), [#7531](https://github.com/rilldata/rill/pull/7531), and [#9582](https://github.com/rilldata/rill/pull/9582). Current code emits `arrayJoin` and has ClickHouse filter tests. Smoke-test one array dimension, then close. | `Type:Bug`, `Area:ClickHouse`, `Needs:Verification` (new); close `completed/superseded` |
| [#5564 — Formatting changes across dashboard](https://github.com/rilldata/rill/issues/5564) | PRs [#5766](https://github.com/rilldata/rill/pull/5766) and [#5832](https://github.com/rilldata/rill/pull/5832) delivered checked items. Create focused issues for any three still-relevant unchecked items, then close this mixed umbrella. | `Type:Improvement`, `Area:Dashboard`, `Needs:Verification` (new); close `completed/split` |
| [#5575 — Canvas dashboard width](https://github.com/rilldata/rill/issues/5575) | Direct fix [PR #5657](https://github.com/rilldata/rill/pull/5657) merged, and current Canvas uses a configurable maximum width. Close completed. | `Type:Feature`, `Area:Dashboard`; close `completed` |
| [#5584 — Canvas UXQA](https://github.com/rilldata/rill/issues/5584) | Notion-only QA umbrella from 2024 with no public checklist. Close; retain each current defect as a reproducible issue. | `Type:Question`, `Area:Dashboard`, `design`; close `not planned/superseded` |
| [#5608 — Consistent global Developer header](https://github.com/rilldata/rill/issues/5608) | Broad, design-dependent 2024 task. Close and refile only the current inconsistencies with screenshots and responsive acceptance criteria. | `Type:Improvement`, `Area:Developer` (new), `design`, `Needs:Scoping`; close `not planned/superseded` |
| [#5645 — Storage usage on project status](https://github.com/rilldata/rill/issues/5645) | [PR #8604](https://github.com/rilldata/rill/pull/8604) added per-model/table database sizes on Status > Tables; [PR #9203](https://github.com/rilldata/rill/pull/9203) added per-connector estimates and [PR #9080](https://github.com/rilldata/rill/pull/9080) added project usage. Confirm the original “component” terminology, then close completed. | `Type:Feature`, `Area:Cloud` (new), `Team:Infra`, `Needs:Verification` (new); close `completed` |
| [#5892 — Remove component data resolvers](https://github.com/rilldata/rill/issues/5892) | Linked PR #5978 closed unmerged; do not call it delivered. The implementation context has aged, so close this prescription and open a current tech-debt delta if it still matters. | `Type:Improvement`, `Team:Platform`, `tech debt`; close `not planned/superseded` |
| [#5905 — Account for OLAP engine during initialization](https://github.com/rilldata/rill/issues/5905) | Linked onboarding PR #6401 closed unmerged and the flow has since evolved. Close the old spec; #6399 is the better current ClickHouse-form tracker. | `Type:Improvement`, `Area:Connectors`, `Area:Developer` (new); close `not planned/superseded` |
| [#5985 — Search an OLAP engine information schema](https://github.com/rilldata/rill/issues/5985) | Current Status > Tables renders a debounced Search input and sends `%term%` through the platform `searchPattern` parameter. The UI arrived in [PR #8864](https://github.com/rilldata/rill/pull/8864) on top of [PR #5980](https://github.com/rilldata/rill/pull/5980). Smoke-test against ClickHouse and close. | `Type:Improvement`, `Area:Connectors`, `Needs:Verification` (new); close `completed` |
| [#7107 — Unify/simplify tooltips](https://github.com/rilldata/rill/issues/7107) | No body, affected surfaces, problem statement, or acceptance criteria. Close and require a scoped replacement. | `Type:Improvement`, `Area:Dashboard`, `Needs:Scoping`; close `not planned` |

## Proposed label system

The repository already has usable priority, size, team, and several area labels. The main problem is inconsistent application, not the absence of a taxonomy.

### Normalize existing labels

Apply exactly one canonical type label to every retained issue:

| Retire or migrate | Canonical target |
|---|---|
| `Bug` | `Type:Bug` |
| `enhancement` | `Type:Feature` for a new capability; `Type:Improvement` for refinement |
| `feature request` | `Type:Feature` |
| `Application` | `Team:Applications` |
| `Platform` | `Team:Platform` |
| `Custom Dashboards` | `Area:Dashboard` |

Keep `customer`, `community`, and `from-slack` as source/impact signals, but never use them as substitutes for priority. Keep `nit`, `tech debt`, `tests`, `design`, and `infrastructure` as work-characteristic labels.

Use the existing labels consistently:

- One of `Priority:P0`, `Priority:P1`, or `Priority:P2` only for work that is actually being considered for the active/next queue.
- One of `Size:S`, `Size:M`, `Size:L`, or `Size:XL` after a technical owner agrees with the estimate.
- One primary `Team:*` label and one or more `Area:*` labels.
- `Needs:Scoping` when the desired outcome is known but the implementation/product contract is not.
- `Needs:Prioritization` for valid Later items without a delivery commitment.

### Add six missing labels

| New label | Meaning |
|---|---|
| `Needs:Reproduction` | A bug cannot enter an engineering queue until reproduced on a supported/current version. |
| `Needs:Verification` | There is evidence the work was delivered or superseded; one smoke test and closure decision remain. |
| `Area:Security` | Authentication, authorization, row/field access policies, and policy-aware UI behavior. |
| `Area:Developer` | Local Developer UI, editor, File Explorer, preview, and authoring workflows. |
| `Area:Cloud` | Rill Cloud project/org management, deploy, hibernation, and shared administration UX. |
| `Area:Metrics` | Metrics View schema, metrics SQL, semantic definitions, and metric metadata. |

Do not add generic `stale`, `duplicate`, or `done` labels. Age is queryable, duplicate/completed are GitHub close reasons, and permanent status labels quickly become another cleanup burden.

## Recommended execution sequence

### First 48 hours

1. Assign and acknowledge contributors on #9395 and #8850.
2. Run the reproduction gates for #5318, #3835, #5061, #5430, and #6350.
3. Close the strongest code-backed completed items first: #3803, #4440, #4817, #4858, #4944, #5076, #5202, #5229, #5269, #5326, #5469, #5493, #5575, #5645, and #5985. Product/smoke-check #4877 and #5544 before closing those two.
4. Consolidate #4044 into #4421; keep #5081/#5082 as the remaining case/prefix search design after closing #5269.

### Current cycle

1. Fix #9395 and #8850 with regression tests.
2. Pull in any reproduced P1 from the gate above, but cap concurrent P1 work to available owners.
3. Land #5988 alongside the policy work.
4. Use remaining capacity for 3–5 items from the 11-item small queue, selected by component locality.

### Following cycles

Pull from Next in this order unless customer commitments change it:

1. #9130 and #6399 — current, owned, and clearly scoped.
2. #4898 and #8085 — finish a partial error-state initiative and close a templating parity gap.
3. #5081/#5082 and #5455 — search performance/semantics and resource naming correctness.
4. #5547 and #8768 — customer-facing temporal/export fidelity.
5. The remaining Next items after a team owner validates size and acceptance criteria.

Later issues should remain unmilestoned until a product owner supplies a problem statement, success metric, and named discovery owner. This is especially important for connectors, forecasting, custom roles, cross-dashboard state, and storage architecture.

## Caveats

- This ordering uses public GitHub evidence. Private Slack, Notion, support severity, revenue impact, and Linear commitments were not available and may legitimately change the order.
- Staleness was used as a confidence signal, not as a proxy for low impact. Old correctness and access-policy defects were elevated when their stated impact warranted it.
- A linked closed PR was counted as delivery evidence only when it was merged or the behavior is present on `main`. Closed, unmerged PRs #5978, #6400, #6401, and #8910 were explicitly not treated as completed work.
- Code inspection can prove presence or absence of implementation paths, but not deployed configuration or browser/runtime behavior. Items marked `Needs:Verification` still require the stated smoke test before closure.
- Estimates should be replaced after reproduction/design review; they are sufficient for triage labels, not sprint promises.
