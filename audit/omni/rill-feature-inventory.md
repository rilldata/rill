# Rill Feature Inventory (current state)

> Basis for the Omni vs Rill comparison. Derived from the codebase (this repo, Aug 2026) and docs.rilldata.com. Facts are codebase-derived unless marked "docs".

## 1. Data & Modeling

### Connectors/drivers — complete list of `runtime/drivers/` subdirectories
`admin` (internal: connects runtime to the cloud control plane), `athena`, `azure` (Blob), `bigquery`, `claude` (Anthropic AI service), `clickhouse`, `databricks`, `druid`, `duckdb`, `file` (local files), `gcs`, `gemini` (AI service), `https`, `mock`, `mysql`, `openai` (AI service), `pinot`, `postgres`, `redshift`, `s3`, `salesforce`, `slack` (notifier only), `snowflake`, `sqlite`, `starrocks`.

Docs additionally list as data sources (built on these drivers): Delta Lake, Iceberg, Kafka (via ClickHouse), Google Sheets, Supabase, local file, HTTPS. Driver interface roles (`runtime/drivers/README.md`): OLAPStore, ObjectStore, warehouse/SQLStore, FileStore, RepoStore, CatalogStore, RegistryStore, Transporter.

### OLAP engines (metrics views can query these directly)
- **DuckDB** (embedded, default), incl. **MotherDuck** (`duckdb/model_executor_motherduck_self.go`) and **DuckLake** (docs `olap/ducklake`)
- **ClickHouse** (self-managed or Rill-managed)
- **Apache Druid**
- **Apache Pinot**
- **StarRocks** (`runtime/drivers/starrocks`, docs page exists)
- Docs also document "OLAP" mode (metrics views directly on warehouse tables) for **BigQuery, Snowflake, Databricks** (`developers/build/connectors/olap/*`). Multiple OLAP connectors per project are supported (`olap/multiple-olap`).

### Models (`runtime/parser/parse_model.go`)
- SQL models (any connector dialect), source models (YAML ingest), staging models (`stage:` for cross-connector flows, e.g. Snowflake→S3→ClickHouse)
- **Incremental models** (`incremental: true`, state-based), **partitioned models** (`partitions:` resolver, `partitions_watermark`, `partitions_concurrency`), partition retry/skip via CLI (`rill project partitions`, `skip-partition`)
- Cross-connector `output:` (write model results to a different connector)
- **Data quality tests** (`tests:` on models; docs `models/data-quality-tests`)
- Scheduled refresh (cron via `refresh:`), full/incremental refreshes, templating with `{{ .env }}` / dev-vs-prod partial data (`parse_partial_data.go`)
- Pre/post exec hooks and migrations (`parse_migration.go`)

### Metrics views (`runtime/parser/parse_metrics_view.go`, `proto/rill/runtime/v1/resources.proto`)
- Bound to exactly **one** model or table (`model:`/`table:` + connector/database/schema). No multi-fact joins at the semantic layer.
- Dimensions: expression or column, **lookup dimensions** (`lookup_table`/`lookup_key_column`/`lookup_value_column`/`lookup_default_expression` — join-at-query-time for dimension enrichment), unnesting, URI dimensions, **additional time dimensions** with per-dimension `smallest_time_grain`.
- Measures: simple aggregates; **derived measures** (`MEASURE_TYPE_DERIVED`, `referenced_measures`, `per_dimensions` for fixed/level-aware measures); **time-comparison measures** (`MEASURE_TYPE_TIME_COMPARISON`); **window measures** (partition, order, frame expression — `MetricsViewMeasureWindow`); quantiles, case statements (docs); `format_preset`/`format_d3` + d3 locale; `valid_percent_of_total`, `treat_nulls_as`, `lower_is_better`.
- **Derived metrics views** (`parent:` + `parent_dimensions`/`parent_measures` with wildcard/regex/exclude selection; can narrow security, override time grain; cannot define new fields) — docs `metrics-view/derived-metrics-views`.
- **Rollups** (pre-aggregated tables with automatic query routing — `MetricsViewSpec.Rollup`), **annotations** (event/range annotations from a table, rendered on time series), result **cache controls** (`cache.key_sql`, `key_ttl`), `watermark`, `max_query_time_range`, `first_day_of_week`/`first_month_of_year`, `ai_instructions`.
- **Security policies** (`SecurityRule` in resources.proto): `access` (boolean/templated expression over user attributes), **row-level filters** (raw SQL or expression), **column-level access** (`field_access` include/exclude per dimension/measure), transitive access; templating against JWT claims (email, domain, admin, groups, custom attributes); mock users for local testing (`rill.yaml mock_users`); "view as user" feature in cloud (`web-admin/src/features/view-as-user`).

