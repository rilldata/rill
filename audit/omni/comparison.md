# Omni vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill has it (rough parity or better) · 🟡 Rill has a partial/different equivalent · ❌ Rill does not have it · ➕ Rill advantage Omni lacks. "How it works in Omni" is kept brief; full detail in the three `omni-features-*.md` files.

---

## 1. Semantic layer & modeling

| Feature | How it works in Omni | Rill status | Notes |
|---|---|---|---|
| Auto-generated schema model | Connecting a DB auto-creates views for every table, dimensions for every column, inferred PKs and join relationships (column-name + FK heuristics) | ❌ | Rill projects start from blank YAML; AI developer agent can scaffold, but there is no automatic whole-database model |
| Three-layer model (schema → shared → workbook) | Governed shared model plus a per-workbook overlay; user changes live in the workbook layer until promoted | ❌ | Rill has a single governed layer (project files). No per-user/ad-hoc model overlay |
| Just-in-time modeling / promotion | Workbook-layer fields, joins, calcs reviewed in a diff drawer and promoted to the shared model in one click | ❌ | Closest analog: edit-in-cloud sessions + AI developer agent, but that is developer-gated, not a business-user flow |
| Joins / relationships (`relationships.yml`, join trees) | Central relationship definitions with cardinality, join types, fan-out protection; topics assemble a base view + join tree | ❌ | **Biggest structural gap.** Metrics views bind to one model/table; Metrics SQL rejects JOINs. Escape valves: lookup dimensions, upstream model SQL |
| Topics (curated datasets) | Base view + joins + field curation + AI context; the unit of governed self-service and all AI querying | 🟡 | Metrics views play the same role (curated, governed, AI-grounding via `ai_instructions`) but are single-table and thinner on curation params |
| Composite topics (multi-fact) | 2+ topics each aggregate in their own subquery, FULL OUTER JOINed on shared dimensions; correct grain, no fan-out | ❌ | No multi-fact/metrics-view stitching. Canvas can only place widgets from different metrics views side by side |
| Symmetric aggregates (fan-out safety) | PK + relationship-type-aware SUM/COUNT DISTINCT rewrites so joins never double-count | ❌ | N/A today (no joins), but required the day joins ship |
| Level-of-detail (fixed-grain) measures | `level_of_detail` param compiles to joined subquery (Tableau LOD-style) | ✅ | Rill's derived measures with `per_dimensions` (fixed measures) cover this |
| Window / running measures | Via quick calcs (running total, % of previous) and SQL | ✅ | First-class `window:` measures (partition/order/frame) — arguably cleaner than Omni |
| Time-comparison measures | Period-over-period via query structure + PoP dashboard control | ✅ | `MEASURE_TYPE_TIME_COMPARISON` + comparison offsets in explore |
| Timeframes (auto date grain expansion) | Date fields expand to ~20 grains incl. fiscal quarters/years (`fiscal_month_offset`) | 🟡 | Rill has time grains + `smallest_time_grain`, `first_month_of_year`; **no fiscal calendar support**, no day-of-week/hour-of-day auto-dimensions (must be modeled manually) |
| Filtered measures | `filters:` on a measure | ✅ | Expression-level (CASE/FILTER) in measure SQL; less declarative but works |
| Dimension groups/bins/Top-N fields | Bins, string groups, Top-N-with-Other as model params | ❌ | Must be hand-written as SQL expressions in the model |
| Field metadata breadth | `synonyms`, `sample_values`, `links`, `markdown` render, `colors` (model-defined series colors), `suggest_from_*`, `tags`, `group_label` | 🟡 | Rill has description, format presets/d3, `lower_is_better`, `ai_instructions`; lacks synonyms, per-field links, model-defined colors, markdown cell rendering |
| Model inheritance (`extends`, shared extension models, dynamic extensions by user attribute) | Hub-and-spoke: child models inherit and override; users routed to extensions by attribute | 🟡 | Derived metrics views (subset-only, no overrides/new fields). No model-level inheritance or attribute-based routing |
| Templated SQL filters / Mustache in model | `{{# field.filter }}` templating, filter-only fields, constants, user-attribute injection into SQL | 🟡 | Rill has Go templating with `.env`/user attributes in models & security policies, but no dashboard-filter→SQL-template binding |
| Row-level security | `access_filters` on user attributes, fail-closed, `values_for_unfiltered` | ✅ | Security policies: row filters, templated attributes — parity |
| Column-level security / masking | `required_access_grants` per field; `mask_unless_access_grants` (MD5 hash to preserve joins) | 🟡 | `field_access` include/exclude = parity for hiding; **no masking/hashing mode** |
| Access grants (named, composable) | Named grants with AND/OR on user attributes, applied across topics/views/fields/skills | 🟡 | Per-resource security rules exist but no named reusable grant objects |
| Content validator | Validates all dashboards against the model; finds broken references; bulk find-and-replace; API | 🟡➕ | Rill's parser/reconciler validates the whole project DAG continuously (stronger guarantee); but no bulk find-and-replace or "what references this field" UI |
| Model version history / restore | Chronological diffs + one-click restore | ✅ | Git history (stronger), but no in-UI restore for upload-deployed projects |
| Aggregate awareness | Register warehouse rollup tables (`materialized_query`); queries silently rewritten to smallest sufficient aggregate; HLL/approx sketch support | 🟡 | New `rollup:` on metrics views with automatic routing — same idea, earlier maturity; no sketch/HLL rollups, single rollup vs many registered aggregates |
| Query result caching | Exact-match shared cache, TTL 0s–30d at model/topic/dashboard level, cache warming via deliveries, reset-cache API | ✅ | Runtime query cache + metrics-view `cache.key_sql`/`key_ttl`; plus Rill's whole architecture (managed OLAP) is the bigger performance answer |
| In-browser requeryable cache | Last ~30 result sets re-sliced client-side for instant interactions | 🟡 | Explore is aggressively query-cached and fast by architecture; no explicit client-side requery engine |
| Compute routing / warehouse override | Route job types (dashboards, schedules, API, AI) to different warehouses/slot reservations | ❌ | N/A for managed OLAP; relevant if warehouse-direct mode grows |
| Data quality tests | Not a first-class feature (delegated to dbt) | ➕ | Rill models support `tests:` natively |
| Incremental/partitioned ingestion, cross-connector pipelines | Not applicable — Omni has **no ingestion layer at all**; depends on warehouse + dbt | ➕➕ | Rill's full-stack ELT (incremental models, partitions, staging, cross-connector outputs) is a core differentiator |
| Managed OLAP / sub-second at scale | None; performance = warehouse + caches | ➕➕ | DuckDB/ClickHouse managed engines, Druid/Pinot/StarRocks support — Rill's core wedge |

