# Cube — Core Platform Inventory (semantic layer, APIs, caching, cloud)

> Researched Aug 2026 from cube.dev blog (dated post index), docs.cube.dev (llms.txt + pages), cube.dev/pricing, GitHub, press, HN/G2. Companion file `cube-ai-and-market.md` covers the AI platform and market landscape in depth.

## 0. Positioning snapshot (mid-2026)

- Homepage: **"The agentic analytics platform built on a semantic layer."** GitHub: "Cube Core is open-source semantic layer for AI, BI and embedded analytics."
- 2025–2026 pivot: "universal/headless semantic layer" → **full agentic BI platform** with a first-party frontend (chat, workbooks, dashboards), with open-source **Cube Core** as the foundation. Definitive framing: CEO post "The Future of Cube Core and Cube" (May 14, 2026). "D3" brand introduced June 2025, retired Oct 2025.
- Cube Dev, Inc.: founded 2019 by Artyom Keydunov (CEO) + Pavel Tiunov (CTO), spun out of Statsbot.

## 1. Semantic layer / modeling

- Models in **YAML or JavaScript**; YAML supports **Jinja + Python**; JS runs in a Node VM. Git-managed (branches, dev mode, CI).
- **Cubes** (tables/SQL queries; measures, dimensions, joins, segments, pre-aggregations, access policies) + **views** (curated facades with explicit join paths — what consumers and AI agents query). **View groups** (v1.7, Jul 2026). **Hierarchies and folders** (v1.2, Feb 2025), nested folders (v1.5).
- **Joins**: many_to_one/one_to_many/one_to_one; automatic join-path resolution (Dijkstra over the join graph); views pin explicit paths; strict validation since v1.6.
- **Multi-fact views**: cross-fact analysis on shared dimensions without fan-out (Tesseract-powered, GA v1.7).
- **Tesseract**: Rust rewrite of the query/modeling engine, announced Oct 2024, **GA and default in Core v1.7 (Jul 8, 2026)**. Powers **multi-stage calculations** (period-over-period, percent-of-total, running aggregates, YTD/QTD), `grain`/`filter` directives, `switch` dimensions, `case` measures, XIRR, conditional data masking, number formatting, links in the model.
- **Custom granularities** (v0.36, Sept 2024); **calendar cubes** (custom 4-5-4 retail/fiscal calendars, time shifts — v1.5, Oct 2025).
- Segments, drill members (`drill_members` + `ResultSet.drillDown()`), proxy/subquery dimensions, `extends` inheritance, polymorphic cubes, data blending.
- **Dynamic data models**: programmatic generation via Jinja/Python or JS (`COMPILE_CONTEXT`, `FILTER_PARAMS`, per-tenant models/data sources).
- **Data Access Policies** (row-level, member-level, **member masking**) — GA Feb 2025, in Core; user attributes, groups, `context_to_groups`; policy viewer UI in Cloud.
- Migration tooling: **`lkml2cube`** (LookML→Cube), **`cube_dbt`**; sustained SSAS-displacement campaign; **dbt integration** shipped Jul 16, 2026 (one-way dbt→Cube with CI/webhook sync + review branches; Snowflake/Redshift/Postgres/BigQuery).

## 2. APIs & integrations

