# ThoughtSpot — Platform, AI & Market Inventory

> Researched Aug 2026 from thoughtspot.com (PR, pricing, docs at docs.thoughtspot.com, developers.thoughtspot.com), GlobeNewswire/TechTarget coverage, runQL benchmark, G2/Gartner Peer Insights, pricing teardowns. Founded 2012 (Ajeet Singh, Amit Prakash); CEO **Ketan Karkhanis** (ex-Salesforce, since Sep 2024); ~1,700 employees.

## 1. Core product

- **Tokenized/relational search** (the original differentiator): queries composed from typed search tokens (measures, attributes, `top`, `growth of`, `vs`) compiling **deterministically** to SQL — not free-text LLM SQL; now marketed as the accuracy backbone under Spotter.
- **Answers** (saved single results), **Liveboards** (interactive dashboards; "AI-augmented" since 2025), **Worksheets → Models** (the semantic layer: joins, formulas, synonyms, row-level filters; Worksheets deprecated, Models default), **SpotIQ** (legacy auto-insights: anomaly/trend/forecast — still shipped, no longer the marketing lead).
- **Live-query only** (Cloud) against Snowflake, Databricks, BigQuery, Redshift, Synapse, Starburst/Trino, Oracle, SAP HANA, Postgres/MySQL, Dremio, Denodo, Teradata; dbt certified on the big four. Legacy on-prem **Falcon** in-memory engine survives as "ThoughtSpot Software" (paid migration services to Cloud; all innovation Cloud-only).
- **SpotCache (Feb 2026)**: new caching layer running queries against snapshots *without* hitting the warehouse — "unlimited analytics usage for AI workloads with fixed cloud costs"; a notable retreat from pure live-query for cost reasons.
- **Mode → Analyst Studio**: acquired Mode for **$200M** (2023; doubled customer count, ARR past $150M); absorbed into **Analyst Studio** (SQL/notebooks, GA early 2025). **Next-gen (Feb 2026)**: SpotCache, agentic data-prep agent, unified data mashup (warehouse + SaaS + Sheets in the SQL IDE), native spreadsheet interface (phased 2026). Benn Stancil's "How an acquisition fails" is the candid account of the integration.

## 2. AI features (timeline)

| Date | Release |
|---|---|
| Mar 2023 | **Sage** (GPT search on the token engine; Sage Embed) — retired, replaced by Spotter |
| Nov 13, 2024 | **Spotter** "the Autonomous Agent for Analytics": conversational multi-turn AI analyst; LLM-agnostic (OpenAI, Gemini, Snowflake Arctic); in-product, embedded, third-party surfaces |
| Jun 2025 | **Agentic Analytics Platform for Snowflake** (Cortex + Snowpark; Snowflake Marketplace) |
| mid-2025 | **Agentic MCP Server** — first major BI vendor with a hosted MCP server (open source; OAuth); exposes Spotter NL-query, models, Liveboards, Answers to Claude/ChatGPT/Gemini/Cursor |
| Sep 10, 2025 | **Spotter 3** + Embedded revamp ("Boundaryless, Agentic Intelligence"): Spotter as **MCP host** blending warehouse data with Slack/Salesforce/Jira; code generation, forecasting, multi-step research; embed via SDK/REST/MCP, Developer Playground, Theme Builder, mobile SDKs |
| Nov 20, 2025 | **StartupSpot**: embedded agentic analytics, flat **$12,999/yr** (50 customer accounts + 50 internal users; <50 employees, <$3M revenue) |
| Dec 10, 2025 | **Four BI agents "that work as a team"**: **Spotter 3** (consumption), **SpotterViz** (NL dashboard creation), **SpotterModel** (NL semantic-model building, human review), **SpotterCode** (AI-assisted embedded-app development). GA early 2026; custom user-built agents promised |
| Feb 18, 2026 | Agentic Data Prep + next-gen Analyst Studio + SpotCache |
| Mar 12, 2026 | **Spotter Semantics**: "the industry's leading agentic semantic layer" — token architecture + TML-based modeling language, deterministic SQL, pitched on *consistency* |
| Jun 2, 2026 | Snowflake **Cortex AI + Semantic Views** integration (agents inside the Snowflake boundary; BYO Snowflake LLM) |