## 2. Querying & self-service analysis

| Feature | How it works in Omni | Rill status | Notes |
|---|---|---|---|
| Workbooks (multi-tab analysis docs) | Tabs of point-and-click queries, SQL, spreadsheets; drafts/autosave/version history; every dashboard is a workbook's presentation layer | ❌ | Rill's consumption surface is dashboards only. Explore is powerful but not a document users compose |
| Point-and-click ad-hoc querying over any field | Field picker across the topic's join tree; quick aggs; custom fields | 🟡 | Explore covers slice/dice of modeled fields only; no ad-hoc fields, no un-modeled columns |
| Viewer-defined custom fields/calcs | SQL custom fields + ~120 Excel-style table calculation functions (XLOOKUP, SUMIFS, AI_CLASSIFY…), AI formula builder, calc pushdown pre-row-limit | ❌ | **Major self-service gap.** No viewer-side calculated fields anywhere in Rill |
| Excel-style spreadsheet tabs | Live query-connected sheets, XLSX import preserving formulas, viewer-editable sheets on dashboards (what-if) | ❌ | Nothing comparable |
| SQL workbench in cloud | SQL tabs (dialect SQL) with full viz on top; bidirectional Omni SQL↔UI with field promotion | ❌ | SQL only in local Developer / cloud edit sessions (developer-gated); no analyst SQL surface |
| Bidirectional SQL↔UI | UI edits regenerate SQL; SQL edits parse back into model state; unmatched exprs become promotable fields | ❌ | One-way (YAML/SQL → dashboards) |
| Filter by another query | `WHERE IN (subquery)` filters via UI | ❌ | Not expressible in explore filters |
| CSV/XLSX upload by end users | Upload ≤500k rows into a scratch schema, joinable, auto-topic | 🟡 | Rill ingests local files as sources (developer flow); no end-user upload in cloud |
| Input columns / writeback | Editable columns on query results persisted to warehouse scratch schema, editable on live dashboards | ❌ | No writeback |
| Query IR / Run Query API | Inspector exposes query JSON = exactly the Run Query API payload; Arrow results | ✅ | MetricsViewAggregation + Metrics SQL + custom APIs with OpenAPI; parity in capability, different shape |
| Explore-from-anywhere (governed drill to ad-hoc) | Any tile → scratch workbook copy | 🟡 | "Go to explore" from canvas widgets; explore itself is the terminus (no deeper ad-hoc layer) |