### Themes (`runtime/parser/parse_theme.go`)
Primary/secondary colors plus **light/dark variants with arbitrary CSS variable maps**; themes referenced by explores/canvases or inlined; org logo/favicon upload (`rill org upload-logo/upload-favicon`). Dark mode toggle in dashboards.

## 2. Dashboards

### Explore dashboards (`runtime/parser/parse_explore.go`, `web-common/src/features/dashboards/`)
- One metrics view per explore. Measures panel (big numbers + sparkline time series), dimension leaderboards (multi-measure leaderboards supported), searchable toplists.
- **Time Dimension Detail (TDD)**: expanded time series per measure, dimension comparison rows, export, jump to pivot.
- **Pivot**: drag-and-drop rows/columns of dimensions + time grains + measures, expandable nesting, **flat table mode**, full-screen, **conditional formatting** (color scales/rules — `pivot/types.ts` PIVOT_FORMATTING_SCHEMES), export via ExportMenu (`pivot/PivotDisplay.svelte`). Limitation: measures only as column splits.
- Time controls: presets + custom ranges, **Rill ISO 8601 extensions / rilltime syntax** (docs `reference/time-syntax`), timezone pinning/locking, comparison periods (configurable `time_ranges` + `comparison_offsets`), comparison dimension mode.
- Filters: dimension include/exclude, measure (having) filters, search; filter state fully URL-encoded (docs `reference/url-syntax/url-parameters`).
- **Bookmarks** (personal + shared/default, cloud-stored, `admin/server/bookmarks.go`), **public shareable URLs** (magic auth tokens: per-metrics-view filters, field restrictions, TTL/expiry, managed via UI/CLI — `admin/server/magic_tokens.go`), banners, `defaults:` presets, `embeds_hide_pivot`.
- Keyboard shortcuts, dark mode, rows viewer (underlying row inspection in Developer).

### Canvas dashboards (`web-common/src/features/canvas/`)
Layout: rows/items with widths, **tab groups** (`canvasTabYAML`), max width/gaps, variables (inputs), canvas-level and widget-level filters (`canvas/filters/`: time range, comparison pill, dimension filters; per-widget filter overrides), banner, theme, security policy, drag-and-drop visual editor with inspector, **AI canvas generation** (`canvas/ai-generation/generateCanvas.ts`), **PDF export of canvas** (`web-common/src/features/exports/pdf/`).

Component types (`canvas/components/`):
- `kpi` (single KPI, legacy), `kpi_grid` (up to 10 measures, delta/percent comparison, sparklines), `leaderboard`, `table`/pivot (measures+dimensions table), `markdown` (incl. **dynamic markdown** via `metrics_sql`/`metrics_sql_rows` Go templating), `image`, reusable `component` references (component resource with inputs/outputs, `parse_component.go`).
- Charts (`canvas/components/charts/`, union in `web-common/src/features/components/charts/types.ts`): **bar_chart, line_chart, area_chart (stacked), stacked_bar, stacked_bar_normalized (100%), donut_chart, pie_chart, funnel_chart, heatmap, combo_chart (bar+line dual axis), scatter_plot (color+size encodings, zoom/pan)** — 11 named types over 6 renderers (Cartesian, Circular, Funnel, Heatmap, Combo, ScatterPlot).
- **Custom chart**: raw **Vega-Lite spec** component with an AI prompt assistant to generate the spec (`custom-chart/AgenticChartPrompt.svelte`, `chart-ai-agent.ts`).
- All data widgets have "Go to explore" drill-through.

## 3. Visualization gaps (chart types that do NOT exist in canvas)
No built-in: **maps/geospatial (choropleth, point map), treemap, sankey, waterfall, gauge/bullet, radar, boxplot/distribution, histogram widget, sparkline widget (only inside KPI), word cloud, network graph, dual pie, candlestick**. Escape hatch is the Vega-Lite custom component (hand-written or AI-generated spec), not first-class widgets. Charts are bound to metrics views (no arbitrary-SQL chart widget except via component resource + custom API/SQL resolver data). No cross-widget "click to cross-filter" interaction model comparable to Omni/Looker dashboards (filters are explicit canvas/widget filters).

