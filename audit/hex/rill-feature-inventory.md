# Rill Feature Inventory (current state, Hex-focused)

> Basis for the Hex vs Rill comparison, derived from the codebase (Aug 2026) + docs.rilldata.com. Full detail: [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md). This file highlights the Hex-relevant surfaces.

## Where the products meet

Hex is a notebook-first AI analytics workspace (SQL + Python + no-code cells → published data apps + Explore self-serve) aimed at data teams. They meet in: **AI-analytics evaluations** for data-team-led orgs, the "governed self-serve" surface (Hex Explore vs Rill Explore), and occasionally data-app/embedded use cases. Hex pressures from the analyst/data-science side; Rill from the operational-BI/metrics side.

## Rill strengths relevant to this fight

- **A real BI product**: explore dashboards (sub-second slice-and-dice for non-technical users), canvas, alerts, scheduled reports, public URLs, embedded analytics with per-user RLS — Hex's app/Explore surfaces are younger and analyst-centric.
- **Own OLAP engine + data plane**: Hex pushes compute to the customer's warehouse (or limited built-in compute); Rill's managed DuckDB/ClickHouse gives interactive speed and predictable cost; ingestion/incremental models included.
- **Governed metrics layer**: one code-defined semantic layer with security policies grounding dashboards *and* AI; Hex's semantic story leans on dbt/external layers.
- **Embedded**: productized iframe/tokens/RLS/pricing; Hex's embedding is secondary.
- **Write-capable MCP server + agent-buildable projects**; BYO LLM below enterprise; per-recipient-security AI reports.
- **OSS + local dev**; transparent seat pricing vs Hex's per-editor + compute pricing.

## Rill gaps most exposed by Hex

- **Python/notebook analysis**: nothing comparable — no code cells, no ad-hoc Python, no forecasting/stats (♻ quicksight roadmap §3.3 proposes an agent-side compute tool).
- **Analyst workflow**: no SQL workbench in cloud (♻ omni §3.5), no document/notebook artifact, no chained-query exploration surface.
- **Agentic notebook work**: Hex's Notebook Agent does multi-step autonomous analysis with code; Rill's analyst agent queries metrics only (no code execution).
- **Data apps**: no interactive-app building (inputs → parameterized logic → published app); Rill canvas variables are a thin analog.
- **Collaboration**: no comments/review threads on analyses.
- **Semantic joins** (♻), viewer-side calcs (♻), SCIM/audit (♻).

## Pricing (Aug 2026)

Starter $20/seat/mo · Growth $30/seat/mo (embedded $500/mo per 100 seats, BYO DB/LLM, AI tokens per seat) · Enterprise custom. OSS free forever.