## 3. Visualization types

Rill canvas native: bar, line, area (stacked), stacked bar, stacked bar 100%, donut, pie, funnel, heatmap, combo (bar+line), scatter/bubble, KPI, KPI grid, leaderboard, table/pivot, markdown (dynamic), image + Vega-Lite custom escape hatch.

| Viz type | Omni | Rill | Notes |
|---|---|---|---|
| Bar (grouped/stacked/100%/stacked measures) | ✅ | ✅ | Omni adds per-series mark override (mixed marks) and totals labels |
| Line (dual axis, interpolation, step) | ✅ | 🟡 | Rill combo covers dual axis bar+line; no line interpolation/step options |
| Area (stacked/100%/overlay) | ✅ | ✅ | |
| Scatter / bubble | ✅ | ✅ | Parity incl. color+size |
| Pie / donut | ✅ | ✅ | |
| Funnel | ✅ | ✅ | |
| Heatmap | ✅ | ✅ | |
| Sankey | ✅ native | ❌ | Vega-Lite escape hatch only |
| Boxplot | ✅ native | ❌ | |
| Maps (point + choropleth + custom GeoJSON, 3D globe) | ✅ native | ❌ | **Most-cited gap class**; Vega-Lite only |
| KPI (composable: number/comparison/progress/sparkline/text/image) | ✅ | 🟡 | Rill KPI grid has number+delta+sparkline; no progress-to-goal, no composable builder |
| Single record | ✅ | ❌ | |
| Table (rich) | ✅ | 🟡 | See §4 |
| Markdown viz w/ Mustache data + Sparkline/ChangeArrow components + iframes | ✅ | 🟡 | Rill dynamic markdown (metrics_sql templating) is close; no iframe, no component library, no per-row loops over pivots |
| AI summary tile | ✅ | ❌ | No LLM-narrative widget on canvas |
| Custom Vega-Lite | ✅ (any chart → editable spec) | ✅➕ | Rill has Vega-Lite component **with AI spec generation**; Omni's "open any built-in chart as Vega" is bidirectional though |
| Waterfall, gauge, sunburst, Gantt, bullet, trellis/small multiples, word cloud, radial | Vega gallery recipes | ❌ | Same escape hatch idea; Omni ships a documented gallery, Rill doesn't |
| Trendlines (moving average, regression) | ✅ | ❌ | |
| Reference lines | ✅ | ❌ | |
| Gradient color pinning, model-defined series colors | ✅ | ❌ | Rill themes set palette; no per-dimension-value fixed colors |
| Conditional formatting (tables) | ✅ | 🟡 | Rill pivot has conditional formatting schemes; canvas table lacks rules UI |
| Forecasting | ✅ (deterministic stats lib via AI prompt) | ❌ | |
| Annotations on charts | ✅ (AI-powered, GA 6/2026) | ✅ | Rill metrics-view annotations from a table rendered on time series — parity, differently sourced |

## 4. Dashboards (consumption)

