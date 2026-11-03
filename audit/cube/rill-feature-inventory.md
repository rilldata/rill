# Rill Feature Inventory (current state, Cube-focused)

> Basis for the Cube vs Rill comparison. Derived from the codebase (this repo, Aug 2026) and docs.rilldata.com. Emphasis on the surfaces where Cube competes: semantic layer, APIs, caching, AI. A broader BI-surface inventory (viz types, dashboards, alerts) lives in [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) and still applies.

## 1. AI features (exhaustive)

### 1.1 Agent architecture (`runtime/ai/`)

Multi-agent AI subsystem built into the runtime (`runtime/ai/ai.go`), with typed Go tools implementing MCP-native specs (`modelcontextprotocol/go-sdk`), auto-generated JSON schemas, and per-tool access control (`CheckAccess`) driven by runtime security claims.

**Agents:**
- **Router Agent** (`router_agent.go`): entrypoint for chat; LLM-routes to candidate agents.
- **Analyst Agent** (`analyst_agent.go`): data Q&A over metrics views; accepts rich context (explore dashboard, canvas + component, dimension/measure allowlists, filters, time + comparison range, `is_report` flag); pre-invokes `get_metrics_view`/`query_metrics_view_summary` for dashboard context.
- **Developer Agent** (`developer_agent.go`): edits the project — 20-iteration tool loop over file tools, SQL tools, bucket introspection, `develop_file`; supports `init_project` (whole-project scaffolding from a prompt).
- **Feedback Agent** (`feedback_agent.go`): records thumbs up/down; on negative feedback runs an LLM **attribution** step (Rill's fault vs project's vs user's, with suggested action) for triage.

**Worker tools:** `list_metrics_views`, `get_metrics_view`, `get_canvas`, `query_metrics_view_summary` (time range + sample values), `query_metrics_view`, `create_chart` (in-chat charts, `chat_charts` flag), `develop_file` (sub-agent: create/edit/delete one YAML/SQL file, returns post-reconcile status), `list_files`, `search_files`, `read_file`, `write_file`, `project_status`, `query_sql` (raw OLAP SQL), `list_tables`, `show_table`, `list_buckets`, `list_bucket_objects` (S3/GCS discovery), `navigate` (drives the Rill UI).

Completion loop (`Session.Complete`): multi-iteration tool-calling (default 10, developer 20), context truncation, structured outputs, OTel tracing, token/tool-call telemetry.

### 1.2 Conversations
Persisted sessions/messages in the per-instance catalog store as a call tree (user input, tool calls/results, thinking). **Sharing** with partial visibility (`SharedUntilMessageID`), **forking** (`ForkSession`), anonymous sessions for public projects. Runtime RPCs: `Complete`, `CompleteStreaming`, `ListConversations`, `GetConversation`, `ShareConversation`, `ForkConversation`, `ListTools`, `GetAIMessage`.

### 1.3 LLM providers
AI is a connector driver (`drivers.AIService`): **OpenAI** (incl. Azure OpenAI), **Anthropic Claude** (default `claude-opus-4-5`), **Google Gemini**, plus the **admin proxy** (Rill-managed AI: runtime proxies completions through the admin service holding the provider key). BYO model via `ai_connector` in `rill.yaml`; `managedAI` flag distinguishes billable vs BYO tokens.

### 1.4 Context injection
`ai_instructions` at project level (`rill.yaml`) and metrics-view level, surfaced in MCP responses; `@`-context picker in the chat UI (`web-common/src/features/chat/core/context/`).

### 1.5 MCP server
- Session-scoped MCP server (`runtime/ai/mcp.go`) exposing exactly the tools the caller's claims permit: metrics tools for readers; file read/**write** tools with edit access. Mounted at `/v1/instances/{id}/mcp` and local `/mcp` (`runtime/server/mcp.go`), stateless streamable HTTP.
- **Full MCP OAuth**: RFC 9728 + RFC 8414 + PKCE + device-code flows (`admin/server/auth/mcp_oauth.go`) so Claude/Cursor/ChatGPT can OAuth into Rill Cloud. MCP connect page in web-admin (`web-admin/src/features/ai/mcp`) and "Connect client" popover in chat.

### 1.6 Chat surfaces
`web-common/src/features/chat/`: **DashboardChat** (context-aware side pane on explore/canvas — knows current filters/time range), **DeveloperChat** (Rill Developer workspace), **ProjectChat** (full page at `/-/ai` in cloud). Renders thinking, tool calls, in-chat charts, file diffs; "Explain and fix error" button. Feature flags: `chat`, `dashboard_chat`, `developer_chat`, `chat_charts`.

**AI in embeds:** `GetIFrame` takes `external_user_id` for per-end-user AI chat history inside embedded dashboards; postMessage API can open/close the AI pane.

### 1.7 AI-powered scheduled reports
Reports support `resolver: ai` (`runtime/reconcilers/report.go`, `runtime/resolvers/ai.go`): a scheduled report runs the Analyst Agent with a prompt + explore/filter context per recipient (respecting their security policies) and emails the generated narrative, linking to a shared AI session.

### 1.8 AI scaffolding / code generation
`GenerateMetricsViewFile` (AI with non-AI fallback), `GenerateCanvasFile`, developer-agent `init_project`, AI canvas generation and Vega-Lite chart generation in the frontend.

### 1.9 Instructions for external coding agents
`runtime/ai/instructions/` ships per-resource authoring guides and exports them: `rill init --agent claude|cursor|agentsmd` writes `CLAUDE.md`/`.claude` skills, `.cursor` rules, or `AGENTS.md` + `.mcp.json` pointed at Rill's MCP server.

### 1.10 AI billing, telemetry, evals
Billable metrics per completion (input/output/cached tokens tagged provider + `managed_ai`; tool-call counts). Internal eval/transcript harness (`runtime/ai/evals/`, ~18 recorded scenarios) — **not customer-facing**.

### 1.11 Other
CLI `rill chat`; `ai` resolver usable in any `data:` block (AI output as a data source for reports/components).

## 2. Semantic layer

- **Metrics views** (`runtime/parser/parse_metrics_view.go`, `MetricsViewSpec` proto): bound to exactly one model/table. Dimensions typed categorical/time/geo, expression or column, unnest, URI, lookup dimensions. Measures: simple, **derived** (`per`/`requires`), **time_comparison**, **window** (partition/order/frame); format presets/d3. Security: templated `access`, row filters, per-field include/exclude with conditions, evaluated against JWT user attributes. `ai_instructions`, watermarks, time-grain controls, rollups (pre-aggregated tables with automatic routing), cache key controls, derived metrics views (`parent:` subset).
- **Query engine** (`runtime/metricsview/`): comparisons (value/delta/ratio), percent-of-total, pivots, spine densification, having, search, annotations; AST → dialect SQL per OLAP engine. **Metrics SQL** dialect (JOINs rejected). Rill time DSL. Column/table profiling APIs.
- **Caching**: in-memory LRU result cache keyed by query + state version (`runtime/query.go`). No pre-aggregation *scheduling/matching* layer comparable to Cube's (Rill's answer is the OLAP engine itself + new metrics-view `rollup:`).
- **Explore spec**: field selection, time ranges + comparison offsets, defaults, theme, security.