## 4. AI features
- **Runtime AI engine** (`runtime/ai/`): multi-agent system — `router_agent` routes to `analyst_agent` (data Q&A), `developer_agent` (edits project files: `develop_file`, `write_file`, `create_chart`, file list/read/search, `query_sql`, `show_table`, `list_tables`, bucket listing, `project_status`, `navigate`), `feedback_agent`; evals directory included. Tools: `list_metrics_views`, `get_metrics_view`, `query_metrics_view_summary`, `query_metrics_view`, `query_sql`, `canvas_get`, etc.
- **MCP server** built into the runtime (`runtime/server/mcp.go`, `runtime/ai/mcp.go`): per-project HTTP MCP endpoint (`/v1/orgs/{org}/projects/{project}/runtime/mcp`, branch-scoped variant, local `localhost:9009/mcp`), OAuth flow for Claude Desktop/Claude Code/ChatGPT; tool set filtered by user permissions; write tools available with edit access.
- **AI Chat in product** (`web-common/src/features/chat/`: `DashboardChat`, `ProjectChat`, `DeveloperChat`): project-level chat tab in Rill Cloud, dashboard-context-aware side panel (knows current filters/time range/comparisons), renders interactive charts inline, links back to Explore with filters pre-applied; conversation persistence APIs (`ListConversations`, `GetConversation`, `ShareConversation` — shareable chats, `ForkConversation`) in `proto/rill/runtime/v1/api.proto`; per-embed-user chat history via `external_user_id`.
- **CLI chat**: `rill chat` (`cli/cmd/chat`).
- Configurable AI providers: `openai`, `claude`, `gemini` driver connectors; `ai_connector` + project/metrics-view `ai_instructions` (docs `developers/build/ai-configuration`).
- AI-assisted authoring: canvas generation, Vega-Lite chart generation, "Explain and fix error" button (`chat/ExplainAndFixErrorButton.svelte`).

## 5. Alerts, reports, exports, sharing, embedding
- **Alerts** (`runtime/reconcilers/alert.go`, `parse_alert.go`): UI-created or YAML; criteria with operators, comparison periods, thresholds, multiple criteria with AND/OR; checked on data refresh or schedule; snooze/renotify; per-user evaluation (`for.user_email`/attributes — alerts respect row-level security); delivery: **email + Slack** (channels, DMs, webhooks). No PagerDuty/Teams/generic-webhook-first-class targets.
- **Scheduled reports** (`parse_report.go`, `admin/reports.go`): scheduled exports of any exportable query, formats CSV/XLSX/Parquet, email + Slack delivery, recipients + unsubscribe, execution history.
- **Exports** (`proto/.../export_format.proto`): CSV, XLSX, Parquet, **PDF** (canvas PDF export). Export surfaces: leaderboards, TDD, pivot (ExportMenu present in code; docs still say "coming soon").
- **Public URLs**: expirable magic tokens with locked pre-applied filters and field restrictions; recipients need no login; managed in UI settings tab and CLI.
- **Embedding** (docs `developers/embed/iframe`, `embed/postmessage`; `web-admin/src/features/embeds`): iframe embed of explore/canvas or whole project dashboard list; service-token → backend-generated iframe URL; `user_email`/`attributes` for row-level security, `external_user_id` for per-user state (incl. AI chat history in embeds), navigation control, `hide_navigation_bar`, theme + theme_mode overrides; **postMessage API** for reading/writing embedded dashboard state.

## 6. Cloud & collaboration (`admin/`)
- Orgs → projects; **org roles: admin, editor, viewer, guest; project roles: admin, editor, viewer** (`admin/database/database.go:960`); fine-grained permission matrix (docs roles-permissions: manage_org, manage_projects, magic token perms, report/alert/bookmark create-vs-manage, dev-vs-prod read/manage split). **User groups** with role assignment; per-user and per-group **resource-level access** (`rill user set-resources`, `usergroup set-resources`); allowlisted email-domain auto-join (`rill user whitelist`); access requests (`web-admin/features/access-request`).
- Auth: Auth0-based (`admin/server/auth/`), Google/GitHub/email; **SAML SSO flows supported via Auth0 enterprise connections** (`handlers.go:289` references SAML flows). **No SCIM provisioning found anywhere in the codebase. No audit log feature found.**
- Tokens: personal access tokens, **service accounts + service tokens** with org/project roles and custom attributes (`rill service create --org-role/--project-role`, token issue/revoke).
- Billing: Orb + Stripe (`admin/billing/`), trial/Team/Enterprise plans, RCU usage-based metering, hibernation of idle projects; usage metrics infra (`admin/metrics/client.go`: usage for billing, autoscaler recommendations, quotas). Admin-visible billing/usage page; no end-user BI-usage/audit analytics surface for customers.
- Custom domains per org (`admin/deployments.go`, `urls.go`), org logo/favicon branding.

