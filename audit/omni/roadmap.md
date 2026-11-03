# Roadmap: Competing with (and beating) Omni

Companion to `comparison.md` (feature matrix) and the three `omni-features-*.md` research files. Effort sizes: S (≤1 eng-month), M (1–3), L (3–6), XL (6+ across teams).

## 1. Strategic frame

### Where the fight actually is
Omni is "Looker rebuilt for the AI era": a semantic layer over the customer's warehouse, with best-in-class self-service (workbooks, Excel calcs, SQL↔UI) and a productized AI feedback loop. Its structural weaknesses are the inverse of Rill's strengths:

- **Omni has no data plane.** No ingestion, no transformation, no storage. Every query hits the customer's warehouse (or a cache of it). At event scale (adtech, product analytics, observability), warehouse cost and latency are Omni's ceiling and Rill's floor.
- **Omni is SaaS-only and quote-priced.** No local product, no OSS, no transparent pricing, key features contract-gated (embed, BYO LLM, whitelabeling).
- **Omni's git is a sync; Rill's git is the source of truth.** Rill's code-first DX (single binary, files, branch deployments, local dev) is authentically better for engineering-led teams.

**Positioning: "The fastest path from raw events to governed, AI-ready dashboards — at a fraction of warehouse BI cost."** Win operational/embedded/event-scale analytics outright; achieve *sufficient* parity on self-service and enterprise IT so Rill stops losing deals on checklists.

### The three deal-loss categories to eliminate
Everything below maps to one of these:
1. **"Rill can't join my tables."** (semantic layer joins / multi-fact)
2. **"My analysts can't do anything the developer didn't model."** (self-service depth)
3. **"It's missing table stakes."** (viz types, drill-to-records, cross-filtering, delivery destinations, SCIM/audit logs)

### What we deliberately do NOT chase (see §6)
Workbook/spreadsheet document paradigm, in-app Excel engine, dbt-compiler integration, compute routing, 14-warehouse live-query breadth. Rationale in §6.

---

## 2. Phase 1 — Checklist parity sprint (0–3 months)

High-visibility, low-architectural-risk items that remove instant disqualifiers in evals.

