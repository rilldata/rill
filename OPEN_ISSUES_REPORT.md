# Rill Open Issues Report — Triage, Phasing & Labels

**Repo:** `rilldata/rill` · **Open issues:** 121 · **Generated:** 2026-07-26 · **Updated:** 2026-07-31
**Code verification:** ~45 of the most-suspect issues were checked against the current codebase. Verdicts are marked 🟡 PARTIAL or ❌ still open; the 12 verified-done issues have been removed from the phases and are listed once in Phase 0 for GitHub closure.

## TL;DR

- Only 8 issues were filed in 2026; the rest are a 2024–2025 backlog, much of it unlabeled and with zero comments.
- **12 issues are already implemented** (verified against code) — they are excluded from the work phases and only need closing on GitHub; 11 more are partially done and rescoped to their remaining gap. See "Phase 0" below.
- Recommended order of work: **(0)** close verified-done issues and re-triage stale ones, **(1)** data-correctness and security bugs, **(2)** customer-requested Explore/filter/time features, **(3)** ClickHouse & connectors, **(4)** Rill Developer UX bugs and quick wins, **(5)** refactors, cloud features and big bets.
- The label taxonomy is inconsistent (`Type:Bug` vs `enhancement`, one-off `Area:ClickHouse`, `nit`, `from-slack`). A normalized `type:` / `area:` / `priority:` scheme is proposed at the end, with per-issue suggestions in every table.

---

## Phase 0 — Close and re-triage (near-zero engineering cost)

### ✅ Already implemented — close on GitHub (removed from the phases below)

These 12 issues were verified as already implemented in current code and are excluded from all work phases. They are still open on GitHub (checked 2026-07-31) and just need closing:

`#3764` (`--no-open` flag) · `#5202` ("now" in rilltime) · `#5061` (fixed by the rilltime replacement) · `#5775` (TDD chart resizer) · `#5575` (canvas `max_width`) · `#4007` (derived metrics views via `parent_*`) · `#4446` (multiple time dimensions + picker) · `#5469` (ClickHouse `database` config) · `#5645` (`ProjectStorage` on status page) · `#4675` (theme parse errors surfaced) · `#4821` (ingesting spinner in inspector) · `#5905` (OLAP-aware project init)

### 🟡 Partially implemented — rescope to the remaining gap before working on them

