# Rill Feature Inventory (current state, Looker-focused)

> Basis for the Looker vs Rill comparison, derived from the codebase (Aug 2026) + docs.rilldata.com. Full detail: [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) (BI surface) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md) (AI subsystem, semantic layer, APIs). This file highlights the Looker-relevant surfaces.

## Where the products meet

Looker is the incumbent git-native semantic-layer BI (LookML) — the product whose philosophy Rill's BI-as-code most resembles, now Google-owned with Gemini AI layered on. They meet in: code-first BI evaluations, **Looker-migration deals** (Rill as a destination for LookML refugees, competing with Omni/Lightdash for them), embedded analytics, and Google-gravity accounts.

## Rill strengths relevant to this fight

- **Speed without PDT engineering**: Looker's interactivity = warehouse latency + PDT/aggregate-awareness tuning; Rill's managed OLAP (DuckDB/ClickHouse/Druid/Pinot) delivers sub-second explore by default — directly answers Looker's "slow dashboards" reputation.
- **Modern code-first DX**: local single binary, YAML/SQL simpler than LookML, branch deployments, reconciler validation, cloud edit sessions — Looker's git-IDE workflow without its learning curve or SaaS-only development.
- **Full data plane**: ingestion + incremental/partitioned models built in (Looker has none — warehouse + dbt assumed).
- **AI**: metrics-layer-grounded chat with URL citations, write-capable MCP server (agents can *author* projects — LookML has no agent-authoring story), AI reports with per-recipient security, BYO LLM, `rill init --agent`.
- **Embedded economics**: flat $500/mo per 100 embedded seats + per-user AI chat vs Looker's quote-priced embed platform fees.
- **Open source + transparent pricing** vs Looker's ~$5k+/mo platform + per-seat quotes and post-acquisition price hikes.

## Rill gaps most exposed by Looker

- **Semantic joins** (♻, the big one): LookML explores model join trees with symmetric aggregates; Rill metrics views are single-table.
- **Semantic-layer maturity**: no extends/refinements-style model reuse beyond derived metrics views; no named access grants; no Liquid-style parameterized dimensions; no aggregate-awareness maturity (rollups are early).
- **Governed self-service explore across joined models**; no viewer-side custom fields/table calcs (♻).
- **Ecosystem**: no marketplace/blocks, no action hub (delivery destinations ♻), no custom-viz marketplace, thin partner ecosystem.
- **Enterprise IT**: no SCIM, audit logs, custom roles (♻); Looker's System Activity/audit is mature.
- **Semantic-layer-for-other-tools**: Looker pushes governed metrics into Sheets/Tableau/PowerBI (Modeler APIs); Rill has no SQL API/BI-tool sync yet (♻ cube roadmap §3.1/3.3).
- **Scheduled delivery breadth**: Looker's scheduler (destinations, formats, per-recipient filters) is deep; Rill has email/Slack + CSV/XLSX/Parquet.

## Migration-play assets (Rill → offense)

`runtime/ai/` developer agent + `develop_file` can drive a **LookML → metrics-view converter** (cube audit already proposes `rill migrate lookml`); Metrics SQL + planned Postgres-wire API ease client cutover; git-native workflow is familiar ground for LookML developers.

## Pricing (Aug 2026)

Starter $20/seat/mo · Growth $30/seat/mo (embedded $500/mo per 100 seats, BYO DB/LLM, AI tokens per seat) · Enterprise custom. OSS free forever.