## 3. APIs

- **Runtime APIs**: gRPC + REST (gRPC-Gateway) — resources, files (+watch), queries (Aggregation, Toplist, Comparison, TimeSeries, Totals, Rows, Search, TimeRanges, Annotations, profiling), git ops (powers cloud editing), AI/conversations.
- **Custom APIs** (`parse_api.go`): YAML endpoints over resolvers (`sql`, `metrics_sql`, `api`, `glob`, `resource_status`, `union`, `text`, `ai`), per-API security, generated **OpenAPI** spec per project.
- **Auth**: admin-issued ephemeral JWTs with granular permissions + user attributes; iframe tokens (`GetIFrame`) with attributes/`external_user_id`/theme/navigation; magic auth tokens (public URLs) with filters + field allowlists.

## 4. Connectors / OLAP

Drivers: athena, azure, bigquery, clickhouse, databricks, druid, duckdb (+MotherDuck, DuckLake), file, gcs, https, mysql, pinot, postgres, redshift, s3, salesforce, slack (notifier), snowflake, sqlite, starrocks, supabase + AI drivers (openai, claude, gemini).
- **OLAP engines**: DuckDB (embedded default), ClickHouse (incl. Rill-managed), Druid, Pinot, StarRocks; warehouse-direct OLAP mode documented for BigQuery/Snowflake/Databricks.
- **Models**: SQL models, incremental + partitioned models (state resolvers, partition retry/skip), staging cross-connector flows, data-quality `tests:`, cron refresh, dev/prod templating.

## 5. Cloud platform

Orgs/projects, org+project roles, user groups, whitelisted domains, access requests, view-as-user; OIDC SSO (Auth0-pattern), PATs, service accounts; billing = Orb + Stripe (usage-based incl. AI tokens); provisioner (static slots, Kubernetes runtimes, managed ClickHouse); GitHub integration + Rill-managed repos; **branch deployments**; **cloud edit mode** (edit-session with commit/branch/merge); custom domains.

## 6. Explicit gaps vs Cube (verified absent in codebase)

- **No SQL API for external BI tools**: no Postgres wire protocol, no JDBC/ODBC. Query surfaces are REST/gRPC, custom APIs, MCP. (Cube's SQL API is its Tableau/PowerBI/Superset/Metabase story.)
- **No GraphQL API.**
- **No pre-aggregation build/match layer**: no equivalent of Cube Store + refresh workers; acceleration is the OLAP engine + in-memory result LRU + metrics-view `rollup:` (early).
- **No Tableau / Power BI / Excel / Looker Studio integrations** (no DAX/MDX/XMLA endpoints, no connector code).
- **No dbt integration** (no manifest/exposure reading).
- **No Kafka/streaming ingestion connector** (Kafka only via ClickHouse per docs).
- **No standalone catalog/discovery product**; discovery = project resource listing + MCP `list_metrics_views`.
- Semantic layer is **YAML-only** (no programmatic Python/JS model definitions like Cube's) and per-project (no cross-project federated semantic graph).

Key files: `runtime/ai/ai.go`, `runtime/ai/mcp.go`, `admin/server/auth/mcp_oauth.go`, `runtime/metricsview/query.go`, `runtime/parser/parse_metrics_view.go`, `runtime/parser/parse_api.go`, `runtime/reconcilers/report.go`, `runtime/resolvers/ai.go`, `proto/rill/runtime/v1/api.proto`, `admin/billing/`, `admin/provisioner/`.
