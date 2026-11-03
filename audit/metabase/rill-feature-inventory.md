# Rill Feature Inventory (current state, Metabase-focused)

> Basis for the Metabase vs Rill comparison, derived from the codebase (Aug 2026) + docs.rilldata.com. Full detail: [`../omni/rill-feature-inventory.md`](../omni/rill-feature-inventory.md) and [`../cube/rill-feature-inventory.md`](../cube/rill-feature-inventory.md). This file highlights the Metabase-relevant surfaces.

## Where the products meet

Metabase is the default open-source / low-cost BI for startups and the go-to cheap embedded-analytics choice. They meet in: **embedded analytics for SaaS** (the core head-to-head), OSS BI selection, and SMB/startup internal BI. The objection Rill faces: "why not just use Metabase — it's free/cheap and everyone knows it."

## Rill strengths relevant to this fight

- **Performance at scale**: Metabase queries the app DB/warehouse directly and struggles at event scale; Rill's OLAP engines (DuckDB/ClickHouse) + incremental ingestion give sub-second dashboards on billions of rows — the classic "we outgrew Metabase" moment is Rill's entry.
- **Real semantic layer + governance-as-code**: metrics views with templated security policies vs Metabase's models/sandboxing (Pro-gated, UI-configured); Rill's whole project is reviewable YAML in git with branch deployments (Metabase serialization is an export format, not a source-of-truth workflow).
- **Explore self-service**: purpose-built slice-and-dice (leaderboards/TDD/pivot, sub-second) vs Metabase's question builder.
- **AI**: full agent stack (chat grounded in metrics, write-capable MCP, AI reports, anomaly Explain) vs Metabot's narrower NL-to-query; BYO LLM.
- **Embedded depth**: per-end-user AI chat, postMessage API, magic-token public URLs with locked filters; server-enforced RLS via arbitrary JWT attributes (vs sandboxing's group/attribute mapping).
- **Both are OSS** — but Rill's OSS includes the OLAP engine and AI; Metabase's OSS gates RLS (sandboxing), SSO, and serious embedding behind Pro/Enterprise.

## Rill gaps most exposed by Metabase

- **Price floor + familiarity**: Metabase OSS is free with a huge install base; Cloud Starter is very cheap. Rill's $20+/seat and newer brand lose "good enough" deals.
- **SQL-first analyst UX**: Metabase's native SQL editor (variables/snippets) is loved; Rill has no cloud SQL surface (♻ omni §3.5).
- **Breadth of DB connectors for direct query**: Metabase connects to ~everything (incl. MongoDB, app DBs) for quick internal BI; Rill's connector list is ingestion-oriented.
- **Embedded SDK**: Metabase ships a React embedding SDK; Rill is iframe/postMessage only (♻).
- **Question-builder simplicity**: non-technical GUI query building over raw tables; Rill requires modeled metrics first.
- **X-rays/auto-dashboards**, collection UX conventions, massive community/docs.

## Pricing (Aug 2026)

Starter $20/seat/mo · Growth $30/seat/mo (embedded $500/mo per 100 seats, BYO DB/LLM, AI tokens per seat) · Enterprise custom. OSS free forever — including the engine and AI (with BYO key).