### 2.1 Visualization breadth: native chart types (M)
**Omni:** 17 native viz types; maps (point + choropleth + custom GeoJSON), sankey, boxplot are native; waterfall/gauge/sunburst/Gantt via a documented Vega gallery.
**Plan:** Rill charts already render through Vega-Lite (`web-common/src/features/components/charts/`), so each new type is a template + config-panel schema, not a new engine:
- **Maps**: choropleth (country/US-state name or ISO code matching, like Omni's built-in region sets) + point map (lat/lon dims, color/size measures). Ship built-in TopoJSON for world + US states; `geojson:` property for custom regions. New `map_chart` in the chart type union (`charts/types.ts`) + a `MapChart` renderer.
- **Sankey** (dimension → dimension flow with a measure), **waterfall** (measure sequence or dimension contribution), **gauge/bullet** (measure vs target, pairs with KPI), **boxplot**, **histogram** (binned measure — trivial in DuckDB/ClickHouse).
- **Chart polish across all types**: reference lines (`reference_lines:` with value/measure/label), trendlines (moving average + linear regression — computable via existing window-measure machinery, no client math needed), line interpolation/step options, totals labels on stacked bars.
- **Model-defined dimension colors**: `color:` mapping on dimension values in metrics-view YAML (Omni's `colors:` param) so a dimension value renders consistently across every chart. Small parser + theme plumbing.
- **Vega gallery**: publish a docs gallery of copy-paste Vega-Lite specs for the long tail (Omni does this; Rill's AI spec generator makes it even easier — lean into "describe the chart, get the spec").

### 2.2 Drill-to-records (M)
**Omni:** click any measure → underlying rows; curated `drill_fields`/`drill_queries` per measure; right-click cell → drill/filter/copy.
**Plan:** `MetricsViewRows` RPC already exists. Add:
- `drill:` block on metrics-view measures (`fields:` list, optional named drill queries) in `parse_metrics_view.go`.
- Click-through UI in explore (leaderboard row, TDD point, pivot cell) and canvas widgets → filtered row viewer modal, respecting row/column security policies (rows drill must re-apply `SecurityRule` field access).
- Gate with an explore/canvas `drill: false` toggle for embeds.
This closes the single most common "basic BI" objection alongside charts.

### 2.3 Canvas cross-filtering (M)
**Omni:** opt-in click-to-filter across tiles, brush-select ranges, multi-select, filter chips.
**Plan:** Canvas already has a shared filter store (`canvas/filters/`) and per-widget filter overrides. Add an opt-in `cross_filter: true` canvas setting: clicking a dimension value in any chart/table/leaderboard adds a dimension filter chip to the canvas filter bar (reuse existing dimension filter model — chips, removal, URL encoding all exist); brush on time axes → time-range filter. Widgets on a different metrics view apply the filter when a same-named dimension exists (name-match now; join-aware later per §4.1).

### 2.4 Delivery destinations & formats (M)
**Omni:** email, Slack, webhook, SFTP, S3, Google Sheets; PNG/PDF/XLSX/CSV; per-recipient bursting; email-only users; sender whitelabeling.
**Plan:** Rill's notifier abstraction (email + `slack` driver) generalizes:
- New notifier/export targets: **generic webhook** (S), **S3/GCS/Azure blob drop** (S — ObjectStore drivers already exist; reuse them as delivery sinks with templated paths, like Omni's Mustache S3 paths), **SFTP** (S), **Google Sheets** (M — single-query append/replace to a sheet).
- **PNG render of canvas/explore** (S–M: the PDF export pipeline in `web-common/src/features/exports/pdf/` already headless-renders; add image output) — needed for Slack unfurls and email bodies too.
- **Report bursting** (M): reports gain `for_each_recipient: true` — re-evaluate the export per recipient under their user attributes (alerts already do per-user evaluation in `runtime/reconcilers/alert.go`; port that pattern to `report.go`).
- **Custom sender domain** for report/alert email (S, admin config + DKIM docs).

### 2.5 Enterprise IT pack, part 1 (M)
**Omni:** SCIM (Okta/Entra/Rippling), audit-log streams to S3/GCS/Azure with 1yr retention, custom roles.
**Plan:**
- **Audit logs** (M): emit structured events from `admin/server/` (auth, membership, token, project, magic-token, export events) + runtime query-context events to a per-org stream; delivery = scheduled batch drop to customer S3/GCS/Azure (reuse ObjectStore drivers). Table-stakes for security review checklists.
- **SCIM v2 endpoints** (M) in the admin service (users + groups), mapping to existing user/usergroup tables; document Okta/Entra setup. Also unlocks IdP-synced **user attributes** feeding security policies (Omni's biggest governance multiplier).
- **SAML/OIDC self-serve config UI** (S–M): surface the existing Auth0 enterprise-connection flow in org settings instead of a support ticket.

### 2.6 Metrics-view quality-of-life (S each)
- **Fiscal calendars**: `fiscal_year_offset` on metrics views; fiscal quarter/year grains in rilltime + explore time controls. (Omni: `fiscal_month_offset` + fiscal timeframes.)
- **Auto time dimensions**: derived day-of-week / hour-of-day / month-of-year dimensions from the time dimension (Omni timeframes) without hand-written SQL.
- **Field `synonyms:`** on dimensions/measures — doubles as AI grounding (see §5.2) and search relevance.
- **Column masking mode**: `field_access` gains `mask: hash` (hash instead of hide — Omni's `mask_unless_access_grants`), preserving group-bys on masked PII.

**Phase 1 exit criteria:** Rill passes a standard mid-market BI RFP checklist on visualization, drill, delivery, and IT-security rows.

---

## 3. Phase 2 — Self-service depth (3–6 months)

Address "viewers can only consume what developers modeled" without importing Omni's workbook paradigm.

### 3.1 Ad-hoc calculated measures in Explore/Pivot (L) — the keystone
**Omni:** viewers add SQL custom fields and ~120 Excel-function table calcs; calcs promotable to the shared model.
**Plan (Rill-native shape):**
- **"Add measure" in explore/pivot for viewers**: expression builder over the metrics view's existing measures and dimensions — arithmetic on measures (ratio, difference, % of), window functions (running total, %, rank — the window-measure machinery in `MetricsViewMeasureWindow` already computes these server-side), and period comparisons. Session-scoped, URL-encoded (bookmarkable/shareable via existing bookmark + URL-state infra), evaluated through MetricsViewAggregation so security policies hold automatically.
- **Format-string + quick-calc menu** (% of total, running sum, rank, moving average) as one-click wrappers.
- **Promotion path = "just-in-time modeling, Rill-style"**: a "Save to project" action hands the ad-hoc measure to the AI developer agent, which opens a change (edit-session branch → commit → PR via existing `web-admin/src/features/edit-session/` flow). Omni promotes into a mutable shared model; Rill promotes **through code review** — pitch that as governance, not friction.

### 3.2 Canvas controls & filter UX (M)
**Omni:** field switcher, multi-field picker, timeframe switcher, PoP control, Top-N control; filter display options (buttons/toggles/calendar/preset lists); required/hidden filters; user-attribute defaults; faceted suggestions.
**Plan:** Canvas `variables` plumbing already exists — build the viewer-facing widgets on it:
- **Measure/dimension switcher** component (swaps the field bound into sibling widgets), **time-grain switcher**, **comparison-period picker**, **Top-N selector** for leaderboards/bar charts.
- Filter display options on canvas dimension filters (buttons, multi-select dropdown, curated value list); `required:` and `hidden:` filter flags; default values templated from user attributes (`{{ .user.region }}`).
- Faceted filter suggestions: scope dimension-value suggestion queries by the other active filters (the search/toplist API already accepts a filter set — wire it through).

### 3.3 Canvas layout v2 (M)
**Omni (new editor):** nested grid/stack containers, ≤15 pages with per-page filters, page-nav styles, custom URL tabs, mobile overrides, presentation mode, auto-refresh, placeholder items, spacers/dividers.
**Plan:** Extend canvas YAML: per-tab (existing tab groups) **own filter sets**; **custom URL tabs** (external links in the tab bar, forwarding filter params — Omni's pattern, cheap and loved); **presentation/kiosk mode** with **auto-refresh interval** (wallboards — natural fit for Rill's real-time story, arguably more "Rill" than Omni); spacer/divider components; per-widget hide-on-mobile. Defer free-form nested containers unless design demands it — Rill's row grid is a legitimate simplicity stance.

### 3.4 Embedded analytics: keep the lead, add creation (L)
**Omni:** Create Mode (embed users build dashboards in-iframe, entity folders), embedded deliveries, embedded AI with per-entity credit caps, vanity domains.
**Plan:** Embedded ops analytics is a Rill core market — extend the moat:
- **Embed "save view" / personal canvases**: `external_user_id` users can save bookmarks (exists for state) and **compose canvases from existing metrics views** in a scoped workspace — a deliberately smaller, safer surface than Omni's full app-in-iframe (no SQL, no model editing, RLS always on). Ship as `embed_mode: create`.
- **Embedded scheduled reports** for embed users (their email, their filters) (M).
- **Embed AI hardening**: whitelabel the chat (name/icon/intro — Omni's AI branding), per-embed-org AI usage caps (see §5.4).
- **Embed vanity domains** productized on the existing custom-domain machinery, with documented first/third-party cookie behavior.

### 3.5 Governed SQL surface in Cloud (M)
**Omni:** SQL tabs + bidirectional SQL↔UI; Querier role gates SQL.
**Plan:** Don't build SQL↔UI bidirectionality (huge, and Rill's model is files). Do ship a **cloud SQL console** for editor+ roles: Metrics SQL + raw OLAP SQL against the project (both engines already exposed via `rill query` / `query_sql` AI tool / `QueryResolver`), with result grid, chart preview, save-as-custom-API or hand-to-AI-agent ("turn this into a model/metrics view"). Closes "my analysts know SQL and have nowhere to type it" without a new document model.

---

## 4. Phase 3 — Structural bets (6–12 months)

### 4.1 Semantic joins: metrics-view relationships + composite metrics views (XL) — the big one
**Omni:** relationships + topics (join trees, symmetric aggregates), composite topics for multi-fact (per-topic aggregation subqueries FULL OUTER JOINed on shared dimensions).
**Plan — leapfrog rather than copy.** Omni joins *tables* then protects against fan-out; Rill can join *at the metrics layer*, which is safer and matches its architecture:
1. **Shared/conformed dimensions** (`runtime/metricsview`): declare `shared_dimensions:` across metrics views in a project (name + mapping per view). Immediately powers cross-metrics-view canvas filters (upgrade of §2.3's name-matching) and one filter bar governing widgets from different views.
2. **Composite metrics views** (new resource or `composite:` mode in `parse_metrics_view.go`): reference 2+ metrics views; each resolves at its own grain via its own executor; results stitched with FULL OUTER JOIN on shared dimensions in the OLAP engine; `shared_measures:` (cross-view derived measures, e.g. revenue/spend across two facts) computed post-join. Grain-correct by construction — fan-out is impossible because aggregation precedes the join. Explore/canvas/AI treat a composite like any metrics view.
3. **Dimension-table relationships** (later): `joins:` on a metrics view to dimension tables (many-to-one only at first) — generalizing today's lookup dimensions to multi-column dimension tables, with the executor emitting the join pre-aggregation. Many-to-one-only sidesteps most symmetric-aggregate complexity; add fan-out-safe distinct aggregates only if/when one-to-many demand is proven.

This single item flips the #1 competitive loss ("no joins") while preserving Rill's "metrics-first, correctness by construction" story — a genuine differentiator vs Omni's LookML-style join web.

### 4.2 AI operationalization suite (L)
**Omni:** evals (prompt sets + LLM judge + branch comparison), suggestions (feedback → model diffs), auto-learn, AI Hub (quality/usage dashboards), credits with per-user/entity caps, skills, routines, Slack agent.
**Plan:** Rill's agent architecture (`runtime/ai/`: router/analyst/developer/feedback agents, internal evals) already contains the raw material — productize the loop:
- **Customer-facing evals** (M): promote `runtime/ai/evals` — orgs define prompt sets against their project, run on any branch deployment (branch-scoped MCP/deployments exist), LLM judge + side-by-side run comparison before merge. Very cheap relative to differentiation: almost nobody but Omni has this.
- **AI suggestions / auto-learn** (M): mine `feedback_agent` output + conversation history into proposed `ai_instructions`/`synonyms`/description diffs, delivered as **PRs via the developer agent** (again: promotion-through-code-review as the governance story).
- **AI usage & quality admin page** (M): conversations, thumbs, token/cost per user & per embed org; per-org and per-embed-entity **AI spend caps** (needed for embedded AI resale).
- **Slack app: ask-Rill** (M): NL question in Slack → analyst agent → chart image (needs §2.4 PNG) + link back to explore with filters applied. High-visibility demo feature; Rill's slack driver + MCP make this a composition job.
- **Scheduled AI reports ("routines")** (S–M once Slack/PNG exist): a report type whose body is an agent prompt; delivers narrative + charts on cron; conditional send reuses alert criteria.
- **Agent skills** (S): `skills:` blocks in `rill.yaml` (name, prompt, params) surfaced as one-click actions in chat and as MCP prompts.

### 4.3 Rich tables & KPI v2 (M)
**Omni:** display-as (bars/images/links/markdown in cells), conditional-formatting rules UI, subtotal groups, totals overrides, transpose; composable KPI (progress-to-goal, comparison, sparkline, image).
**Plan:** Port pivot's conditional-formatting schemes to the canvas table + rules editor; in-cell data bars; `markdown`/`link`/`image` render modes per dimension (pairs with §2.6 field metadata); row/column totals with per-measure override; measure-as-rows in pivot (removing a long-standing limitation). KPI grid gains progress-to-target (`target:` from a measure or constant) and comparison-vs-dimension modes.

### 4.4 Theming v2 + localization (M)
**Omni:** deep document themes (fonts incl. Google Fonts, tile/control styling, JSON import/export), 16 locales.
**Plan:** Extend theme YAML to fonts + component tokens (canvas/explore already consume CSS vars — formalize the token set and document it); theme JSON import/export for embed customers. Finish the in-flight i18n migration and ship the top 5–6 locales — cheap differentiation vs "checkbox localization" and required in EU/JP deals.

---

## 5. Positioning & GTM moves (parallel, not eng-gated)

1. **Publish a Rill vs Omni comparison page** anchored on: total cost at query volume (warehouse compute vs Rill managed OLAP), time-to-first-dashboard (rill start vs SaaS onboarding), embedded latency SLOs, transparent pricing. Omni does this against Looker/Sigma/etc. — absence reads as weakness.
2. **"Speed benchmark" content**: event-scale dataset (10B rows), Rill vs warehouse-BI latency and cost per 1k dashboard interactions. Omni cannot rebut architecturally.
3. **Looker-migration play, Rill flavor**: Omni's biggest funnel is Looker refugees. Rill's AI developer agent + `create_chart`/`develop_file` tools can ingest LookML and emit metrics views — ship a `rill migrate lookml` skill (M) and market it. Also target **Explo refugees** (Omni acquired Explo; embed customers forced to migrate are in play *now*).
4. **Lead every AI conversation with MCP-first**: Rill ships a governed, write-capable MCP server even in the free local product; Omni's is read-oriented and SaaS-gated. "Your agents can *build* in Rill, not just ask."
5. **Price against seats**: Omni = per-seat + $1/AI-credit overage + contract-gated embed. Rill's usage-based/viewer-free model should be named loudly in embed and large-viewer-count deals.

---

## 6. Deliberate non-goals (and why)

| Omni feature | Why we skip it |
|---|---|
| Workbook document paradigm (tabs, drafts, linked tabs) | Different product philosophy; §3.1/§3.5 capture the underlying jobs (ad-hoc calcs, SQL) without a second content model to govern |
| In-app Excel spreadsheet engine + XLSX import w/ formulas | Enormous surface, low fit with Rill's buyer; live Sheets delivery (§2.4) covers the "get it into a spreadsheet" job |
| Bidirectional SQL↔UI parsing | Only coherent when the UI owns the model; Rill's model is files + AI agent |
| dbt compiler embedding / virtual schemas | Rill *is* the transformation layer; deepening dbt-source interop (read manifests for lineage/docs) is enough (candidate S item) |
| Compute routing / per-user warehouse OAuth | Warehouse-centric problems Rill's architecture makes moot; revisit only if warehouse-direct OLAP mode becomes a first-class sales motion |
| 14-warehouse live-query breadth | Rill's story is "bring data to a fast engine," not "query anything in place"; keep investing in connector breadth for *ingestion* instead |
| SaaS-only polish items (session recording integrations, etc.) | Not deal-deciding |

---

## 7. Sequencing summary

| Phase | Theme | Items |
|---|---|---|
| 1 (0–3 mo) | Kill checklist losses | Charts (maps/sankey/waterfall/gauge/boxplot/histogram, trendlines, ref lines), drill-to-records, canvas cross-filtering, delivery destinations + PNG + bursting, audit logs, SCIM, SSO UI, fiscal time, masking, synonyms |
| 2 (3–6 mo) | Self-service depth | Ad-hoc measures + AI promotion, canvas controls & filter UX, layout v2 (pages/kiosk/auto-refresh), embed create-mode + embedded reports, cloud SQL console |
| 3 (6–12 mo) | Structural bets | Shared dimensions → composite metrics views → dimension joins; AI ops suite (evals, suggestions, usage/caps, Slack agent, routines, skills); rich tables/KPI v2; theming + i18n |
| GTM (now) | Positioning | vs-Omni page, speed/cost benchmark, LookML + Explo migration plays, MCP-first AI story, seat-pricing contrast |

**The one-line strategy:** neutralize Omni's checklist advantages fast (Phase 1), close the self-service gap in a Rill-native way (Phase 2), and convert Rill's architectural advantages — metrics-first correctness, managed OLAP speed, code-first governance, write-capable AI — into categories Omni structurally cannot follow (Phase 3 + GTM).
