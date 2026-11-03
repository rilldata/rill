# ThoughtSpot vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill parity or better · 🟡 partial/different · ❌ absent · ➕ Rill advantage. Detail in `thoughtspot-features-and-market.md` and `rill-feature-inventory.md`.

**Framing:** ThoughtSpot is the AI-BI incumbent — Gartner Leader 2025/2026, the most-marketed "AI analyst" (Spotter), a real embedded business — carrying two public wounds: **accuracy** (54.67% in runQL's 2025 benchmark, worst of the four tested, with no benchmark of its own published) and **cost** (opaque credits, Vendr median $92.5K/yr, $50–150K implementations). Rill fights it in AI-BI evals and embedded deals where grounding, speed, and price decide.

## 1. AI & agentic features

| Feature | ThoughtSpot | Rill status | Notes |
|---|---|---|---|
| AI analyst chat | **Spotter 3**: multi-turn, MCP-host blending Slack/Salesforce/Jira context, code gen, forecasting, multi-step research | ✅🟡 | Rill analyst agent: dashboard-context, in-chat charts, URL citations. Spotter is broader (forecasting, unstructured blending); Rill is better-grounded |
| Grounding architecture | Models (semantic layer) + **Coaching** (NL instructions) + **verified answers** + pre-SQL validation + "Spotter Semantics" (deterministic token SQL) | ✅🟡 | Rill grounds in the metrics layer by construction; lacks verified-answers workflow (♻ certified queries) and coaching UI. TS's determinism pitch is a *response* to its 54.67% problem — Rill can make the same claim with cleaner architecture |
| Published accuracy | **54.67% (runQL 2025, last place)**; publishes no own benchmarks | ➕ | Neither publishes a benchmark — but TS's number is public and bad. Rill's eval launch (♻) should include a head-to-head on the runQL pattern |
| Agent team | **SpotterViz** (NL dashboards), **SpotterModel** (NL semantic modeling, human review), **SpotterCode** (embedded-app dev) | ✅ | Rill's developer agent covers all three jobs (canvas gen, `develop_file` metrics views, embed config) in one — less marketed, equally capable |
| MCP | Hosted **open-source MCP server** (mid-2025, "first major BI vendor"), OAuth; Spotter as MCP host | ✅➕ | Parity on server; Rill adds write tools + free local. TS beat Rill to the marketing |
| Slack/Salesforce surfaces | Native Slack app; Spotter in Salesforce records, ServiceNow, Teams/Sheets sync | ❌ | ♻ Slack agent gap; Salesforce/ServiceNow are enterprise-surface plays Rill lacks entirely |
| Auto-insights | SpotIQ (anomaly/trend/forecast scans), change analysis | 🟡 | Rill "Explain" is click-driven root-cause; no proactive scans, no forecasting (♻ Python agent tool) |
| Scheduled/autonomous AI | Limited — no Cube/Quick-style scheduled agent tasks found | ➕ | Rill AI reports (per-recipient RLS) are ahead here |
| AI pricing | "Unlimited LLM tokens" but **Spotter capped 25 queries/user/mo** + overage; AI add-on on free embed tier | ➕ | Rill token quotas are more generous/honest; BYO LLM (TS: OpenAI/Gemini/Arctic choice exists) |
| Analyst tooling | **Analyst Studio** (ex-Mode: SQL, notebooks, agentic data prep, spreadsheet coming) | ❌ | ♻ SQL console / Python gaps (omni §3.5, qs §3.3) |

## 2. Platform

| Feature | ThoughtSpot | Rill status | Notes |
|---|---|---|---|
| Engine | Live-query to warehouse (latency + cost passthrough); legacy Falcon on-prem; **SpotCache** (2026) snapshots — an admission that live-query breaks under AI workloads | ➕➕ | Rill's managed OLAP is the architecture SpotCache approximates. "Fixed cost, sub-second, AI-ready" is Rill's native story |
| Semantic layer | Models: joins, formulas, synonyms, RLS; TML modeling language | 🟡 | TS models join; Rill single-table (♻). Rill's is code-first/git-native; TML is export-oriented |
| Search UX | Tokenized typeahead search-to-answer | ❌ | Distinctive TS surface; Rill chat covers the job differently |
| Liveboards vs Explore | Liveboards = curated dashboards with drill | ✅➕ | Rill Explore's slice-and-dice at sub-second is stronger for operational data |
| BI-as-code | TML + git via API; UI-first product | ➕ | Rill git-native, local dev, OSS |
| Data plane | None (live query) + Mode-side mashups | ➕➕ | Rill ingestion + incremental models |
| Embedded | Mature: Visual Embed SDK, SpotterEmbed + headless SpotterAgentEmbed, Orgs multi-tenancy, Theme Builder, mobile SDKs | 🟡 | Rill embed is solid but lacks SDK wrapper (♻), headless chat API (♻), mobile. TS's headless Spotter embed is the reference for Rill's Chat API item |
| Enterprise | SAML/OIDC + **SCIM**, audit API, Orgs | 🟡❌ | ♻ SCIM/audit gaps |
| Open source / local | None; Cloud-only for all AI | ➕➕ | |

## 3. Pricing & market

| Dimension | ThoughtSpot | Rill | Notes |
|---|---|---|---|
| Entry | Essentials $25/user/mo (no AI); Pro ~$50 + credits; Enterprise median $92.5K/yr actual | $20–30/seat transparent | Rill dramatically cheaper with AI *included* |
| Credits | Consumption credits per interaction; bill-surprise complaints | Predictable seats + usage | A top-3 TS complaint — attack it |
| Embedded | ~$12K/mo floor; StartupSpot $12,999/yr flat (startups only) | $500/mo per 100 seats | Rill wins mid-market decisively; StartupSpot shows TS knows it |
| Implementation | +$50–150K services | Self-serve | |
| Analyst standing | **Gartner MQ Leader 2025+2026**, G2 Embedded BI Leader | Absent from analyst lists | TS's biggest asset; ♻ analyst-relations gap |
| Trajectory | $4.2B 2021 valuation, no raise since, secondaries discounted, layoff history, Mode integration troubled | Small but focused | Stability questions cut both ways (TS: big but bruised) |

## 4. Where each wins

**Rill structural advantages:** grounding-by-construction (one metrics layer, no Topics/coaching curation tax — TS's accuracy problem is architectural); managed OLAP speed at fixed cost (SpotCache validates the need); price and pricing legibility; BI-as-code + OSS + local dev; write-capable agents; per-recipient-security AI reports.

**ThoughtSpot structural advantages:** Gartner Leader brand + enterprise sales machine; Spotter marketing + 64%-adoption proof points; search UX; embedded maturity (headless agent embed, Orgs, mobile SDKs); Slack/Salesforce/ServiceNow surfaces; Analyst Studio; joins in the model; SCIM/audit/enterprise IT.

**ThoughtSpot's exploitable weaknesses:** the public 54.67% accuracy number and absence of self-published benchmarks; cost + credit-bill surprise + implementation services; live-query latency/cost (SpotCache is a patch, not an engine); Cloud-only AI, no OSS, no local; modeling-curation burden; Mode-acquisition damage and no funding since 2021 at a marked-down valuation.
