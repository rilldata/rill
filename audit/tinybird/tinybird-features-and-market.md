# Tinybird — Platform, AI & Market Inventory

> Researched Aug 2026 from tinybird.co (site/docs/changelog/blog), HN, funding databases. Founded 2019 Madrid (Jorge Sancha CEO, Javier Santana — ex-CARTO), HQ now NYC. Tagline: **"Ship faster with a Managed ClickHouse®"**; mid-2025 self-description "the Supabase for ClickHouse." Runs **its own ClickHouse fork** (backup engine, multi-writer ingestion, compute-compute separation).

## 1. Core product

- **Ingestion**: **Events API** (streaming HTTP `POST /v0/events`, 1,000 req/s + 20 MB/s per workspace); native **Kafka connector** (Schema Registry; Confluent/MSK/Redpanda); S3/GCS batch; **DynamoDB CDC** (Jul 2026); Postgres/MySQL/URL table functions; **Iceberg reads**; no first-party Postgres CDC (Redpanda Connect/Estuary/Sequin patterns); ~50 guided recipes (Stripe, Clerk, Sentry, Vercel log drains, LLM-call instrumentation…); CSV/NDJSON/Parquet + quarantine.
- **Pipes**: chained SQL nodes in `.pipe` files → published as **API Endpoints** (the core product: parameterized, authenticated REST returning JSON/CSV/NDJSON/Parquet/Prometheus; auto OpenAPI docs; per-endpoint observability via queryable Service Data Sources), **Materialized Views** (incremental at ingest), **Copy Pipes** (scheduled), **Sinks** (Kafka/S3/GCS export, Jun 2025).
- **Templating**: Jinja-flavored typed params (`{{ String(x, 'default') }}`), control flow, pagination helpers.
- **Query API** for ad-hoc SQL; **ClickHouse-native HTTP interface** (Sep 2025) so standard CH clients/ORMs/BI tools connect read-only.
- **Dev workflow (analytics-as-code)**: `.datasource`/`.pipe`/`.connection` datafiles in git; **Tinybird Local** (full runtime in Docker, ~150ms build loop); `tb deploy --check` reviewable deployment plans with automatic schema migrations/backfills; **Cloud Branches** (zero-copy envs sharing prod data, **GA Feb 24, 2026** — "data infrastructure for agents"); **TypeScript SDK** (Feb 2026) + **Python SDK** (Mar 2026); CLI 4.0 (Mar 2026).

## 2. 2025–26 evolution

- **Tinybird Forward** (announced Mar 14, 2025): the relaunch — new CLI, local-first Docker runtime, deployments with auto-migration, datafiles as sole source of truth, AI scaffolding (`tb create --prompt`), `tb mock`, LLM-assisted `tb test`, `tb infra` (self-managed regions). CLI v1.0 Nov 2025. **Classic sunset**: free/developer workspaces must migrate by **Sept 15, 2026**; "80% of users already on Forward"; Enterprise exempt for now.
- **AI agent arc — build → sunset**: `tb create --prompt` "vibe data engineering" (Mar 2025) → **Explorations** chat UI (May 2025) + Birdwatcher Slack bot + public **LLM-SQL benchmark** (llm-benchmark.tinybird.live) → **Tinybird Code**, a full agentic CLI ("How we built our own Claude Code," Aug 2025) → **sunset Feb 20, 2026** ("We built our own AI coding agent. Here's why we're sunsetting it"): Claude Code/Cursor/Codex matured; platforms should ship **primitives** — clear CLI errors, typed SDKs, docs, **Agent Skills** (`tinybirdco/tinybird-agent-skills`), MCP — not custom agents. Building eval frameworks to measure external-agent performance against Tinybird.
- **MCP**: very early adopter (local `mcp-tinybird` Nov–Dec 2024); **hosted remote MCP server Jun 23, 2025** — every workspace is an MCP server (`mcp.tinybird.co`, Streamable HTTP); exposes published endpoints as tools + query/exploration tools; **scoped by the same tokens/JWTs as the APIs** so agents inherit row-level security ("Don't trust the prompt: use RLAC to secure LLM database access"). "Analytics for AI agents" positioning: docs section, templates, LLM usage credits in every plan.
- **Managed-ClickHouse repositioning**: pivoted marketing to "Managed ClickHouse®" with aggressive comparison SEO vs ClickHouse Cloud; every page carries a trademark disclaimer (ClickHouse® of ClickHouse, Inc.) — proactive compliance amid open tension. Deployment: shared infra, dedicated clusters (Enterprise), **self-managed regions** (their control plane on your infra); no bring-your-own-existing-cluster (they run their fork).
- **Ship cadence mid-2025→mid-2026**: Sinks, Windows CLI, MCP (Jun 2025) → self-managed regions (Aug) → CH HTTP interface (Sep) → compute-compute separation (Oct) → CLI 1.0, branches beta (Nov) → **Playgrounds** (browser SQL) + **Time Series** viz + TS SDK + Branches GA (Jan–Feb 2026) → CLI 4.0, Python SDK, new UI, logs explorer (Mar) → Hong Kong + Sydney regions (Apr) → vector-search 48s→130ms, QPS-limit removal (May) → lightweight deletes beta (Jun) → DynamoDB connector, Classic deprecation (Jul) → org-wide observability (Aug).

