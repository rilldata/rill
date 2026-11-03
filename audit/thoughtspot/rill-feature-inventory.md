# Rill Feature Inventory (current state, ThoughtSpot-focused)

> Basis for the ThoughtSpot vs Rill comparison, derived from the codebase (Aug 2026) + docs.rilldata.com. Full detail: [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md). This file highlights the ThoughtSpot-relevant surfaces.

## Where the products meet

ThoughtSpot sells search/AI-first analytics ("ask, don't dashboard") with Spotter as its agentic analyst, plus a large embedded business. They meet in: **AI-BI evaluations** (NL accuracy + grounding), **embedded analytics** deals, and self-service analytics replacements. ThoughtSpot's public accuracy problem (54.67% in runQL's 2025 benchmark) makes grounding the battleground.

## Rill strengths relevant to this fight

- **Grounding by construction**: chat runs over the same governed metrics views that power dashboards — no separate worksheet-synonym curation layer; answers cite live dashboard URLs (verifiable, not just plausible).
- **Speed + cost**: ThoughtSpot live-queries the customer's warehouse (latency + cost) or requires loading into its stack; Rill's managed OLAP is sub-second by default and priced transparently ($20–30/seat vs ThoughtSpot's famously six-figure contracts).
- **Full data plane**: ingestion + transformation included.
- **Write-capable MCP + agent buildability**: external agents query *and build* Rill projects; ThoughtSpot's AI is consumption-only.
- **AI reports with per-recipient security**, feedback attribution, BYO LLM, conversation persistence/sharing.
- **BI-as-code + OSS**: git-native, local dev, free tier that is a complete product (ThoughtSpot has neither).
- **Embedded**: flat pricing, per-user RLS via attributes, per-end-user AI chat history, postMessage API.

## Rill gaps most exposed by ThoughtSpot

- **Search-style UX**: no tokenized search-to-answer surface; Rill's NL entry is chat only (no typeahead over dimension values/measures as a query builder).
- **Auto-insights**: no SpotIQ-style automated insight scans (Rill's "Explain" is click-driven root-cause, narrower).
- **AI accuracy proof**: no published benchmark/evals product (♻ shared gap — the runQL story shows buyers now demand evidence).
- **Semantic joins/multi-fact** (♻): ThoughtSpot worksheets model joins.
- **Change analysis/KPI monitoring**: ThoughtSpot's KPI monitor + change analysis vs Rill's threshold alerts (stronger governance, weaker analytics).
- **Enterprise IT**: SCIM, audit, orgs-style multi-tenant admin (♻).
- **Chart breadth + maps** (♻).

## Pricing (Aug 2026)

Starter $20/seat/mo · Growth $30/seat/mo (embedded $500/mo per 100 seats, BYO DB/LLM, AI tokens per seat) · Enterprise custom. OSS free forever.
