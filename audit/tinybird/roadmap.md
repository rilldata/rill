# Roadmap: Competing with (and beating) Tinybird

Companion to `comparison.md`. Sizes S/M/L/XL; shared items ♻ per prior roadmaps.

## 1. Strategic frame

Tinybird and Rill sell to adjacent buyers on the same wedge (real-time, ClickHouse-class, code-first, agent-ready analytics). Tinybird's buyer is a **developer who wants APIs and will build the UI**; Rill's is a team that wants **the metrics layer and product shipped**. Deals blur when a dev team debates build-vs-buy for user-facing analytics — and in "analytics backend for AI agents" evaluations, where both claim the story.

Rill's decisive advantages: the governed metrics layer (Tinybird's agents query raw SQL endpoints; Rill's query semantics) and the finished BI/embed product. Tinybird's decisive advantage: **streaming ingestion** and API-product polish. Also note the strategic backdrop: ClickHouse Inc. ($15B, Jan 2026) squeezes Tinybird's core abstraction far harder than it does Rill (multi-engine, BYO-CH, differentiated up-stack).

**Positioning: "Everything you'd build on Tinybird — already built, and governed."** APIs when you want APIs, product when you want product, one metrics layer for both.

### Deal-loss categories
1. "We need Kafka/streaming ingestion." (real-time table stakes)
2. "We want APIs, not a BI tool." (API productization)
3. "Tinybird is the agent-analytics backend." (narrative)

## 2. Phase 1 — Close the real-time gap (0–3 months → L overall)

- **Kafka connector** (L): the single most important item in this fight — flagged in three audits now (cube, quicksight, tinybird). Land Kafka → incremental model ingestion for DuckDB and native Kafka table engines for managed ClickHouse. Include Schema Registry.
- **HTTP events endpoint** (M): a Tinybird-Events-API-style `POST /events` ingest path per project (buffered → incremental model). Massively simplifies the "instrument my app" story and demos.
- **Streaming-freshness docs/benchmark** (S): publish end-to-end freshness numbers (event → dashboard) for the new paths.

## 3. Phase 2 — API-product polish (3–6 months)

Rill's custom APIs already exist; make them a headline product ("Metrics APIs"):
- **Typed parameters + validation** on custom APIs (S–M), matching Tinybird's templating DX.
- **Per-endpoint observability** (M): latency/QPS/error metrics per custom API + MCP tool, surfaced in a usage page (♻ usage-analytics item) and as queryable metrics.
- **Prometheus format + client snippets** (S); **starter kits** (S–M): web-analytics kit, Next.js dashboard template, LLM-usage tracker — Tinybird's template motion works; Rill's versions ship with the metrics layer + explore included ("the template *is* a full BI project").
- **JWT quickstart for third-party developers** (S): self-serve docs for customer-signed embed/API tokens (machinery exists; DX packaging doesn't).

## 4. Phase 3 — Win the agent-backend narrative (6–12 months)

- **Governed-vs-raw agent story** (content + ♻ evals): benchmark agent accuracy over Rill metrics tools vs raw SQL endpoints (Tinybird's own LLM-SQL benchmark shows models writing analytical SQL imperfectly — that's Rill's argument). Publish alongside the shared accuracy benchmark (♻).
- **Agent skills repo** (S, ♻cube §2.1): `rilldata/rill-agent-skills` — Tinybird's is already public.
- **Zero-copy branch deployments** (M–L): for managed-CH projects, branch against production data without re-ingest (Tinybird's Branches GA sets the expectation; Rill's branch deployments currently rebuild).
- **SDK evaluation** (M if pursued): TS/Python resource definitions compiling to Rill YAML — demand-gated; YAML + agents may suffice.

## 5. GTM

1. **"Rill vs Tinybird" build-vs-buy page**: TCO of engineering the UI/semantic layer on Tinybird vs shipping Rill; "APIs *and* the product."
2. **Court the same developer ecosystem**: Vercel/Next.js templates, engineering-blog cadence, HN presence — Tinybird's channel is winnable with Rill's stronger free product (full BI locally).
3. **BYO-ClickHouse contrast**: Tinybird = their fork, their cloud; Rill = your cluster, zero-copy. Aim at teams already running CH.
4. **Migration-window play**: Tinybird Classic sunsets Sept 15, 2026 — forced-migration moments are switching moments; publish a Tinybird→Rill guide (datasources/pipes → models/metrics views/custom APIs) timed to it.
5. **ClickHouse co-marketing** (♻cube GTM): Rill's BYO posture makes it ClickHouse Inc.'s friend where Tinybird is its rival.

## 6. Non-goals

Managed-CH ops depth as a product (replica weights, cluster APIs) beyond what managed offerings need; raw-SQL-endpoint positioning (governed metrics APIs instead); competing for pure data-infrastructure (no-UI) workloads on price; a charts component library.

**One-line strategy:** close streaming ingestion (Kafka + events endpoint), package Metrics APIs with Tinybird-grade DX and starter kits, and win the agent-backend narrative with the thing Tinybird structurally lacks — a governed semantic layer — while catching its Classic-sunset and ClickHouse-squeeze fallout.
