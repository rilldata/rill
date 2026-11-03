# Tinybird vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill parity or better · 🟡 partial/different · ❌ absent · ➕ Rill advantage. Detail in `tinybird-features-and-market.md` and `rill-feature-inventory.md`.

**Framing:** The most architecturally-kindred rival in the audit set: both are ClickHouse-adjacent, local-first, datafiles-in-git, MCP-native, real-time-analytics platforms with token/JWT row-level security. The fork in the road is the consumption layer: Tinybird sells **APIs for developers to build UIs on**; Rill sells **the finished BI + metrics layer**. The buyer decision — "build on Tinybird APIs" vs "ship Rill's product" — plus the shared "analytics for AI agents" positioning make this a real head-to-head on Rill's own wedge.

## 1. Architecture & data plane

| Feature | Tinybird | Rill status | Notes |
|---|---|---|---|
| Engine | Managed ClickHouse (own fork; dedicated clusters; self-managed regions) | ✅🟡 | Rill: DuckDB default + managed ClickHouse + **BYO ClickHouse/MotherDuck zero-copy** (Tinybird has no BYO-cluster) + Druid/Pinot/StarRocks. Tinybird deeper on CH ops (compute-compute separation, replicas) |
| Streaming ingestion | **Events API** (1k req/s HTTP streaming) + native **Kafka connector** + DynamoDB CDC | ❌ | **Rill's biggest gap in this fight** — no Kafka, no streaming endpoint; micro-batch models only |
| Batch ingestion | S3/GCS, table functions, Iceberg reads | ✅ | Rill broader on warehouses (Snowflake/BQ/Databricks/Redshift/Salesforce) |
| Transformations | Pipes + **incremental Materialized Views at ingest** + Copy pipes + Sinks | ✅🟡 | Rill incremental/partitioned models are batch-shaped; MV-at-ingest is better for streaming; Rill has tests, staging flows, cross-connector outputs Tinybird lacks |
| Local-first dev | `tb local` Docker, ~150ms builds, `tb deploy --check` plans, auto-migrations | ✅ | `rill start` single binary (no Docker needed) + reconciler; Tinybird's deployment-plan UX (reviewable migrations/backfills) is ahead |
| Branches | **Zero-copy Cloud Branches GA** (share prod data) | 🟡 | Rill branch deployments re-ingest; zero-copy branching is a superior pattern to copy |
| SDKs | TypeScript + Python SDKs define resources in code | ❌ | Rill is YAML-only (no programmatic SDK — same note vs Cube) |

## 2. Serving & APIs

| Feature | Tinybird | Rill status | Notes |
|---|---|---|---|
| SQL → REST endpoints | **The core product**: parameterized, typed templating, per-endpoint OpenAPI docs + observability, multiple formats incl. Prometheus | 🟡 | Rill custom APIs (resolvers + OpenAPI + security) do the job but lack the polish: no per-endpoint metrics UI, no typed-param DX, no Prometheus format |
| Token security | Static scoped tokens + **customer-signed JWTs with per-tenant row filters** | ✅ | Rill JWT attributes + security policies = parity-or-better (policy layer is richer than raw row filters) |
| DB wire access | ClickHouse HTTP interface → any CH client/BI tool | 🟡 | Rill has no SQL wire access (♻ Postgres-wire, cube §3.1 — over *metrics*, a governed upgrade on Tinybird's raw access) |
| Endpoint observability | Latency/error/QPS per endpoint, queryable service data sources, org-wide observability | ❌ | Rill has no per-API usage analytics surface (♻ query-history/usage, qs) |
| Semantic/metrics layer | **None** — raw SQL endpoints | ➕➕ | Rill's defining advantage here: governed metrics for humans *and* agents |

## 3. Consumption & BI

| Feature | Tinybird | Rill status | Notes |
|---|---|---|---|
| Dashboards/BI | None (Time Series/Playgrounds are operator tools; `@tinybirdco/charts` beta; routes buyers to Grafana/Metabase/etc.) | ➕➕ | Rill's whole explore/canvas/alerts/reports/embed product — "Tinybird gives you APIs; Rill gives you the product" |
| Embedded analytics | DIY: your React app on their APIs (Next.js templates, Vercel integration) | ➕ | Rill: finished embeds + per-user RLS + AI chat + flat pricing. But dev teams who *want* full UI control choose Tinybird's shape — Rill's counter is custom APIs + components |
| Templates/starter kits | Web Analytics Starter Kit, dashboards, LLM tracker, logs explorer | ❌ | Rill has example projects, no productized starter-kit motion |
| Alerts/reports | None (build on APIs) | ➕ | |

## 4. AI & agents

| Feature | Tinybird | Rill status | Notes |
|---|---|---|---|
| MCP server | Hosted, every workspace (Jun 2025); endpoints-as-tools; token-scoped RLS ("RLAC") | ✅➕ | Parity on concept; Rill's tools are *governed metrics* not raw SQL endpoints — better grounding; Rill adds write tools |
| Agent-facing primitives | **Agent Skills repo**, typed SDKs, clear CLI errors, evals-in-progress; sunset their own agent ("don't build agents, build for agents") | ✅🟡 | Rill ships both primitives (`rill init --agent`, instructions) *and* its own agents. Tinybird's public skills repo + LLM-SQL benchmark are marketing moves Rill should mirror (♻ skills repo, benchmark) |
| In-product AI | Explorations chat (operator-oriented) | ➕ | Rill's analyst/developer agents + dashboard chat are a class above |
| AI metering | LLM credits in every plan ($25–200/mo) | ✅ | Similar shape (Rill tokens/seat) |
| "Analytics for AI agents" positioning | Loud and early (Nov 2024 MCP) | 🟡 | Both claim it; Tinybird's developer mindshare on this narrative is strong — contest it |

## 5. Pricing & company

| Dimension | Tinybird | Rill | Notes |
|---|---|---|---|
| Model | Pure usage (vCPU + storage), **no seats**; free forever tier; Developer $25–799/mo self-serve | Seats + usage | Tinybird wins seatless API workloads on shape; Rill wins when consumption seats deliver the UI |
| Funding/scale | ~$70M, $240M valuation (2024), no round since; strong logos (Vercel, Canva, FanDuel, Ghost) | ~$14M | Both small vs ClickHouse Inc. ($15B) — which pressures Tinybird existentially more than Rill (Rill is multi-engine + BYO) |
| Momentum | Weekly changelog, disciplined AI strategy pivot, 2 new APAC regions | — | High execution quality; respect it |

## 6. Where each wins

**Rill structural advantages:** the finished consumption layer (BI + embeds + alerts + AI chat) on the same class of engine; a governed metrics layer grounding APIs *and* agents (Tinybird has none); multi-engine + BYO ClickHouse (no fork lock-in); warehouse connector breadth; analyst/developer agents; seat+usage pricing when UIs are the deliverable.

**Tinybird structural advantages:** streaming ingestion (Events API + Kafka) — table stakes for true real-time; API-endpoint productization (docs, observability, typed params, Prometheus); zero-copy branches; TS/Python SDKs; developer-brand strength (Vercel ecosystem, starter kits, engineering blog); pure-usage pricing; deeper managed-CH operations.

**Tinybird's exploitable weaknesses:** no semantic layer, no BI, no business-user surface — every deal requires frontend engineering the buyer must staff; ClickHouse Inc. squeeze ($15B war chest, ClickPipes/Cloud closing the abstraction gap); fork lock-in + no BYO cluster; migration fatigue from the Forward/Classic sunset (deadline Sept 15, 2026); single-region orgs; small G2/analyst footprint.