| # | What's done | What remains |
|---|---|---|
| [#6357](https://github.com/rilldata/rill/issues/6357) Custom time ranges | Rich rilltime grammar; explore/canvas `time_ranges` validated through it | No `label`/`display_name` on `ExploreTimeRange` — custom ranges can't be given friendly names |
| [#5026](https://github.com/rilldata/rill/issues/5026) Auto-populate dimensions | Explore YAML supports `dimensions: '*'` (and defaults to all) | No wildcard at the metrics-view level; "all columns" only via non-AI generator fallback |
| [#4898](https://github.com/rilldata/rill/issues/4898) Inline 4xx errors | `InlineErrorIndicator.svelte` on measure charts and big numbers | Leaderboards render no error at all; TDD shows a generic message without the server error |
| [#5564](https://github.com/rilldata/rill/issues/5564) Dashboard formatting | Data-driven column widths (leaderboard + pivot), d3 locale/currency, tooltip formatters | Generic table formatting rules; resize-driven dynamic formatting |
| [#5647](https://github.com/rilldata/rill/issues/5647) Sharing thumbnail | Org-level thumbnail upload (`ThumbnailSettings.svelte`, `thumbnail_asset_id`) | `og:image` still hardcoded in `web-admin/src/app.html`; no per-dashboard thumbnail |
| [#5699](https://github.com/rilldata/rill/issues/5699) `metrics_sql` aliases | Aliases honored for function-call expressions | `parseSelect` still has `TODO no alias handling for column names`; filter parser rejects aliases |
| [#5975](https://github.com/rilldata/rill/issues/5975) Idle detection | Per-tab SSE pause on blur/visibility with idle timeouts | No cross-tab "keep 2 most-recent tabs active" coordination (no BroadcastChannel/tab registry) |
| [#4838](https://github.com/rilldata/rill/issues/4838) File-explorer icons | Special icons for `.env` and `rill.yaml`; resource-kind icons | `.py`/`.json`/`.md` etc. still fall through to the generic file icon |
| [#5713](https://github.com/rilldata/rill/issues/5713) New-file positional context | "Add folder"/"Add blank file" respect the current directory | Typed resources (canvas, metrics view, API…) still land in fixed top-level folders (`new-files.ts:119-137`) |
| [#5985](https://github.com/rilldata/rill/issues/5985) Info-schema search | `search_pattern` API + search box on the local Tables page | No search input in the connector explorer sidebar itself |
| [#6399](https://github.com/rilldata/rill/issues/6399) ClickHouse source forms | Add-data flow emits `output.connector: clickhouse` based on project OLAP | No ClickHouse-specific source-step form fields (`templates/schemas/clickhouse.ts` is connector-step only) |

### Stale-triage sweep

Roughly a third of the backlog is from H1 2024, unlabeled, with zero comments. Several likely no longer apply after the explore/canvas rewrites — the 12 verified-done issues above prove drift is real. Re-confirm or close:

`#3760 #3784 #3792 #3803 #3820 #3854 #3896 #3906 #4308 #4309 #4310 #4439 #4440 #4441 #4445 #4579 #4887 #5055 #5229 #5330 #5584 #5608 #5633 #5634 #5892 #5958`

Suggested label: `stale-candidate` (or a stale-bot with a 90-day window going forward).

---

## Phase 1 — Correctness & security bugs (work on these first)

These produce **wrong numbers, broken dashboards, or broken access control**, which is existential for a BI tool.

| # | Title (abridged) | Verified status / why now | Suggested labels |
|---|---|---|---|
| [#9395](https://github.com/rilldata/rill/issues/9395) | Pivot "Absolute change" ignores new dimension members, shows 0 change | Wrong numbers in time comparison; recent (May 2026) | `type:bug` `area:pivot` `priority:p0` |
| [#3835](https://github.com/rilldata/rill/issues/3835) | An `include` access policy breaks the dashboard | Security policies failing is a trust issue | `type:bug` `area:security-policies` `priority:p0` |
| [#8573](https://github.com/rilldata/rill/issues/8573) | "Invalid interval UTC" when selecting time interval | ❌ Confirmed still broken: fallback text still in `RangeDisplay.svelte`. Note: a suppression-only fix exists on unmerged branch `origin/cursor/APP-761-…` (`97044728f`) — the underlying resolution failure in `new-time-controls.ts:363-403` is unfixed | `type:bug` `area:time-controls` `priority:p1` |
| [#4001](https://github.com/rilldata/rill/issues/4001) | Data column missing in TDD table | ❌ Confirmed, root cause found: table strips boundary buckets the chart renders (`time-dimension-data-store.ts:222` — `data?.slice(1, -1)`) | `type:bug` `area:explore` `priority:p1` |
| [#5076](https://github.com/rilldata/rill/issues/5076) | Runtime: don't cache connection errors indefinitely | 7 comments; projects stuck in error until restart | `type:bug` `area:runtime` `priority:p1` |
| [#5799](https://github.com/rilldata/rill/issues/5799) | `metrics_sql` `date_trunc` day precision not formatted to day | Wrong output from the metrics SQL layer | `type:bug` `area:metrics-layer` `priority:p1` |
| [#4638](https://github.com/rilldata/rill/issues/4638) | Pivot: one row dim + one column dim shows an empty table | Core pivot layout broken | `type:bug` `area:pivot` `priority:p1` |
| [#6350](https://github.com/rilldata/rill/issues/6350) | Waking a hibernated project: settings navigation stops the wake in UI | Cloud onboarding/resume friction | `type:bug` `area:cloud` `priority:p2` |

## Phase 2 — Explore dashboard & filter/time features (biggest customer cluster)

The largest coherent cluster of `customer`-tagged asks. Clubbing them lets one engineer own filters + time controls end to end. All verified ❌ still open unless noted.

### 2a. Filters (club together)

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#5040](https://github.com/rilldata/rill/issues/5040) | Include **and** exclude filter on the same dimension | ❌ One expression per dimension (`dimension-filters.ts`); read-only "Advanced" chip can display but not build it | `type:feature` `area:filters` `customer` |
| [#4044](https://github.com/rilldata/rill/issues/4044) | Define and apply filter states across dashboards | — | `type:feature` `area:filters` `customer` |
| [#4421](https://github.com/rilldata/rill/issues/4421) | Cross-dashboard filters (overlaps #4044 — consider merging) | — | `type:feature` `area:filters` `customer` `duplicate?` |
| [#5060](https://github.com/rilldata/rill/issues/5060) | Order filters by order added, not alphabetically | ❌ Still `a.name > b.name` sort (`selectors/filters.ts`) | `type:feature` `area:filters` `good-first-issue` |
| [#5428](https://github.com/rilldata/rill/issues/5428) | Ability to unset column sorting | ❌ `enableSortingRemoval: false` in all four table components | `type:feature` `area:explore` `good-first-issue` |

### 2b. Time ranges & grains (club together)

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#6357](https://github.com/rilldata/rill/issues/6357) | Custom time-range **labels** (rescoped; grammar itself is done) | 🟡 | `type:feature` `area:time-controls` |
| [#4366](https://github.com/rilldata/rill/issues/4366) | Custom time-grain formatting | ❌ Grain labels hardcoded in `web-common/src/lib/time/config.ts` | `type:feature` `area:time-controls` |
| [#3958](https://github.com/rilldata/rill/issues/3958) | Grey out unavailable time grains rather than hiding them | ❌ Selectors still filter grains out (`TimeGrainSelector.svelte`, `new-grains.ts:219-239`) | `type:feature` `area:time-controls` `customer` `good-first-issue` |
| [#8768](https://github.com/rilldata/rill/issues/8768) | Support ISO identifiers while retaining temporal type | — | `type:feature` `area:time-controls` `from-slack` |

### 2c. Leaderboard, TDD & pivot enhancements

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#4405](https://github.com/rilldata/rill/issues/4405) | Leaderboard comparison: previous value as a column | ❌ Still hover-only (`LeaderboardRow.svelte:218`) | `type:feature` `area:explore` |
| [#5547](https://github.com/rilldata/rill/issues/5547) | Pivot exports should respect selected granularity | ❌ Server writes full timestamps regardless of grain (`runtime/queries/metricsview.go:345`) | `type:feature` `area:pivot` `customer` |
| [#5624](https://github.com/rilldata/rill/issues/5624) | Show comparison dates in pivot headers | ❌ Headers still generic "Δ"/"Δ %" (`pivot-column-definition.ts:218-228`) | `type:feature` `area:pivot` |
| [#5612](https://github.com/rilldata/rill/issues/5612) | Raw data viewer for TDD | ❌ Rows viewer explicitly excluded in TDD mode (`Dashboard.svelte:285`) | `type:feature` `area:explore` |
| [#4898](https://github.com/rilldata/rill/issues/4898) | Inline 4xx errors on leaderboards + TDD (rescoped) | 🟡 | `type:feature` `area:explore` |
| [#5648](https://github.com/rilldata/rill/issues/5648) | Additional Time Dimension graph capabilities | — | `type:feature` `area:explore` `needs-triage` (vague scope) |

### 2d. Metrics layer (club together — definition reuse & governance theme)

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#8085](https://github.com/rilldata/rill/issues/8085) | Templating of model names/properties in metrics views | ❌ Parser explicitly disallows structural templating (`runtime/parser/template.go:26`) | `type:feature` `area:metrics-layer` |
| [#8711](https://github.com/rilldata/rill/issues/8711) | Native, exportable dictionary of metrics and definitions | — | `type:feature` `area:metrics-layer` |
| [#5026](https://github.com/rilldata/rill/issues/5026) | Wildcard dimensions at metrics-view level (rescoped) | 🟡 | `type:feature` `area:metrics-layer` `customer` |
| [#5699](https://github.com/rilldata/rill/issues/5699) | Column-name aliases in `metrics_sql` (rescoped) | 🟡 | `type:feature` `area:metrics-layer` `good-first-issue` |

## Phase 3 — ClickHouse & connectors

### 3a. ClickHouse (club together)

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#5199](https://github.com/rilldata/rill/issues/5199) | Better error handling for ClickHouse OLAP engine | — | `type:feature` `area:clickhouse` |
| [#5544](https://github.com/rilldata/rill/issues/5544) | `unnest` throws SQL compilation error on ClickHouse | — | `type:bug` `area:clickhouse` |
| [#5318](https://github.com/rilldata/rill/issues/5318) | Debug query cancellation in `clickhouse-go` | — | `type:bug` `area:clickhouse` |
| [#5081](https://github.com/rilldata/rill/issues/5081) / [#5082](https://github.com/rilldata/rill/issues/5082) | LIKE vs ILIKE + prefix-LIKE for suggestions | ❌ Search hardcodes `ILIKE '%x%'` (`executor.go:994-1021`); no toggle exists. Club as one "configurable dimension search matching" issue with [#5269](https://github.com/rilldata/rill/issues/5269) | `type:feature` `area:clickhouse` |
| [#6399](https://github.com/rilldata/rill/issues/6399) | ClickHouse-specific source form fields (rescoped) | 🟡 | `type:feature` `area:clickhouse` `area:dev-ux` |

### 3b. Connectors & sources

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#5798](https://github.com/rilldata/rill/issues/5798) | Trino connector support (5 comments — community interest) | — | `type:feature` `area:connectors` |
| [#8911](https://github.com/rilldata/rill/issues/8911) | Arrow Flight SQL transport for StarRocks connector | — | `type:feature` `area:connectors` |
| [#5101](https://github.com/rilldata/rill/issues/5101) | MongoDB connector | — | `type:feature` `area:connectors` `customer` |
| [#5173](https://github.com/rilldata/rill/issues/5173) | Parquet files without type extension fail to ingest | — | `type:bug` `area:connectors` |
| [#5430](https://github.com/rilldata/rill/issues/5430) | VARCHAR → LIST conversion error on ingest | — | `type:bug` `area:connectors` |
| [#4260](https://github.com/rilldata/rill/issues/4260) | `limit` source property | ❌ No dedicated `limit`, but `dev:`/`environment_overrides:` cover the use case — consider closing as "workaround supported" | `type:feature` `area:connectors` `customer` |
| [#4579](https://github.com/rilldata/rill/issues/4579) | Schema relaxation: cast numbers to strings when needed | — | `type:feature` `area:connectors` |
| [#4895](https://github.com/rilldata/rill/issues/4895) | Noisy `gosnowflake` DBUS warning | — | `type:bug` `area:connectors` `good-first-issue` |
| [#3784](https://github.com/rilldata/rill/issues/3784) | Warn when Add Source URI lacks a file format | — | `type:feature` `area:connectors` `good-first-issue` |
| [#4887](https://github.com/rilldata/rill/issues/4887) | Create a dev partition when generating initial source spec | — | `type:feature` `area:connectors` |
| [#5985](https://github.com/rilldata/rill/issues/5985) | Search in connector explorer sidebar (rescoped; API + Tables page done) | 🟡 | `type:feature` `area:connectors` `area:dev-ux` |

## Phase 4 — Rill Developer UX: bugs first, then quick wins

Individually small; batching them gives an outsized "quality" release. Several are `good-first-issue` material.

### 4a. Developer bugs

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#4917](https://github.com/rilldata/rill/issues/4917) | SQL autocomplete fails when table name is quoted | — | `type:bug` `area:editor` |
| [#4933](https://github.com/rilldata/rill/issues/4933) | Duplicate column missing from `rows` API response | — | `type:bug` `area:editor` `area:runtime` |
| [#3807](https://github.com/rilldata/rill/issues/3807) | Model results sometimes don't show when typing slowly | — | `type:bug` `area:editor` |
| [#3949](https://github.com/rilldata/rill/issues/3949) | "X of Y rows" doesn't update Y when switching models | — | `type:bug` `area:editor` `good-first-issue` |
| [#5379](https://github.com/rilldata/rill/issues/5379) | 404 flash when a file is dropped onto Rill Developer | — | `type:bug` `area:dev-ux` |
| [#5455](https://github.com/rilldata/rill/issues/5455) | Same-name/different-case file collision not caught | — | `type:bug` `area:parser` |
| [#4944](https://github.com/rilldata/rill/issues/4944) | Source inspector sometimes missing reconciling spinner | — (base feature exists per #4821; this is a flicker/timing bug) | `type:bug` `area:dev-ux` |
| [#5681](https://github.com/rilldata/rill/issues/5681) | File header jumps after zooming timestamp chart + navigating | — | `type:bug` `area:dev-ux` |
| [#5958](https://github.com/rilldata/rill/issues/5958) | Welcome screen: example project click sometimes hangs | — | `type:bug` `area:dev-ux` |
| [#3960](https://github.com/rilldata/rill/issues/3960) | Hiding a measure/dimension cancels outstanding queries | — | `type:bug` `area:explore` |
| [#4213](https://github.com/rilldata/rill/issues/4213) | Leaderboard filter often shows a cancelled query | ❌ No cancellation handling in `Leaderboard.svelte:175-265`; still plausible | `type:bug` `area:explore` |
| [#4258](https://github.com/rilldata/rill/issues/4258) | Misaligned columns for measures without time series | — | `type:bug` `area:explore` |
| [#5601](https://github.com/rilldata/rill/issues/5601) | Clicking pivot cell doesn't expand model viewer | — | `type:bug` `area:pivot` |

### 4b. Quick wins & polish

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#5326](https://github.com/rilldata/rill/issues/5326) | File explorer: remember open directories | ❌ In-memory store only; connector explorer already shows the localStorage pattern to copy (`connector-explorer-store.ts:65-67`) | `type:feature` `area:dev-ux` `good-first-issue` |
| [#5713](https://github.com/rilldata/rill/issues/5713) | Typed resources should respect current directory (rescoped) | 🟡 | `type:feature` `area:dev-ux` `good-first-issue` |
| [#4838](https://github.com/rilldata/rill/issues/4838) | Extension-specific icons for `.py`/`.json`/`.md`… (rescoped) | 🟡 | `type:feature` `area:dev-ux` `good-first-issue` |
| [#4828](https://github.com/rilldata/rill/issues/4828) | Close the application from the Developer UI | ❌ No shutdown RPC in `proto/rill/local/v1/api.proto` | `type:feature` `area:dev-ux` `community` `good-first-issue` |
| [#5588](https://github.com/rilldata/rill/issues/5588) | Table preview: show underlying datatypes | ❌ `ProfileColumn` has a single `type`; non-DuckDB OLAPs return runtime codes (`table_columns.go:136-141`) | `type:chore` `area:dev-ux` |
| [#5970](https://github.com/rilldata/rill/issues/5970) | Generated metrics YAML should match Visual Metrics Editor layout | ❌ Generator and editor still emit different key orders | `type:chore` `area:dev-ux` `good-first-issue` |
| [#5975](https://github.com/rilldata/rill/issues/5975) | Cross-tab idle coordination (rescoped) | 🟡 | `type:feature` `area:dev-ux` |
| [#5495](https://github.com/rilldata/rill/issues/5495) | Support creating mapping tables in Developer | ❌ No UI table-authoring path exists | `type:feature` `area:dev-ux` |
| [#4871](https://github.com/rilldata/rill/issues/4871) | Comparison selector color selection | — | `type:chore` `area:explore` `good-first-issue` |
| [#3906](https://github.com/rilldata/rill/issues/3906) | Background line CSS animation on welcome page | — | `type:chore` `area:dev-ux` `good-first-issue` |

## Phase 5 — Refactors, cloud features & big bets (schedule deliberately)

### 5a. Frontend tech debt & state management (re-validate before starting — most predate current architecture)

| # | Title | Suggested labels |
|---|---|---|
| [#4308](https://github.com/rilldata/rill/issues/4308) / [#4309](https://github.com/rilldata/rill/issues/4309) / [#4310](https://github.com/rilldata/rill/issues/4310) | State management revamp: filters, time controls, measure filters to backend | `type:chore` `area:frontend` `needs-triage` |
| [#4439](https://github.com/rilldata/rill/issues/4439) / [#4440](https://github.com/rilldata/rill/issues/4440) / [#4441](https://github.com/rilldata/rill/issues/4441) / [#4445](https://github.com/rilldata/rill/issues/4445) | Menu replacement, MeasureChart refactor, loading states, scrollbar styling | `type:chore` `area:frontend` `stale-candidate` |
| [#7107](https://github.com/rilldata/rill/issues/7107) | Unify/simplify tooltips | `type:chore` `area:frontend` |
| [#5229](https://github.com/rilldata/rill/issues/5229) / [#5055](https://github.com/rilldata/rill/issues/5055) | Superforms migration / form-library research | `type:chore` `area:frontend` `stale-candidate` |
| [#5330](https://github.com/rilldata/rill/issues/5330) | Dimension detail: reuse the pivot table component | `type:chore` `area:frontend` |
| [#5608](https://github.com/rilldata/rill/issues/5608) | Consistent global header in Rill Developer | `type:chore` `area:dev-ux` |
| [#3803](https://github.com/rilldata/rill/issues/3803) | Remove profiling queries from dashboard component | `type:chore` `area:frontend` |
| [#5633](https://github.com/rilldata/rill/issues/5633) / [#5634](https://github.com/rilldata/rill/issues/5634) | Chart brush / stacked-area brush styling | `type:chore` `area:charts` |
| [#5584](https://github.com/rilldata/rill/issues/5584) | Canvas UXQA | `type:chore` `area:canvas` `needs-triage` (likely partially done) |

### 5b. Runtime & infra

| # | Title | Suggested labels |
|---|---|---|
| [#3854](https://github.com/rilldata/rill/issues/3854) | Persistent storage for distributed workers | `type:feature` `area:runtime` |
| [#3760](https://github.com/rilldata/rill/issues/3760) | Serve from local SSD with persistent GCS backup | `type:feature` `area:runtime` |
| [#3792](https://github.com/rilldata/rill/issues/3792) | Flush catalog when adding a trigger | `type:chore` `area:runtime` `stale-candidate` |
| [#3820](https://github.com/rilldata/rill/issues/3820) | Don't send reports for obviously invalid queries | `type:bug` `area:alerts-reports` |
| [#5892](https://github.com/rilldata/rill/issues/5892) | Remove data resolvers for components | `type:chore` `area:runtime` |
| [#5284](https://github.com/rilldata/rill/issues/5284) | Fix pagination in GitHub List APIs | `type:bug` `area:admin` |
| [#3896](https://github.com/rilldata/rill/issues/3896) | Add dependency validation to CI | `type:chore` `area:ci` |
| [#5988](https://github.com/rilldata/rill/issues/5988) | e2e test for editing security policies | `type:chore` `area:testing` |

### 5c. Rill Cloud features

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#8756](https://github.com/rilldata/rill/issues/8756) | Custom role creation | ❌ Roles hardcoded (`admin/database/database.go:960-1010`); no CreateRole RPC | `type:feature` `area:cloud` `from-slack` |
| [#8850](https://github.com/rilldata/rill/issues/8850) | Filter canvas KPI/Table based on metrics security policy | — | `type:feature` `area:canvas` `area:security-policies` `from-slack` |
| [#5664](https://github.com/rilldata/rill/issues/5664) | Dashboard bookmarks for user groups | ❌ `Bookmark` has only `UserID`/`Default`/`Shared` — no group scope | `type:feature` `area:cloud` |
| [#4877](https://github.com/rilldata/rill/issues/4877) | Organize content in Rill Cloud (folders) | — | `type:feature` `area:cloud` `customer` |
| [#4429](https://github.com/rilldata/rill/issues/4429) | Global asset search | ❌ All web-admin search is scoped; no cross-project RPC | `type:feature` `area:cloud` `customer` |
| [#5647](https://github.com/rilldata/rill/issues/5647) | Per-dashboard sharing thumbnail / og:image (rescoped) | 🟡 | `type:feature` `area:cloud` |
| [#5493](https://github.com/rilldata/rill/issues/5493) | Improve journey from Project Status page to dashboard after deploy | — | `type:feature` `area:cloud` |
| [#5755](https://github.com/rilldata/rill/issues/5755) | Alerts: support dimension criteria | ❌ Criteria limited to measures (`alerts/form-utils.ts:31`, `CriteriaForm.svelte:30-77`) | `type:feature` `area:alerts-reports` |

### 5d. Big bets (need product decision before engineering)

| # | Title | Verified | Suggested labels |
|---|---|---|---|
| [#4404](https://github.com/rilldata/rill/issues/4404) | Plot two metrics on a single chart (6 comments) | — | `type:feature` `area:charts` `epic` |
| [#4863](https://github.com/rilldata/rill/issues/4863) | Forecasting in time-series charts | — | `type:feature` `area:charts` `epic` `customer` |
| [#4817](https://github.com/rilldata/rill/issues/4817) | Postgres wire protocol support | — | `type:feature` `area:runtime` `epic` |
| [#4235](https://github.com/rilldata/rill/issues/4235) | Local LLM support for privacy | — | `type:feature` `area:ai` `epic` |
| [#4407](https://github.com/rilldata/rill/issues/4407) | Locked-down enterprise environment support | — | `type:feature` `area:deployment` `epic` |
| [#4577](https://github.com/rilldata/rill/issues/4577) | Interactive UI for custom API management | ❌ APIs only get the generic code editor; no API workspace | `type:feature` `area:dev-ux` `epic` |
| [#4858](https://github.com/rilldata/rill/issues/4858) | Panning for alternate charts | — | `type:feature` `area:charts` |
| [#9130](https://github.com/rilldata/rill/issues/9130) | Canvas heatmap: `normalize` option for row/column % | — | `type:feature` `area:canvas` |

---

## Proposed label taxonomy

Current labels are inconsistent: `Type:Bug`/`Type:Feature` coexist with `enhancement`, `nit`, `from-slack`, a single `Area:ClickHouse`, and `Team:Applications`. Suggested normalized scheme:

- **Type (exactly one):** `type:bug`, `type:feature`, `type:chore` (replaces `nit`, refactors, tech debt)
- **Area (one or more):** `area:explore`, `area:pivot`, `area:canvas`, `area:time-controls`, `area:filters`, `area:metrics-layer`, `area:charts`, `area:clickhouse`, `area:connectors`, `area:runtime`, `area:cloud`, `area:admin`, `area:dev-ux`, `area:editor`, `area:cli`, `area:parser`, `area:security-policies`, `area:alerts-reports`, `area:ai`, `area:frontend`, `area:testing`, `area:ci`, `area:deployment`
- **Priority:** `priority:p0` (correctness/security), `priority:p1`, `priority:p2`
- **Source (keep, they're useful signal):** `customer`, `community`, `from-slack`
- **Process:** `good-first-issue`, `needs-triage`, `stale-candidate`, `epic`, `duplicate?`

Notable clubbing opportunities: **#4044 + #4421** (cross-dashboard filter state — likely duplicates), **#4308/#4309/#4310** (one state-management epic), **#5081/#5082/#5269** (one "configurable dimension search matching" epic), **#3958/#4366/#6357** (time-grain/range UX polish bundle).
