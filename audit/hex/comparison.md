# Hex vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill parity or better · 🟡 partial/different · ❌ absent · ➕ Rill advantage. Detail in `hex-features-and-market.md` and `rill-feature-inventory.md`.

**Framing:** Hex attacks analytics from the data-team side (notebooks → agents → self-serve); Rill from the operational-BI side (metrics → dashboards → agents). They increasingly collide in "agentic analytics" evaluations — and Hex's AI-ops packaging (Context Studio, Evals, DataBench, Slack, MCP both directions, model picker) is currently the most complete in the market, ahead of even Cube. Rill's counters are governance depth (real RLS), a finished BI/embedded product, its own OLAP engine, and price.

## 1. AI & agentic features

| Feature | Hex | Rill status | Notes |
|---|---|---|---|
| Conversational self-serve | **Threads**: grounded chat, Explore cells in answers, endorsed/sensitive modes, thread→notebook/app | ✅ | Rill chat: dashboard-context-aware, in-chat charts, URL citations, share/fork. Parity on core; Hex ahead on modes + convert-to-artifact |
| Agentic notebook work | **Hex Agent**: plans multi-step analyses, chains SQL/Python cells, forecasts, builds apps | ❌🟡 | Rill analyst agent queries metrics only — no code execution, no multi-cell artifacts (♻ Python agent tool, qs §3.3) |
| Agent context curation | **Context Studio**: semantic models + Guides + endorsed tables + Review Agent + observability with audit trails | 🟡 | Rill: `ai_instructions` + feedback attribution. Context Studio is the category benchmark — richer than Cube's rules/memories |
| Customer evals | **Evals** (CLI suites, rubrics, drift runs) + **DataBench** public benchmark | ❌ | ♻ cube §2.3 — Hex went further: they published the benchmark. Raises the bar for Rill's eval launch |
| Slack agent | @Hex in channels, charts in-thread, full thread context | ❌ | ♻ shared gap |
| MCP server | Remote MCP (OAuth; Claude/ChatGPT/Cursor/Codex), interactive Claude Connector | ✅ | Parity; Rill adds **write tools** (build the project, not just query) |
| MCP client (outbound) | Yes (May 2026) + agent web search | ❌ | ♻ cube §3.4 |
| Agent in CLI/API | Hex CLI + full agent via API; Agent Tasks (scheduled Qs → Slack/email) | 🟡 | Rill: `rill chat` CLI + AI reports (per-recipient security ➕); no agent-over-API productization |
| Model choice | Picker (Claude/GPT/Kimi/Fable), Auto mode, Fast mode, BYOK (Ent) | ✅🟡 | Rill BYO LLM from Growth (cheaper tier than Hex's Ent BYOK); no in-product model picker |
| Generative dashboards | Generative Apps: prompt→app, direct code editing, brand `design.md` | ✅ | Rill: AI canvas generation + developer agent `init_project` (includes ingestion — broader) |
| User memory | Automatic per-user agent memory | ❌ | ♻ cube §2.2 memories |
| AI grounding | Semantic Sync'd models + Guides + endorsed tables; but agent can emit raw SQL/Python — traceability complaint | ➕ | Rill grounds in the governed metrics layer by construction; the complaint Hex gets is Rill's core pitch |

## 2. Platform

| Feature | Hex | Rill status | Notes |
|---|---|---|---|
| Engine/data plane | None — warehouse compute + hosted kernels; no ingestion/transformation of its own | ➕➕ | Rill: ingestion, incremental models, managed OLAP, sub-second |
| Semantic layer | Import-first (MetricFlow/Cube/LookML/Snowflake/Databricks) + native authoring; thinner joins than Looker-class | 🟡➕ | Rill's is native + governance-complete (RLS); Hex's is broader via imports. Note: **Hex imports Cube & LookML — Rill could too** (♻ importers) |
| RLS/governance | **No RLS engine**; embed attributes + parameterized queries only | ➕➕ | Rill security policies (row + field, templated JWT attrs, per-recipient reports) — decisive in governed-BI evals |
| Notebook/SQL analyst UX | Best-in-class polyglot notebooks, reactive DAG | ❌ | Rill has no notebook/SQL workbench (♻ omni §3.5); different product philosophy — don't chase the document model |
| Self-serve explore | Explore (drill-down Team+), reads semantic models | ✅➕ | Rill Explore is faster, richer (TDD, pivot, comparisons), viewer-tier not editor-tier |
| Dashboards/BI ops | Apps + schedules; no alerts-on-metrics depth, no public URLs, thin theming | ➕ | Rill: alerts (per-user RLS), reports, bookmarks, public URLs, embedded product |
| Embedded | Signed embeds, attribute RLS; **Enterprise-only add-on**, expensive | ➕ | Rill embedded is productized + flat-priced + includes AI chat |
| Version control | History + git export (versions→commits) | ✅➕ | Rill: git *is* source; local dev; OSS |
| Compute pricing | Editor seats + PAYG kernels + AI credits | ➕ | Rill simpler; viewers free in both, but Hex's Explorer seat is a paid gate on self-serve |

## 3. Where each wins

**Rill structural advantages:** real governance (RLS everywhere incl. AI), finished BI + embedded product with flat pricing, own OLAP engine + data plane (speed + predictable cost), grounding-by-construction (Hex's traceability complaint), OSS + local dev, cheaper seats ($20–30 vs $36–75/editor + Explorer add-ons).

**Hex structural advantages:** the most complete AI-ops stack shipping today (Context Studio, Evals, DataBench, memory, model picker, Fast mode), Python/code execution for real analysis (forecasting, ML), MCP both directions + web search, notebook workflow data teams love, Snowflake partnership halo, high shipping cadence, strong logos (Reddit, Anthropic, Figma).

**Hex's exploitable weaknesses:** no RLS/governed layer of its own; not in the Gartner MQ; embedding Enterprise-gated and expensive; performance complaints on large data (no engine); price cliffs (Pro→Team, Explorer seats sales-gated); agent SQL traceability for non-technical users; semantic layer is a consumer of others' models more than a governed source of truth.
