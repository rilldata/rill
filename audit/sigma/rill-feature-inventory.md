# Rill Feature Inventory (current state, Sigma-focused)

> Basis for the Sigma vs Rill comparison, derived from the codebase (Aug 2026) + docs.rilldata.com. Full detail lives in [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) (BI surface) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md) (AI subsystem, semantic layer, APIs) — both current. This file highlights the surfaces where Sigma competes.

## Where the products meet

Sigma is a warehouse-native, spreadsheet-UI BI tool: every interaction compiles to live warehouse SQL; self-service depth (spreadsheet functions, pivots, input tables/writeback) is its identity. Rill is a full-stack OLAP-backed BI-as-code platform. They meet in modern-data-stack BI evaluations, embedded analytics, and increasingly AI-BI evals.

## Rill strengths relevant to this fight

- **Own data plane + engine**: ingestion (S3/GCS/warehouses), incremental + partitioned SQL models, managed DuckDB/ClickHouse (+Druid/Pinot/StarRocks) → sub-second explore **without passing every click through the customer's warehouse bill** (Sigma's architecture makes warehouse compute the hidden cost of self-service).
- **BI-as-code**: YAML/SQL in git, branch deployments, local single-binary dev, cloud edit sessions, continuous reconciler validation. Sigma is UI-first with version tagging, not git-native source.
- **Governed semantic layer**: metrics views (measures incl. derived/window/time-comparison; security policies with row filters + field access templated on JWT attributes) as the single source for dashboards *and* AI.
- **AI/agents**: multi-agent runtime (analyst/developer/feedback), dashboard-context chat, write-capable **MCP server** (free incl. local), AI reports with per-recipient security, BYO LLM from Growth tier, `rill init --agent` for coding agents.
- **Embedded**: iframe + postMessage, server-enforced per-user RLS via attributes, per-end-user AI chat history, flat $500/mo per 100 embedded seats.
- **Open source**: the whole product, free locally.

## Rill gaps most exposed by Sigma

- **Self-service depth**: no viewer-side calculated fields/formulas, no spreadsheet-style calcs, no ad-hoc joins by analysts, no un-modeled-column exploration. Sigma's core identity. (Tracked: omni roadmap §3.1 ad-hoc measures.)
- **Input tables / writeback**: nothing comparable (Sigma writes back to the warehouse; used for forecasting/what-if/data-entry apps).
- **CSV upload by end users** in cloud: developer-flow only.
- **Semantic joins/multi-fact** (♻ recurring): metrics views are single-table.
- **Actions/interactivity framework**: no cross-widget actions, no click-to-filter, no workflow triggers from dashboards.
- **Enterprise IT**: no SCIM, no audit logs (♻).
- **Chart breadth**: no maps/waterfall/gauge/sankey/boxplot (♻); Vega-Lite escape hatch only.
- **Version history UI**: git history exists but no in-UI restore/tagged-version workflow for non-git deploys.

## Pricing (Aug 2026, for comparison)

Starter $20/seat/mo · Growth $30/seat/mo (embedded $500/mo per 100 active embedded seats; BYO DB + BYO LLM; AI token quotas per seat) · Enterprise custom (unlimited embedded, VPC peering). OSS product free forever. Note: Rill's price includes the compute engine; warehouse-native tools' seat price excludes the warehouse bill their queries generate.