## 7. APIs
- **Custom APIs** (`runtime/parser/parse_api.go`, resolvers in `runtime/resolvers/`): YAML-defined HTTP endpoints backed by resolvers — `sql`, `metrics_sql`, `api`, `glob`, `resource_status`, plus built-in `builtin_sql`/`builtin_metrics_sql`; per-API security policies, `skip_nested_security`, **OpenAPI spec generation** (parameters, request/response schemas); templating with user attributes.
- **Runtime query APIs** (`proto/rill/runtime/v1/queries.proto`, 32 RPCs + 56 in api.proto): MetricsViewAggregation, Toplist, Comparison, TimeSeries, Totals, Rows, Schema, Search, TimeRanges, Annotations, Export, ExportReport, column profiling APIs, ResolveCanvas/Component, ConvertExpressionToMetricsSQL. **Metrics SQL** dialect: SQL against metrics views with governance applied — **JOINs are explicitly not supported** (`runtime/metricsview/metricssql/parser_test.go: "metrics sql: join is not supported"`).
- **Admin API**: 168 RPCs (`proto/rill/admin/v1/api.proto`) mapped to REST at api.rilldata.com (orgs, projects, members, groups, tokens, magic tokens, iframe, deployments, billing); documented at docs `/api/admin`.
- **CLI** (`cli/cmd/`): start (local dev server), deploy, project (clone/refresh/partitions/logs/status/hibernate/deployment mgmt), org, user, usergroup, service, token, public-url, env (configure/push/pull), billing, query (`rill query --sql`/`--resolver` against local or cloud), chat, sudo (internal), devtool, validate, docs.

## 8. Dev workflow
- Code-first projects: YAML + SQL, full Git workflow; **GitHub integration** (connected repo, auto-deploy on push) or **direct upload deploy** (Rill-managed repo, `rill deploy`; `admin/server/assets.go`).
- **Branch deployments**: per-branch dev deployments (`cli/cmd/project/deployment`, `web-admin/features/branches`, branch-scoped MCP URLs), roles distinguish read/manage dev vs prod.
- **Edit-in-cloud**: web-admin edit sessions with branch/commit/merge/publish flow (`web-admin/src/features/edit-session/` — CommitPopover, MergePopover, PublishPopover) — the Developer IDE experience surfaced in Rill Cloud.
- Env separation: `dev:`/`prod:` YAML overrides + `environment_overrides`, `vars`, `.env` with `rill env push/pull`, dev partial data limits; templating (`parse_partial_data.go`, `template.go`).
- Reconciler architecture (`runtime/reconcilers/`): declarative DAG, **incremental/partial reconciles** of only changed resources (`runtime/controller.go`), refresh triggers, project parser resource.
- Local Developer app: visual metrics/explore editors (`web-common/features/visual-editing`, `visual-metrics-editing`), column profiling, SQL model workspace, trace viewer for debugging (docs `build/debugging/trace-viewer`).

## 9. Honest limitations vs classic BI tools
- **No joins in the semantic/query layer**: metrics views bind to a single model/table; Metrics SQL rejects JOINs; only escape valves are lookup dimensions (key→value enrichment), derived metrics views (subset of one parent), and doing joins upstream in models. No Omni/Looker-style relationship/topology model across tables.
- **Self-service ad-hoc analysis is bounded**: viewers can only slice predefined dimensions/measures; no viewer-side calculated fields, no ad-hoc dataset exploration, no drag-in of raw columns not modeled by a developer.
- **No SQL runner UI in Rill Cloud** for viewers/analysts: SQL is available locally in Developer, via `rill query` CLI/API, via AI chat's `query_sql` tool, and in cloud edit sessions (developer role) — but there is no general-purpose SQL workbench surface in the cloud UI.
- Pivot: measures cannot be row splits; explore/pivot per single metrics view; no cross-metrics-view dashboards except canvas (canvas widgets can each target different metrics views).
- Exports: no Google Sheets / direct destination push; report delivery is email/Slack only; no alert targets beyond email/Slack.
- Enterprise IT: no SCIM, no audit logs, no customer-facing usage analytics of BI consumption (only billing usage); SSO via Auth0 (SAML possible) rather than native SAML/OIDC config UI.
- Visualization breadth is narrower than incumbent BI (see section 3); theming is colors/CSS-variable level, not full component styling.
- No cache/extract-based semantic acceleration configuration comparable to Looker PDTs beyond models + new rollups feature; no scheduled "delivery to destination" beyond email/Slack.
- No native mobile apps; no offline; write-back only via custom APIs.

Key files for citation: `runtime/drivers/`, `runtime/parser/parse_metrics_view.go`, `proto/rill/runtime/v1/resources.proto` (security rules, measure windows, rollups), `web-common/src/features/canvas/components/charts/index.ts` (chart union), `runtime/ai/` (agents + MCP), `admin/database/database.go` (roles), `admin/server/magic_tokens.go` (public URLs), `runtime/metricsview/metricssql/parser_test.go` (join unsupported), `web-common/src/features/exports/pdf/` (canvas PDF), `web-admin/src/features/edit-session/` (cloud editing).