- **Trust machinery**: human-in-the-loop **verified answers** (curated reference questions), **Coaching** (NL instructions steering Spotter), Spotter Model readiness scoring, pre-SQL validation against the semantic model.
- **The accuracy problem**: **runQL 2025 benchmark — Spotter 54.67%, last of four tested** (runQL 92%, Cortex Analyst 88.67%, Genie 84.67%). No direct rebuttal; the functional response is Spotter Semantics' determinism/consistency framing. Press noted ThoughtSpot publishes **no benchmarks of its own**.
- **Surfaces**: in-product, embedded (SpotterEmbed / headless SpotterAgentEmbed), **Slack** (Marketplace app), **Salesforce** (in CRM records), ServiceNow, Teams/Sheets via Sync, mobile, any MCP client.
- AI pricing posture: "unlimited LLM tokens" but **Spotter capped at 25 queries/user/mo** on Pro/Enterprise with overage billing (~$0.10+/query); embedded AI agents are a paid add-on on the free dev tier.

## 3. Embedded ("ThoughtSpot Embedded," ex-Everywhere)

Visual Embed SDK (Search, Liveboard, full app, **SpotterEmbed** conversational UI, **SpotterAgentEmbed** headless bring-your-own-UI), REST APIs, webhooks, custom actions, mobile SDKs, Developer Playground, no-code Theme Builder, MCP embedding. Multi-tenancy via **Orgs**. Free Developer tier (10 users, 25M rows, 1 yr; AI add-on paid); Enterprise Embedded custom (~$12K/mo floor per third parties; $5–6/dashboard-view under usage models); StartupSpot $12,999/yr. Named embedded customers: Verisk, WEX, Harri, Act-On, CarTrawler, Accern, MDAudit, Abrigo, Goldcast, Loan Market Group.

## 4. Enterprise & governance

RLS (group/user rules, enforced across Spotter), groups/privileges, **Orgs** multi-tenancy (per-Org IdP claims), SAML/OIDC + SCIM, audit-log API. Cloud SaaS strategic; on-prem legacy only — **Spotter/agents/Analyst Studio/MCP are Cloud-only**. AI governance claims: no token metering, RLS respected, Snowflake-boundary execution.

## 5. Pricing

- **Analytics Essentials $25/user/mo** (5–50 users, 25M rows, **no Spotter**) · **Analytics Pro** (~$50/user/mo cited; usage credits $0.10/credit; Spotter 25 queries/user/mo; up to 1,000 users/250M rows) · **Analytics Enterprise** custom — Vendr median actual **$92,521/yr** (range ~$37K–$231K); large deals $400K+.
- **Consumption/credit model** underneath: every interaction consumes credits — recurring bill-surprise complaints (30–50% overage months reported).
- Implementation commonly adds $50–150K or 15–40% of year-one subscription. Historic: ~$200K average deals pre-2022; $95/mo Team Edition era ended.

## 6. Company & market

- **Funding**: ~$663–674M raised; Series F Nov 2021 $100M at **$4.2B**; no round since; secondary pricing reportedly well below the 2021 mark. **Still private, no S-1**; stated goal $1B ARR before IPO.
- **Scale**: $150M+ ARR & 1,000+ customers (FY2023, post-Mode); ~40% SaaS growth 2024; FY2025 claims: usage +133% YoY, **64%+ of customers actively use Spotter**. Layoff history through the cloud transition and post-Mode.
- **Customers**: Verizon, CVS Health, Capital One, Comcast, Wells Fargo, Toyota, Snowflake itself; digital natives DoorDash, Calm, Zillow, Airtable, Huel.
- **Gartner MQ ABI: Leader 2025 (first time) and 2026** — markets itself as "one of the only independent vendors" among Microsoft/Salesforce/Google/Qlik/Oracle. G2 Leader grids incl. Embedded BI; ~89% satisfaction across ~490 reviews.
- **Complaints**: cost (opaque credits, spiky bills, expensive mid-market), semantic-modeling burden (Spotter quality ∝ model curation), token-search learning curve, support responsiveness, docs lag, weak pixel-perfect reporting/custom viz.
- **Momentum mid-2025→mid-2026**: Snowflake-native platform (Jun 2025) → Spotter 3 + embedded revamp (Sep) → StartupSpot (Nov) → four BI agents (Dec) → Analyst Studio/SpotCache (Feb 2026) → Spotter Semantics (Mar) → Cortex/Semantic Views (Jun). Monthly `26.x.cl` cloud releases. OSI participant.

## 7. Competitive posture

- Long-running **"dashboards are dead"** anti-Tableau/Power BI campaign, evolved into "agentic analytics"/"the agentic enterprise" under Karkhanis (ironically, 2025's SpotterViz agent *builds dashboards*).
- Vs **Genie/Cortex Analyst**: partner-and-differentiate — cross-cloud neutrality, governed semantic layer with coaching/verification, consumer UX, embedding, MCP interop; while simultaneously co-engineering deeply with Snowflake. The runQL result (behind both) is the awkward data point.
- In embedded: leads with "AI analyst in your product in a few lines of code," Orgs multi-tenancy, Gartner Leader status; StartupSpot/free tier target the low end where consumption pricing lost deals.