- **SQL API ("Semantic SQL")** — **Postgres-wire-compatible**; extends SQL with `MEASURE()`; any Postgres client/BI tool connects; query pushdown GA; HTTP transport (`/v1/cubesql`, NDJSON streaming) since v1.5; window functions (v1.7). Declared "**the primary Cube Core API going forward**" (May 2026) — JSON REST API de-prioritized.
- **REST (JSON) API** (`/v1/load`, `/v1/meta`, WebSocket subscriptions), **GraphQL API**.
- **DAX API** — native **Power BI** connectivity (Feb 2025, Enterprise). **MDX API** — native **Excel pivot tables** (Jul 2024, Cloud). **Cube Cloud for Excel** add-in (Feb 2025); **Cube Cloud for Sheets** (Sept 2024).
- **Semantic Layer Sync** — pushes the model into BI tools: Superset/Preset + Metabase (paid plans), Tableau (Enterprise).
- **Orchestration API** — pre-aggregation refresh from Airflow (`airflow-provider-cube`), Dagster, Prefect.
- **Control Plane / Platform API** — REST automation of deployments, env vars, tokens, model files/branches, dbt sync, workbooks/reports/**dashboards-as-code**, users/groups/policies, SCIM 2.0, OIDC, embed sessions/tenants; OpenAPI published.
- **Chat API (Embed API)** — streaming conversational-agent API; supports agent-to-agent wrapping.
- **AI API** (May 2024; value search + BYOLLM + Slack app Mar 2025; Snowflake Cortex LLM Apr 2025).
- **MCP server** — hosted `https://cubecloud.dev/mcp` (OAuth + PKCE); Premium/Enterprise; role-scoped tools; **official Claude Connectors Directory listing (Aug 20, 2026)**.
- **Rust CLI** (`cube`, Jul 21, 2026, open source): OAuth device flow, `--json` everywhere, covers the entire REST API — explicitly designed as an **agent interface**.
- JS SDKs (`@cubejs-client/core`, react, vue3, ngx, ws). No native JDBC/ODBC — Postgres wire compat is the interop path.
- Accumulated tool integrations: Tableau, Power BI, Excel, Sheets, Superset/Preset, Metabase, Hex, Retool, Streamlit, Observable, Deepnote, Klipfolio, Hashboard, Embeddable, Budibase, Appsmith, Bubble.

## 3. Caching & performance

- **Two-level caching**: in-memory result cache + **pre-aggregations** declared in the model, background-refreshed, matched via **aggregate awareness**.
- **Cube Store**: purpose-built distributed OLAP store for pre-aggregations (Rust; Parquet + Arrow + DataFusion; router/worker architecture; sub-second, high concurrency). v1.6 (Dec 2025) upgraded DataFusion 4→46: DECIMAL(38), CTE planning, **pre-aggregation acceleration of multi-stage calculations**.
- **Refresh workers** (scheduled `refresh_key`, partitioning, incremental builds); **lambda pre-aggregations** (merge historical rollups with fresh/streaming data for near-real-time).
- v1.7: pre-aggs with multiple time dimensions, per-subquery matching, pre-agg-specific data sources; columnar JS↔Rust transport (~5–20x faster result handling). v1.5: 2–3x model-compilation speedup OSS ("up to 30x" in Cloud).
- Cloud extras: export buckets (S3/GCS), customer-managed keys for Cube Store, BYO pre-agg storage bucket, deployment warm-up.

## 4. Data sources

Vendor-supported: Snowflake, BigQuery, Databricks (OAuth service principals), Redshift, Microsoft Fabric, ClickHouse, SingleStore, Pinot, Firebolt, Vertica; Athena, DuckDB/MotherDuck, Presto, Trino; Postgres, SQL Server, MySQL, Oracle (community), SQLite (community); **QuestDB, ksqlDB, Materialize, RisingWave** (streaming/real-time); MongoDB (BI Connector), Druid (community); Parquet/CSV/JSON via DuckDB. Community drivers: ArangoDB, CosmosDB, CrateDB, Dremio, SAP Hana, etc. Multiple simultaneous data sources per deployment; per-user warehouse OAuth passthrough; cross-database rollup joins.

## 5. Cube Cloud platform

- **Dev workflow**: browser Data Model IDE (branch dev mode, branch-scoped API endpoints, staging envs), **Visual Modeler** (no-code, Mar 2025), Playground 2.0 (chart prototyping), **Content Validator**, git integrations (GitHub app, GitLab/Bitbucket/SSH), continuous deployment.
- **Deployment types**: Shared (free, dev-only, auto-suspend, 1,000 queries/day, no refresh workers) · Dedicated (2 API instances + refresh worker + 2 Cube Store routers, autoscaling) · **Multi-cluster** (Premium+; up to 10 deployments/100 API instances routed by `context_to_app_id` for >100-tenant multi-tenancy).
- **Infra options**: SaaS; single-tenant dedicated with VPC peering / PrivateLink / Private Service Connect; **BYOC on AWS, Azure, GCP** (data plane in customer account); OIDC workload identity (keyless federation incl. Bedrock).
- **Observability**: Usage Analytics (incl. AI usage), Query History (1-day Free → 7-day Premium → custom Enterprise; export to S3), pre-agg console, Performance Insights, **Audit Log** (Enterprise), **Chats History** (admin view of AI chats), Datadog/Grafana/New Relic/CloudWatch integrations.
- **Auth**: SAML SSO (Google/Entra/Okta) + **SCIM** (Enterprise), LDAP (Feb 2025), roles Admin/Developer/Explorer/Viewer + **custom roles** (Enterprise), user groups/attributes, data access policies, JWT security context for API consumers. SOC 2.
- **Pricing** (Oct 2025 model): Free (5 workbooks, 1k queries/day, limited agent requests) · Starter $40/dev/mo (Dedicated $0.60/hr) · Premium $80/dev/mo + Explorer $40 + Viewer $20 seats, embedded dashboards/chat, 99.95% SLA, multi-cluster $1.20/hr · Enterprise custom (99.99% SLA, BYOC, BYOLLM, SSO/SCIM, custom roles, audit, DAX API, dbt integration). Embedded stays **compute-priced**. AI tokens: per-seat grants + token packages.

## 6. Embedded analytics

Four embedding options (expanded as **Embedded Agentic Analytics**, June 11, 2026):
1. **Analytics Chat API** (custom chat UX; agent-to-agent capable)
2. **Chat & Dashboard iframes** (theming, SSO, private + signed auth, postMessage events, localization, per-embed time zones)
3. **Creator Mode** (end customers build their own workbooks/dashboards in-product)
4. **Core Data APIs** (classic headless: JSON/GraphQL/SQL)

**React Embed SDK** in public preview. White-labeling: app theming (colors/logo/fonts), chart palettes, dashboard themes. Native viz layer: bar/line/area/scatter/pie/heatmap/boxplot/table (pivots, conditional formatting)/KPI (progress, sparklines, rich text)/**maps (point + choropleth)**/custom HTML (Handlebars+JS)/raw Vega-Lite; small multiples, dual axes, reference lines. Flagship embedded customers: **Brex "Spaces"**, Webflow, Deel; claims "thousands of organizations."

