# Rill Feature Inventory (current state, QuickSight-focused)

> Basis for the QuickSight vs Rill comparison. Derived from the codebase (this repo, Aug 2026) and docs.rilldata.com. Emphasis on the surfaces where QuickSight competes: AWS-native data access, embedded analytics economics, reporting/delivery, governance, and AI. Full inventories: [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) (BI surface detail) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md) (AI subsystem + semantic layer detail) — both still current.

## 1. AWS-relevant data plane

- **AWS-native connectors** (`runtime/drivers/`): **S3**, **Athena**, **Redshift**, plus Postgres/MySQL (covers RDS/Aurora), Snowflake, BigQuery, Databricks, GCS, Azure Blob, Salesforce, HTTPS, local files, DuckLake/Iceberg-adjacent via DuckDB. No Kafka/streaming connector; no SaaS-app connector library (no Jira/ServiceNow/Adobe-style sources).
- **Engine model (the SPICE analog, done differently)**: data lands in a real OLAP engine — embedded **DuckDB** (default), **ClickHouse** (incl. Rill-managed), Druid, Pinot, StarRocks, MotherDuck — rather than a proprietary in-memory cache. No per-GB "capacity" purchasing; managed compute billed by usage (CU-hr) with plan-tier storage limits. Warehouse-direct OLAP mode exists for BigQuery/Snowflake/Databricks (secondary path).
- **Transformation is first-class** (QuickSight's data prep is thin): SQL models, **incremental + partitioned models** with state resolvers and partition retry/skip, staging cross-connector pipelines (e.g. Snowflake→S3→ClickHouse), data-quality `tests:`, cron refresh, dev/prod templating with partial data in dev.
- Deployment: single Go binary locally; Rill Cloud is its own SaaS (not in the customer's AWS account; VPC peering at Enterprise). **No AWS-marketplace/IAM-native integration**: auth is Rill Cloud's own (OIDC/Auth0), not IAM Identity Center.

## 2. Semantic layer & governance

- **Metrics views**: one model/table each; typed dimensions (categorical/time/geo), lookup dimensions, unnest; measures simple/derived/window/time-comparison; format presets; `ai_instructions`. Governance: templated `access`, **row-level filters**, **field-level include/exclude** — all evaluated per user attribute from JWTs; mock users for testing; view-as-user in cloud. Derived metrics views; rollups (early); cache controls.
- **Explicitly absent** (relevant vs QuickSight): joins/multi-fact across metrics views; fiscal calendars; column masking mode; dataset-level "folders" metadata.
- **Governance-as-code**: every dashboard/metric/policy is YAML in git with branch deployments and cloud edit sessions — the whole-project reconciler validates continuously. (QuickSight's analog is JSON asset definitions via API — no native git.)

## 3. Dashboards, reporting, delivery

- **Explore** (auto-generated slice-and-dice: leaderboards, TDD, pivot with conditional formatting, comparisons, URL-encoded state) + **Canvas** (composed: 11 chart types over 6 renderers, KPI grid, tables, dynamic markdown, Vega-Lite escape hatch, tab groups, per-widget filters).
- **No native**: maps/geospatial, boxplot, waterfall, gauge, radar, histogram widget, small multiples; no cross-widget click-to-filter; no viewer-side calculated fields; no mobile apps; no offline.
- **Delivery**: scheduled reports (CSV/XLSX/Parquet export, email + Slack), **AI-narrative reports** (`resolver: ai`, per-recipient security), threshold **alerts** (criteria AND/OR, comparison periods, per-user RLS evaluation, snooze/renotify, email/Slack), canvas **PDF export**. **No paginated/pixel-perfect reports**, no PNG render service, no bursting, no Sheets/S3/SFTP delivery targets.
- Sharing: bookmarks (personal/shared/default), **public URLs** (magic tokens with locked filters/field restrictions/TTL), custom domains, themes incl. dark mode.

## 4. Embedded analytics

- iframe embed of explore/canvas/component or whole project; backend-issued embed tokens with `user_email`/arbitrary attributes (**row-level security enforced server-side per end user**) + `external_user_id` for per-user state and **per-user embedded AI chat history**; navigation control; theme + theme-mode overrides; **postMessage API** (state get/set, navigation, AI pane control). No React SDK, no anonymous-embedding tag-based RLS shortcut (attributes serve that role), no embed "creator mode," no localization.
- **Economics**: flat **$500/mo for up to 100 active embedded seats** (Growth), unlimited at Enterprise — no per-session metering.

## 5. AI (summary; full detail in `../cube/rill-feature-inventory.md` §1)

- Multi-agent runtime (`runtime/ai/`): router → **analyst** (data Q&A, dashboard-context-aware) / **developer** (writes project files, scaffolds whole projects) / **feedback** (failure attribution) agents; ~19 tools incl. `query_metrics_view`, `query_sql`, `create_chart`, file write, bucket introspection.
- Chat surfaces: dashboard side-pane (knows filters/time range), project chat, developer chat, CLI `rill chat`; conversation persistence/sharing/forking; in-chat charts; citations = live dashboard URLs.
- **MCP server** per project (read tools for viewers; **write tools** with edit access) incl. free local product; full OAuth (PKCE/device-code) for Claude/Cursor/ChatGPT; `rill init --agent claude|cursor|agentsmd` exports agent instructions + `.mcp.json`.
- AI reports (scheduled analyst narratives, per-recipient RLS); AI anomaly "Explain" (root-cause on metric shifts); AI scaffolding (`GenerateMetricsViewFile`, canvas + Vega-Lite generation); BYO LLM (`ai_connector`: OpenAI/Azure, Claude, Gemini) from Growth tier; AI-token metering in billing; internal eval harness (not customer-facing).
- Gaps (vs the market, incl. QuickSight's agentic suite): no customer-facing evals, no user-defined skills, no Slack agent, no outbound tool connectors, no forecasting/what-if engine, no automation/workflow agent.

## 6. Cloud platform & enterprise

- Orgs/projects; org + project roles; user groups; whitelisted email domains; access requests; PATs + service accounts; OIDC SSO (SAML via Auth0 enterprise connections); billing Orb+Stripe (seats + usage incl. AI tokens); provisioner (k8s runtimes, managed ClickHouse); GitHub integration + direct upload; branch deployments; cloud edit sessions.
- **Absent**: SCIM, audit logs, custom roles, customer-facing usage/query-history analytics, BYOC, IAM-native auth, compliance surface beyond SOC 2-level basics (no HIPAA/FedRAMP story).

## 7. Pricing (public, Aug 2026)

Starter **$20/seat/mo** (≤20 seats, 3 projects, 1M AI tokens/seat, ≤10GB, $0.15/CU-hr) · Growth **$30/seat/mo** (≤100 seats, 10 projects, embedded $500/mo per 100 seats, 2M tokens/seat + 200k/embedded seat, BYO DB + BYO LLM, ≤1TB, $0.20/CU-hr) · Enterprise custom (unlimited embedded, VPC peering). Open source: the **entire product** (local Developer incl. dashboards, AI chat with BYO key, MCP) is free and self-hostable.