## 3. Consumption surfaces

**API-first; no BI/dashboard product.** In-product operator tools only: Playgrounds, **Time Series** (config-driven charting, Feb 2026), Explorations (AI chat), observability. `@tinybirdco/charts` React/ECharts library (beta, Classic-era). Official **Grafana plugin** + 3 documented Grafana paths. Since the HTTP interface: **BI connection guides** for Power BI, Tableau, Metabase, Superset, Redash, Hex, Deepnote, Holistics, Luzmo, DataGrip… **Templates**: Web Analytics Starter Kit (flagship), real-time dashboards (Next.js + Tremor), LLM Performance Tracker, logs explorer, Clerk multi-tenant JWT, Vercel Marketplace integration. Use-case pages: "User-facing dashboards" — "expose real-time data as REST APIs **for your app to visualize**."

## 4. Enterprise / governance

- **Tokens**: static scoped tokens + **JWTs** (customer-signed, not stored by Tinybird) carrying **row-level filters per tenant** — the canonical multi-tenant pattern; same tokens gate MCP. Clerk template.
- Multi-tenancy: shared-table + JWT-RLS (default), workspace-per-tenant, or dedicated infra.
- **SOC 2 Type II**; SSO/SAML Enterprise; MFA; encryption; GDPR deletion guide; trust.tinybird.co. Regions: AWS/GCP US + EU + new Hong Kong/Sydney; one region per org.
- Deployment: shared, dedicated clusters (replica management, traffic weights), self-managed regions.

## 5. Pricing (mid-2026)

History: processed-GB usage pricing (unpredictability complaints) → Jan 2025 plan/vCPU model → **May 18, 2026: QPS limits removed** ("QPS limits are the worst. We agree").

| Plan | Price | Included |
|---|---|---|
| Free | $0 forever | 0.25 vCPU (3× burst), 1,000 API req/day, 10 GB, $25/mo LLM credits |
| **Developer** | **$25–$799/mo by vCPU size (S-1/4 → S-8)**, self-serve, no commitment | 25 GB storage, $100/mo LLM credits; overage $0.0002/vCPU-s, $0.058/GB; ingestion never throttled |
| SaaS | Credit-based, 12-mo min, ≥4 vCPU, ≥1 TB | $200/mo LLM credits; mandatory support |
| Enterprise | Dedicated cluster, credit-based, 12-mo min | Cluster CPUs, ingestion, workers, storage SKUs; self-serve replicas |

Contrast they push: permanent free tier vs ClickHouse Cloud's $300/30-day credits. Egress charged only on Sinks.

## 6. Company & market

- **~$70M raised**: Series A $37M (Apr 2022, CRV) + Series B $30M (Jun 2024, Balderton) at reported **$240M**. No 2025–26 round found.
- **Customers**: **Vercel** (billions of events/day: web analytics, WAF, usage billing), **Canva**, **FanDuel**, The Hotels Network (PB/month), Framer, Dub, Phantom, **Ghost** (Tinybird is Ghost 6.0's built-in analytics backend), Linux Foundation (700M+ events), Fever, Factorial.
- **vs ClickHouse Cloud** (existential): CH Inc. raised **$400M Series D at ~$15B (Jan 2026)**. Tinybird's differentiation: higher abstraction (no cluster ops), the endpoint layer + token/JWT RLS vs build-your-own API tier, Events API, local-first dev + branches + CI/CD + SDKs, plan pricing vs compute-hour billing, permanent free tier, per-endpoint observability. Concedes direct DB control (mitigated by HTTP interface). Both offer hosted MCP.
- Other set: Rockset (acquired by OpenAI 2024), StarTree, Firebolt, Materialize, Propel, DIY ClickHouse/Altinity.
- Sentiment: strong practitioner praise (speed, SQL workflow, API publishing); criticisms: customization limits, historic pricing opacity (both since engineered away and marketed as such), SaaS-only pushback (partly answered by self-managed regions). Small G2 footprint; excellent engineering blog; steady HN presence.

## 7. Overlap with BI/embedded deals

Primary marketed use case = **user-facing analytics built by developers**: sub-second parameterized APIs + JWT RLS + your own React components instead of buying embedded BI. Where customers want BI, Tinybird routes to the HTTP interface + third-party BI guides — positioning as **the serving database under BI, not the BI**. Notable surface gaps: **no semantic/metrics layer, no hosted dashboards, no business-user drill UI** — consumption is code, agents (MCP), or third-party BI. In AI deals: workspace-as-MCP-server + token RLS = "the analytics backend for agents," with generative-UI recipes replacing dashboards.
