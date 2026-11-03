# Metabase Competitive Analysis (Aug 2026)

Research basis: metabase.com pricing/releases/paid-features/embedded/AI pages (fetched directly), docs, GitHub, press, and the Rill codebase + docs.rilldata.com.

**TL;DR:** Metabase remains the default free/cheap BI and embedded choice (AGPL core, ~49k stars, "100,000+ companies," $100/mo Starter, React embed SDK GA, guest embeds with unlimited viewers) — and in 2026 it repositioned hard: Data Studio semantic layer with glossary, Remote Sync git workflows, and an aggressive AI arc (Metabot GA, **AI open-sourced with BYO model**, official MCP server with inline charts, Slack Metabot, AI governance, a CLI "built for coding agents"). Its structural limits persist: no engine (performance at scale is complaint #1), RLS/SSO/serialization paywalled at a $575/mo cliff, $12/authenticated-embed-user pricing, shallow governed-metrics depth. Rill's play is **the graduation path**: a Metabase importer + scale benchmark + embed-pricing calculator to catch teams outgrowing it, closing the two conveniences its users miss (cloud SQL console, embed SDK), a quick-start scaffolding flow to blunt the setup-speed gap — and noting that Metabase shipping MCP/Slack/AI-governance confirms Rill's shared Phase-1 AI-ops items are table stakes.

| File | Contents |
|---|---|
| [`metabase-features-and-market.md`](metabase-features-and-market.md) | Product (builder, SQL, models/metrics, Data Studio), OSS-vs-paid split + pricing, embedded suite (SDK, guest embeds, Tenants), Metabot AI arc v58–v63 + roadmap, BI-as-code (serialization, Remote Sync), market/complaints/momentum |
| [`rill-feature-inventory.md`](rill-feature-inventory.md) | Rill current state, Metabase-angled: engine/governance/AI strengths; gaps (price floor, SQL editor, embed SDK, connector breadth, familiarity) |
| [`comparison.md`](comparison.md) | Feature matrix + structural advantages/weaknesses both ways |
| [`roadmap.md`](roadmap.md) | Phase 1 graduation path (importer, benchmark, OSS positioning, charts), Phase 2 DX conveniences (SQL console, embed SDK, quick-start scaffolding), Phase 3 compounding moats, GTM, non-goals |
