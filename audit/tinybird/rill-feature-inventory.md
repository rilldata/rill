# Rill Feature Inventory (current state, Tinybird-focused)

> Basis for the Tinybird vs Rill comparison, derived from the codebase (Aug 2026) + docs.rilldata.com. Full detail: [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md). This file highlights the Tinybird-relevant surfaces.

## Where the products meet

Tinybird is managed ClickHouse + SQL-pipes published as parameterized REST APIs — developer-first real-time analytics with **no BI surface**. They meet on Rill's own wedge: **real-time / user-facing / embedded analytics on ClickHouse-class engines**, with as-code workflows and AI-agent stories. The buyer decision is "build the analytics UI on Tinybird APIs" vs "ship Rill's embedded BI."

## Rill strengths relevant to this fight

- **A finished consumption layer**: explore + canvas + alerts + reports + embedded dashboards + per-user RLS — everything a team building on Tinybird must hand-build in React. "Tinybird gives you APIs; Rill gives you the product."
- **Semantic/metrics layer**: governed metrics views with security policies grounding dashboards *and* AI/MCP — Tinybird pipes are raw SQL endpoints with token-scoped row filters, no metric semantics.
- **Multiple engines**: DuckDB (zero-ops local + small/mid data), ClickHouse (incl. **BYO ClickHouse** — zero-copy on the customer's cluster, where Tinybird historically required its hosted platform), Druid/Pinot/StarRocks.
- **AI**: analyst/developer agents, dashboard chat, write-capable MCP, AI reports; Tinybird's AI is developer-workflow-focused.
- **BI-as-code parity**: YAML/SQL in git, local single binary (`rill start` ≈ `tb local`), branch deployments — Rill matches Tinybird's celebrated DX and adds the BI on top.
- **Custom APIs**: YAML endpoints over `sql`/`metrics_sql` resolvers with OpenAPI + security — Rill *does* have a "publish SQL as an API" story (less productized than Pipes).

## Rill gaps most exposed by Tinybird

- **Streaming ingestion**: Tinybird's Kafka connector + HTTP streaming endpoint (events API) are core; Rill has **no Kafka/streaming connector** (♻ noted in cube audit) — real-time freshness relies on ClickHouse-side ingestion or micro-batch models.
- **API-endpoint productization**: no per-endpoint tokens/rate limits/versioning/usage analytics; no JWT-token self-serve flow for third-party developers; custom APIs lack Tinybird's polish (docs, playground, observability per endpoint).
- **Materialized-view chains / real-time rollups**: Tinybird's incremental MVs at ingest time vs Rill's batch incremental models + early `rollup:`.
- **Endpoint observability**: per-API latency/QPS dashboards, query analysis tooling.
- **Developer ecosystem**: Vercel/Next.js templates, framework integrations, starter kits.

## Pricing (Aug 2026)

Starter $20/seat/mo · Growth $30/seat/mo (embedded $500/mo per 100 seats; ≤1TB managed DB at $0.20/CU-hr; BYO ClickHouse/MotherDuck at no usage cost; AI tokens per seat) · Enterprise custom. OSS free forever. Rill bills seats + compute; Tinybird bills pure usage (no seats) — expect Tinybird to win seatless-API deals on pricing shape, Rill to win when a UI/consumption layer is needed.
