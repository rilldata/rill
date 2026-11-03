# Hex Competitive Analysis (Aug 2026)

Research basis: hex.tech blog/changelog/pricing/docs, press, G2, and the Rill codebase + docs.rilldata.com. Structure mirrors the other `audit/` competitor analyses.

**TL;DR:** Hex ($172M raised; Series C May 2025 framed on "agentic analytics") evolved from AI notebook to the market's most complete AI-ops stack: Hex Agent across notebook/Slack/MCP/CLI/API, Threads self-serve chat, Context Studio (guides, endorsed tables, Review Agent, observability), customer-facing Evals plus the public **DataBench** benchmark, MCP server *and* client, model picker with Fast mode, user memory, Generative Apps, and a semantic layer that imports MetricFlow/Cube/LookML/Snowflake/Databricks models (plus the Hashboard acquisition). Its structural gaps are Rill's strengths: no RLS/governance engine (agent traceability is a documented complaint), no data plane or OLAP engine, Enterprise-gated expensive embedding, price cliffs, and no Gartner MQ presence. Rill's plan: match the AI-ops packaging cheaply (context page, evals **with a public benchmark**, Slack, memory), answer analysis depth with an agent Python tool + cloud SQL console instead of a notebook, and press governance/speed/embedded/price — where Hex can't follow without becoming a different product.

| File | Contents |
|---|---|
| [`hex-features-and-market.md`](hex-features-and-market.md) | Notebook platform, semantic-layer strategy (Sync/Authoring/Modeling Agent, OSI), Explore + Generative Apps, full AI timeline (Magic → Notebook Agent → Threads → Context Studio → Evals/DataBench), embedded, enterprise, pricing, company/market |
| [`rill-feature-inventory.md`](rill-feature-inventory.md) | Rill current state, Hex-angled: BI/embedded/governance strengths; gaps (Python/notebook, agentic code analysis, data apps, SQL workbench) |
| [`comparison.md`](comparison.md) | Feature matrix (AI first) + structural advantages and exploitable weaknesses both ways |
| [`roadmap.md`](roadmap.md) | Phase 1 AI-ops parity with a raised bar (public benchmark, context page, memory), Phase 2 analysis-depth answers (Python agent tool, SQL console, agent API, outbound MCP), Phase 3 governance offensive + embedded v2, GTM (sell-alongside, governance-first), non-goals |