## 7. Open source vs commercial

- **Cube Core** (github.com/cube-js/cube): **20.7k stars**, ~1.1k open issues, actively developed. Dual **Apache 2.0/MIT** per package. Includes Cube Store, all drivers, Rust CLI, cubesql. LTS program since Oct 2024.
- **In Core**: full modeling (Tesseract, hierarchies/folders, calendar cubes, granularities), access policies, pre-aggregations + Cube Store, SQL/REST/GraphQL APIs, Playground, multi-tenancy primitives, all drivers.
- **Cloud-only**: the entire agentic frontend (chat, workbooks, dashboards, agents, skills, scheduled tasks, evals, Python analysis, MCP server + connectors, Slack agent), Semantic Catalog, Visual Modeler, IDE, Semantic Layer Sync, MDX/DAX, Excel/Sheets add-ins, embed iframe/SDK/Creator Mode, SSO/SCIM/custom roles/audit, managed infra/BYOC.
- HN commentary (2025): "a lot of the work being done has been for the cloud service so some parts of the OSS version are starting to fall behind."

## 8. Company & market

- Funding ~**$49M** total: seed (~$6.5M, 2019) → Series A $15.5M (2021, Decibel) → **Series B $25M (June 2024, Databricks Ventures)**. No 2025–2026 round found. **Narrative BI acqui-hire Sept 2025** (founder Michael Rumiantsau → VP Product). Headcount ~72 (2026, third-party estimate); ~$7.9M ARR (Getlatka 2025 estimate, low confidence).
- Scale claims (own PR): 90k server installs, 5M+ users, 20% of Fortune 1000.
- Analyst recognition: #1 Dresner 2025 Semantic Layer study; Leader & Outperformer, GigaOm Radar 2025; **Gartner 2026 Market Guide for Agentic Analytics**; Snowflake OSI founding participant.
- Positions against: dbt Semantic Layer/MetricFlow, AtScale, Looker/LookML ("Open-Source Looker Alternative for the AI Era," Jul 2026), Microsoft SSAS, legacy semantic layers.
- Common criticisms (HN/G2/StackShare): caching complexity in multi-tenant setups; generated SQL "could be better optimized"; docs gaps; OSS lagging Cloud; upgrade regressions; BI integrations "not totally straightforward"; steep initial learning curve; opt-out telemetry. Praise: best-in-class pre-aggregation/aggregate awareness, saves building an analytics API, strong embedded multi-tenant story.

