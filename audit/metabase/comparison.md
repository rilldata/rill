# Metabase vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill parity or better · 🟡 partial/different · ❌ absent · ➕ Rill advantage. Detail in `metabase-features-and-market.md` and `rill-feature-inventory.md`.

**Framing:** Metabase is the "good enough, everyone knows it" objection: free OSS, five-minute setup, the default cheap embedded choice, and — as of 2026 — surprisingly fast-moving on AI (open-sourced Metabot, MCP server, Slack, BYO model, AI governance) and as-code (Remote Sync, Data Studio). Rill wins when data gets big, embeds get serious, or governance/AI accuracy matters; it loses "just give me dashboards" deals on price floor and familiarity.

## 1. AI & agentic features

| Feature | Metabase | Rill status | Notes |
|---|---|---|---|
| NL chat | Metabot GA (v58): questions, charts, SQL gen/fix, summaries; "every answer inspectable" | ✅ | Rill parity+ (dashboard context, conversation persistence/sharing, citations) |
| Grounding | Data Studio semantic layer + **glossary**; surfaces existing dashboards first; respects permissions | ✅🟡 | Rill's metrics layer is deeper (measures/policies); Metabase's glossary is a nice idea (♻ synonyms/field metadata) |
| Open-source AI | **AI open-sourced (v60)** incl. self-hosted Metabot + BYO key | ✅ | Both now OSS-with-BYO-key — Metabase matched Rill's differentiator; Rill's agent stack is deeper (multi-agent, developer agent, write tools) |
| MCP server | Official (v60), interactive charts inline (v62) | ✅➕ | Parity; Rill adds write tools + OAuth cloud story |
| Slack agent | Metabot in Slack (v60) | ❌ | ♻ the recurring gap — now even Metabase has it |
| Coding-agent DX | CLI "built for coding agents," dashboards-as-code via MCP, agent skills for embedding | ✅➕ | Rill's `rill init --agent` + reconciler-validated YAML is stronger; Metabase is converging fast |
| AI governance | Per-group access, token limits (user/group/tenant), system prompts, usage analytics (v61) | 🟡 | Ahead of Rill's admin surface (♻ AI admin console, cube §2.6) — even Metabase ships this now |
| Developer/authoring agent | AI SQL + transform codegen; no project-building agent | ➕ | Rill's developer agent (whole projects, ingestion included) is a class above |
| AI reports/scheduled | None | ➕ | Rill AI reports with per-recipient RLS |
| Anomaly/insights | X-rays (dated) | ✅ | Rill "Explain" is the modern version |

## 2. Platform

| Feature | Metabase | Rill status | Notes |
|---|---|---|---|
| Engine | None — live queries to source; caching + model persistence as mitigation; **performance on large data is complaint #1** | ➕➕ | Rill's OLAP + incremental ingestion = the "we outgrew Metabase" landing zone |
| Semantic layer | Models + Metrics + Data Studio (2026) — real progress, still thinner than metrics-view spec | ✅ | Rill deeper (measure types, security, time semantics); Metabase broader on GUI query building over raw tables |
| Ad-hoc GUI queries over raw tables | Core strength — no modeling required | ❌ | Rill requires modeled metrics first; deliberate stance, but loses quick-start internal BI |
| SQL editor | Loved: variables, field filters, snippets | ❌ | ♻ cloud SQL console (omni §3.5) |
| RLS | **Sandboxing = paid**; impersonation; DB routing; Tenants | ✅➕ | Rill's templated policies are OSS-available and more expressive |
| BI-as-code | Serialization + Remote Sync (UI→git YAML) — export-shaped, not source-of-truth | ➕ | Rill git-native with validation; Metabase closing the gap for its audience |
| Writeback | Actions (DB writeback on dashboards) | ❌ | Rill has no writeback (♻ noted vs Sigma too) |
| Viz breadth | Waterfall, Sankey, box plot, treemap, maps + custom-viz React SDK | 🟡 | Metabase now ahead on chart-type count (♻ maps etc.); its custom-viz SDK is the shape Rill's Vega escape hatch competes with |
| Documents | Notebook-style narrative docs (all plans) | ❌ | Low priority |
| Connector breadth (live query) | ~20+ incl. MongoDB, app DBs | 🟡 | Different philosophy (ingest vs query-in-place) |

## 3. Embedded (the head-to-head)

| Feature | Metabase | Rill status | Notes |
|---|---|---|---|
| SDK | **React SDK GA** + framework-agnostic JS library, component-level (incl. AI chat component) | 🟡 | Rill iframe/postMessage only (♻ SDK) — Metabase is the reference implementation here |
| Guest/anonymous embeds | Free unlimited viewers (static/guest) | ✅ | Rill public URLs (magic tokens, locked filters) — comparable |
| Multi-tenant RLS | Sandboxing + Tenants + DB routing (paid) | ✅ | Rill attributes-based RLS at least parity, simpler |
| Embedded AI | Metabot chat component in embeds | ✅ | Rill per-end-user chat history is ahead |
| White-label | Full (paid) | ✅ | |
| Pricing | Pro $575/mo + **$12/authenticated embed user/mo**; Enterprise $20K+ for scale packages | ➕🟡 | Rill $500/mo per 100 seats ≈ $5/user — beats Metabase's $12/user for authenticated embeds; Metabase wins where guest embeds suffice |
| Performance at concurrency | Live-query bottleneck — the known embedded weakness | ➕➕ | Rill's core wedge |

## 4. Where each wins

**Rill structural advantages:** OLAP speed at event scale (Metabase's #1 complaint is Rill's #1 strength); real ingestion/transformation; deeper semantic + security layer (OSS, not paywalled); deeper agent stack (developer agent, AI reports, write MCP); embedded performance + cheaper authenticated-user pricing; git-native source of truth.

**Metabase structural advantages:** free OSS with massive familiarity + install base ("100,000+ companies", 49k stars); five-minute setup over any DB with GUI queries and no modeling; loved SQL editor; React embed SDK + guest embeds; chart-type breadth + custom-viz SDK; writeback Actions; huge community/docs; monthly release cadence now aimed at AI + as-code.

**Metabase's exploitable weaknesses:** no engine (performance cliff at scale — the classic graduation moment); RLS/SSO/serialization paywalled (OSS users hit the wall exactly when they get serious); $100→$575 pricing cliff; per-authenticated-user embed pricing at scale; semantic layer still shallow for governed metrics; no AI analyst depth (reports, anomaly, developer agent); permissions complexity at scale.