| Feature | Omni | Rill status | Notes |
|---|---|---|---|
| Grid + nested flex containers, lockable | New editor: Grid & Stack containers, nesting, named, lockable | 🟡 | Canvas rows/items+widths is simpler; no nesting/free-form positioning |
| Pages/tabs (≤15) + page navigation styles + custom URL tabs | Multi-page dashboards, tabs/buttons/dropdown nav, external URL tabs | 🟡 | Canvas tab groups exist; no per-page filters, no custom URL tabs |
| Mobile responsiveness with per-container overrides | Auto-adapt + hide-on-mobile etc. | 🟡 | Canvas is responsive-ish; no per-widget mobile controls |
| Cross-filtering (click chart → filter dashboard, brush ranges) | Opt-in per dashboard; joins resolved via model | ❌ | **Major interaction gap** — no click-to-filter on canvas |
| Faceted filters (filter narrows another's suggestions) | Opt-in | ❌ | |
| Dynamic filter controls (viewers add own filters) | Creator scopes fields; viewer filters URL-only | 🟡 | Explore gives viewers full filtering; canvas viewers get authored filters only |
| Filter display options (buttons, toggles, calendars, curated presets) | Rich per-filter display config | ❌ | Canvas filter pills only |
| Controls: field switcher, multi-field picker, timeframe switcher, PoP control, Top-N control | Viewer-facing widgets that mutate queries | ❌ | Canvas has variables (inputs) plumbing but no such control widgets |
| Required / hidden filters, user-attribute filter defaults | Per-filter settings | ❌ | |
| Filter state in URL | `f--` params | ✅ | Full URL-encoded state in explore and canvas |
| Measure drill to records + curated drill fields + multiple named drill queries | Click value → records; `drill_fields`/`drill_queries` in model | ❌ | Rill has leaderboard→TDD→pivot flows and rows viewer in Developer, but no click-a-number→underlying-rows in cloud dashboards |
| Hierarchical dimension drill | `drill_fields` on dimension swaps in children + filters | ❌ | |
| Link drilling (`links` on fields, cross-dashboard links carrying values) | Model-defined per-row URLs | 🟡 | URI dimensions render links; no templated cross-dashboard link builder |
| Auto-refresh (wallboards) | Interval refresh bypassing cache | ❌ | |
| Timezone override UX (viewer picker, triple opt-in) | Dashboard header TZ picker | ✅ | Explore timezone pinning/selector — parity |
| Bookmarks / personal views | Not a first-class feature (URL state + saved views) | ➕ | Rill bookmarks (personal/shared/default) are stronger |
| Scheduled exports from dashboard | ✅ | ✅ | |
| Dashboard download PDF/PNG/XLSX/CSV | 4 formats + layout options | 🟡 | Canvas PDF; exports CSV/XLSX/Parquet per query; **no PNG, no whole-dashboard XLSX/CSV bundle** |
| Themes (fonts, tiles, controls, JSON import/export) | Deep document themes | 🟡 | Rill themes = colors + CSS vars, org logo; no font/tile-level theming, no theme JSON |
| Custom error/no-results messages | Per-theme Markdown | ❌ | |
| Performance profiler + per-tile SQL inspect | Tile Inspect shows SQL/cache/execution | 🟡➕ | Developer has trace viewer + query inspection; not exposed on cloud dashboards |
| Document abilities gating (download/drill/schedule/AI per doc) | Org defaults + per-document overrides | 🟡 | Rill has `embeds_hide_pivot`, export perms by role; not per-document ability matrix |
| Presentation/full-screen mode | ✅ (2026) | ❌ | |
| Spreadsheet tiles on dashboards (viewer-editable what-if) | ✅ | ❌ | |

## 5. AI

| Feature | Omni | Rill status | Notes |
|---|---|---|---|
| Semantic-grounded NL querying (never text-to-SQL for viewers) | Agent emits semantic query JSON compiled by Omni; RLS always applies | ✅ | Rill analyst agent queries metrics views with governance — same architecture claim |
| Standalone chat + dashboard-context chat | Omni Agent + Dashboard Agent (tile-scoped) | ✅ | ProjectChat + DashboardChat (filter/time-context aware) — parity |
| AI builds dashboards | Plans layout, publishes dashboard, iterates; screenshot replication | ✅ | AI canvas generation + developer agent; screenshot-to-dashboard not verified |
| AI builds the semantic model | Modeling Agent (3 apply modes), model generation GA, "suggest model from query history" | 🟡 | Developer agent edits files incl. metrics views; no query-history mining, no guided modes |
| Agent skills (modeler-defined AI workflows in YAML) | `skills:` blocks → one-click packaged prompts with access grants | ❌ | `ai_instructions` is context, not packaged workflows |
| AI Routines (scheduled NL prompts → email/Slack narrative) | Cron-scheduled agent runs, conditional notify | ❌ | Rill reports/alerts are query-based, not agentic |
| Slack Agent (@Omni in Slack) | NL Q&A with charts in Slack threads | ❌ | Slack is delivery-only in Rill |
| Evals (prompt sets + LLM judge + branch comparison) | Customer-facing eval harness for their own model | ❌ | `runtime/ai/evals` is internal-dev only |
| Suggestions (AI mines feedback → model-improvement diffs) | Prioritized YAML improvement queue | ❌ | |
| Auto-learn from conversations | Captures stated business context → proposes model updates | ❌ | |
| AI Hub (admin: quality/usage/credit dashboards) | Central AI observability | ❌ | No customer-facing AI usage/quality surface |
| MCP server | Remote MCP, OAuth 2.1, per-user perms, askOmni agentic jobs | ✅➕ | Rill's runtime MCP (OAuth, per-project, branch-scoped, write tools) is at least parity; Rill also ships MCP in the free local product |
| BYO LLM | Bedrock default; Anthropic/OpenAI/Vertex/Cortex/Genie/Grok | ✅ | openai/claude/gemini connectors — parity in kind, fewer providers |
| AI credit metering & limits (org/user/embed-entity) | Credits, caps, downgrade tiers, $1/credit overage | ❌ | No AI usage metering/limits exposed to customers |
| AI context curation params | `ai_context` at 6 levels, `synonyms`, `sample_queries`, `ai_fields`, `ai_chat_topics` | 🟡 | `ai_instructions` at project + metrics-view level; no synonyms/sample_queries/field allowlists |
| Voice input, file/image upload as context | ✅ | ❌ | |
| AI forecasting | ✅ | ❌ | |
| Shareable chat sessions | ✅ | ✅ | ShareConversation API |
| External agent tool connections (Drive/Notion/GitHub in chat) | Per-user OAuth integrations | ❌ | (An external-MCP connector design exists on branch `mcp_external`) |
| Open-source agent skills for Claude Code/Cursor | 9 skills, 3 agents on GitHub | 🟡 | Rill ships project-level Claude skills (rill-development etc.) with the repo; not marketed as a pack |

## 6. Embedded analytics

| Feature | Omni | Rill status | Notes |
|---|---|---|---|
| Signed-URL SSO embed (1-step & 2-step) | HMAC-signed URLs; TS SDK | ✅ | Service-token → iframe API; parity in mechanism |
| Row-level tenant security via attributes | access_filters + userAttributes | ✅ | `user_email`/`attributes` on iframe API |
| Per-user state in embeds | externalId identity | ✅ | `external_user_id` (incl. AI chat history) |
| Embed events JS bridge | Rich bidirectional postMessage events | ✅ | postMessage API — parity in kind, fewer event types |
| Embedded self-service ("Create Mode": end users build dashboards) | Full app in iframe; entity folders; per-entity sharing | ❌ | **Embed differentiator gap** — Rill embeds are read-only consumption |
| Embedded AI chat | Whitelabeled agent page, credit limits per entity | 🟡 | Chat works in embeds w/ per-user history; no whitelabel branding or credit limits |
| Vanity domains + wildcard | analytics.customer.com, first-party cookies | 🟡 | Org custom domains exist; embed-specific vanity/wildcard + cookie strategy not productized |
| Embed themes via URL | theme JSON/id + prefersDark | ✅ | theme + theme_mode params |
| Deliveries by embed users | Embed users can schedule | ❌ | |
| Embed session/user admin UI, multiple secrets | Session mgmt, impersonation, secret rotation | 🟡 | Service tokens; no embed-session admin surface |
| Per-user database OAuth in embeds | Snowflake/BQ/Databricks passthrough | ❌ | |
| Dynamic schema/DB/model routing per tenant | dynamic_schemas, connection environments, dynamic extensions | 🟡 | Templating + env vars can vary sources; no per-user-attribute connection routing |

## 7. Delivery, alerts, exports

| Feature | Omni | Rill status | Notes |
|---|---|---|---|
| Destinations | Email, Slack, **Webhook, SFTP, S3, Google Sheets** | 🟡 | Email + Slack only |
| Formats | PNG, PDF, XLSX, CSV | 🟡 | CSV/XLSX/Parquet (+canvas PDF); no PNG; ➕ Parquet |
| Alert conditions | results changed/same/none/any vs baseline | ✅➕ | Rill criteria (thresholds, comparisons, AND/OR) are richer; Omni's "results changed" baseline diffing is a nice primitive Rill lacks |
| Per-recipient personalization (bursting) | Rebuild content per recipient's user attributes | 🟡 | Alerts evaluate per-user RLS (`for.user_email`); reports don't burst per recipient |
| Email-only (license-free) recipients w/ attributes | Auto-created, attribute-assignable | 🟡 | Report recipients can be external emails; no attribute assignment |
| Email whitelabeling, domain allowlist | Custom sender domain | ❌ | |
| Cron + visual schedules, test-send | ✅ | ✅ | |
| Public share links | Not a documented feature (signed embed URLs instead) | ➕ | Rill magic-token public URLs with filters/expiry are a differentiator |

## 8. Governance, enterprise, admin

| Feature | Omni | Rill status | Notes |
|---|---|---|---|
| RBAC | 6 connection-role tiers + custom roles (granular ability toggles) + per-model overrides | 🟡 | Org/project roles + groups + resource-level access; no custom roles, no per-ability toggles |
| SSO SAML/OIDC | Okta/Entra/Google/Rippling/generic OIDC, config UI | 🟡 | Via Auth0 enterprise connections; no self-serve config UI |
| SCIM provisioning + attribute sync | Okta/Entra/Rippling | ❌ | Absent from codebase |
| Audit logs | JSON event streams to S3/GCS/Azure, 1yr retention | ❌ | Absent |
| Usage analytics for admins | Omni-in-Omni dashboards (users, content, model perf, AI quality) | ❌ | Only billing usage |
| User attributes system | Typed, multi-value, SCIM-synced, drives everything | 🟡 | Custom attributes on users/service tokens for policies; thinner (no typed admin UI, no IdP sync) |
| Impersonation / view-as | Full session impersonation + attribute preview | 🟡 | View-as-user in dashboards; no full impersonation |
| Content org: folders (7 levels), labels, verified, access requests | Cascading folder permissions | 🟡 | Projects as containers + resource-level grants + access requests; no folders/labels within a project |
| Compliance | SOC 2 II, ISO 27001, HIPAA, GDPR | 🟡 | (Verify current Rill certifications; not in codebase scope) |
| Deployment | SaaS only, AWS+Azure, 8 regions | ➕ | Rill: local free product + cloud; single-binary developer experience Omni cannot match |
| Localization (16 locales) | ✅ | ❌ | English only (i18n migration in progress) |
| Pricing transparency | Quote-only, per-seat + AI credits | ➕ | Rill has public usage-based pricing, free local tier |

## 9. Where Rill is ahead (Omni lacks or is weaker)

1. **Full-stack data plane**: ingestion connectors (25+ drivers), incremental/partitioned models, cross-connector staging, data quality tests, scheduled refresh orchestration. Omni starts where the warehouse ends.
2. **Managed real-time OLAP**: embedded DuckDB, managed ClickHouse, Druid/Pinot/StarRocks — sub-second interactive dashboards at event-scale without warehouse cost-per-query. Omni's answer is caching in front of the customer's warehouse.
3. **Explore UX for operational analytics**: opinionated leaderboards + TDD + comparisons — faster time-to-insight than assembling a workbook.
4. **Local-first, open developer workflow**: free `rill start`, single binary, project as plain files, branch deployments, true git-native (Omni syncs YAML to git; Rill *is* files).
5. **Public share URLs** with scoped filters/expiry.
6. **Bookmarks** (personal/shared/default states).
7. **Parquet exports**; alert criteria richness; per-user alert evaluation under RLS.
8. **MCP in the free local product**; write-capable MCP (Omni's MCP is read/query oriented).
9. **Custom APIs** (YAML-defined governed endpoints with OpenAPI) — Omni has a query API but not user-defined governed endpoints.
10. **Transparent usage-based pricing** and an OSS core.

## 10. Omni's biggest structural advantages over Rill (summary)

1. **Joins/topics/multi-fact semantics** — the single largest gap.
2. **Self-service depth**: workbooks, viewer calcs (~120 Excel functions), spreadsheets, SQL tabs, bidirectional SQL↔UI, promotion workflow.
3. **Visualization breadth + dashboard interactivity**: maps, sankey, boxplot, trendlines, reference lines; cross-filtering, controls, multi-page, drill-to-records.
4. **Delivery matrix**: 6 destinations × 4 formats × per-recipient bursting.
5. **Enterprise IT pack**: SCIM, audit logs, custom roles, usage analytics, impersonation, localization.
6. **AI operationalization**: evals, suggestions, auto-learn, credits, AI Hub, skills, routines, Slack agent — Omni productized the AI *feedback loop*, not just the chat.
7. **Embedded self-service (Create Mode)** including embedded creation, deliveries, and AI with per-tenant credit limits.
8. **dbt bi-directional integration** and warehouse-native ops (compute routing, aggregate awareness, database OAuth).