## 9. Ship cadence mid-2025 → mid-2026 (chronological)

D3 launch (Jun 2, 2025) → Narrative BI team (Sep 9) → OSI (Sep 23) → native AI/BI frontend manifesto (Oct 1) → **Agentic Analytics GA + seat pricing** (Oct 29) → Core v1.5 (calendar cubes, SQL-over-HTTP, nested folders) (Oct 29) → AI Widget + dashboard iframes + map/KPI/custom-HTML charts (Nov 14) → **Core v1.6** (DataFusion-46 Cube Store, multi-stage pre-aggs, pushdown GA) (Dec 18) → AI-assisted onboarding + stateless Explore + prompt suggestions (Jan 12, 2026) → Gartner recognition (Mar 3) → LLM reliability benchmark (Apr 28) → **Slack Agent** (May 7) → Core/Cloud strategy (May 14) → **Agent Skills** (Jun 2) → **Embedded Agentic Analytics** (Jun 11) → **MCP Connectors** (outbound; Notion/Linear/Sentry/Attio) (Jun 15) → **Cube Evals + AI Studio** (Jun 25) → **Core v1.7: Tesseract GA** (Jul 8) → **Scheduled Tasks** (Jul 9) → **dbt integration** (Jul 16) → **Rust CLI** (Jul 21) → **Python analysis** (Jul 30) → agentic-loops roadmap (Aug 13) → **Claude Connector + open-source `cube-agent-skills`** (9 Claude Code skills, Apache 2.0) (Aug 20). Roughly one substantive launch every two weeks, nearly all AI/agent-centric.

Announced-not-shipped (Aug 2026 roadmap): agent session environments (sandboxed parallel agents proposing merges), model unit tests/artifact checks, event-triggered background agent loops (dbt changes, dashboard errors, query-history signals), Cube→dbt write-back, in-channel Slack mentions, model-graded evals.

## 10. AI agent configuration surface (docs, mid-2026)

Code-first config in `agents/` in the model repo (default since Apr 30, 2026; UI config deprecated):
- `config.yml`: `llm` (default `auto`; pinnable — list includes `claude_4_6_opus`…`claude_5_opus`, `gpt_5_4`, `o3`), `embedding_llm`, `runtime: plain|reasoning`, `accessible_views`, `memory_mode: space|user|disabled`.
- **Rules** (markdown instructions: always + agent-requested), **Certified Queries** (trusted SQL example library), **Skills** (`agents/skills/*.md`, frontmatter + instructions; run via chat buttons, slash menu, or auto-matching; versioned/reviewed like code), **Memories** (tenant/space/user isolation), **Multi-agent** (multiple agents per deployment, each with own model/views/rules), **BYOM** (Enterprise), **MCP Connectors**, **Evals** (`agents/eval_questions.yml`: NL question + ground truth SQL/certified query; sort-invariant, numeric-tolerant result comparison; per-branch runs; CI-gateable; surfaced in AI Studio).
- **Scheduled Tasks** (NL prompt + cadence; runs headless under creator's security context; can query, search web, build/update reports & dashboards, email members; cannot yet edit the model).
- **Python analysis** (sandboxed Python attached to workbook tabs; input = tab's Semantic SQL result; stats/ML libs preinstalled; agent auto-invokes for forecasting/regression/clustering/anomaly asks).
