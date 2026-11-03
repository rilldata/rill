# Amazon QuickSight / Quick Suite / Amazon Quick — AI Features & Market Perception

> Researched Aug 2026 from AWS blogs/docs/what's-new, press, analyst coverage, HN, review aggregators. **Naming**: QuickSight (2015–Oct 2025) → **Amazon Quick Suite** (Oct 9, 2025) → **Amazon Quick** (2026 H1), with the BI module surviving as "Quick Sight" inside the suite. Reddit was inaccessible; community sentiment triangulated via HN + review sites. The 2026 Gartner "Leader" claim rests on one secondary source — verify before external use. Companion file `quicksight-features-core.md` covers the core BI platform.

## 1. AI feature timeline

| Date | Milestone |
|---|---|
| Mar 2019 | **ML Insights GA**: ML anomaly detection (Random Cut Forest), point-and-click forecasting, auto-narratives (Enterprise Edition) |
| Sep 2021 | **QuickSight Q GA**: NL questions → auto-generated visuals, grounded in author-curated **Topics** |
| Nov 2022 | NL forecast queries + "why" questions in Q |
| Sep 2023 | Generative BI authoring preview (build visuals from NL, NL calculated fields) |
| Apr 2024 | **Amazon Q in QuickSight GA**: multi-visual Generative Q&A, **executive summaries**, **data stories**, NL dashboard building |
| Mar 25, 2025 | **Scenarios GA** — agentic what-if analysis ("every employee their own data analyst"; claimed 10x faster than spreadsheet scenario work) |
| Apr 1, 2025 | **Amazon Q in embedded QuickSight GA** (embed Generative Q&A, summaries, NL authoring) |
| **Oct 9, 2025** | **Amazon Quick Suite launch** — QuickSight absorbed into an "agentic AI-powered workspace"; **Amazon Q Business folded in** |
| 2026 H1 | Renamed **Amazon Quick**; web GA + desktop/mobile apps (June 2026); **Quick Autonomous Agents** at AWS Summit 2026 |
| Jul 2026 | AWS puts **Q Business, Kendra, classic Bedrock Agents into maintenance mode** — consolidating on Bedrock + AgentCore + Quick |

## 2. Amazon Quick (Suite) components, 2026

Six components behind one chat interface:
- **Quick Sight** — the surviving BI module: NL queries, dashboard creation, what-if Scenarios, one-click actions from dashboards (create tickets, trigger workflows).
- **Quick Index** — unified knowledge repository (documents, files, app data, warehouses, email) grounding AI answers; effectively Kendra/Q Business retrieval reborn.
- **Quick Research** — deep-research agent: decomposes questions into research plans, gathers from enterprise + external sources, returns cited analysis "in minutes or hours."
- **Quick Flows** — no-code NL-described multi-step workflow automation.
- **Quick Automate** — enterprise agentic process automation: multi-department workflows, **UI agents that autonomously navigate websites**, custom agent building, human-in-the-loop approvals, audit trails.
- **Chat agents + Spaces** — chat over all connected data; custom per-department chat agents; "spaces" = curated contexts combining documents, dashboards, datasets.

**MCP: yes — as a client.** Quick connects to remote MCP servers for **actions** and **data access** (feeding Quick Index). Hosted integrations: Asana, Atlassian, Box, Canva, HubSpot, HuggingFace, Intercom, Linear, Monday, Notion, PagerDuty, Workato, Zapier. Constraints: **remote servers only** (no stdio/local), **no VPC connectivity**, requires Author role + Professional subscription. 50+ built-in connectors + OpenAPI connectivity. No evidence of Quick exposing an MCP *server* for external agents to query its BI data.

**AWS architecture consolidation**: three anchors — **Bedrock** (models/retrieval), **AgentCore** (agent runtime), **Quick** (business-user UX). Quick does transparent **model routing** across Bedrock models (no model choice exposed). Quick Autonomous Agents (Summit 2026): long-running projects, scheduled/triggered events, goal-based outcomes; customers named GoDaddy, NFL, NBA.

## 3. Grounding & the curation burden

- **Topics** remain the semantic backbone for structured NL Q&A: authors curate fields, synonyms, named entities, semantic types, default aggregations. Auto-curation exists but real accuracy requires manual tuning.
- **Verified answers**: author-verified Q&A pinned atop suggestions, with ask counts + helpfulness votes. **Custom instructions** (2024–25) inject domain guidance beyond topic metadata.
- **The burden is documented**: AWS's own guides concede accuracy depends on synonym/entity/verified-question tuning; re:Post threads report persistent Topic errors, accuracy issues, custom instructions ignored.
- **Structural weakness**: Topics are a **parallel semantic artifact maintained separately from dashboards** — datasets, analyses, and Topics can drift. In Quick Suite, unstructured grounding shifts to **Quick Index + Spaces** (RAG with citations) — two grounding systems for two data shapes. (Rill contrast: one code-defined metrics layer serves dashboards and AI alike.)

## 4. AI pricing

**QuickSight era (2024–Oct 2025)**: Reader $3/mo, Author $24/mo (no GenAI); **Reader Pro $20/user/mo** (Generative Q&A, summaries, stories); **Author Pro $50/user/mo** (adds NL authoring, Topics, Scenarios). Historical Q add-on: $250/mo base (~500 questions, $0.50/extra).

**Quick era (Oct 2025 →)**: **Professional $20/user/mo** (chat agents, spaces, Quick Sight consumption, Quick Research, Quick Flows; 25 GB pooled Index/user) · **Enterprise $40/user/mo** (adds authoring of Automate + Quick Sight dashboards; 50 GB/user) · Index overage $5/GB/mo · agentic overage in **"agent hours"**: $6/hr Research, $3/hr Flows/Automate · plus a **flat $250/account/mo AI infrastructure fee** (HN read this as corporate-only signal).

**Embedded AI**: Q in embedded QuickSight GA Apr 2025; priced as **question-capacity packs $250/mo per 500 questions, $0.50/extra**, stacked on session-capacity dashboard pricing — a meaningful cost adder on already-confusing embed pricing.

## 5. Practitioner-reported limitations

- Topics curation effort (accuracy ∝ curation investment); persistent accuracy complaints on re:Post.
- **English-optimal only**; longstanding unresolved language-support asks.
- **Region gaps**: Q in ~16 regions by mid-2025; Scenarios GA'd in just 4 regions; agentic features launch US-first.
- **Rebrand fatigue**: four names in ~18 months; re:Post literally asks "Quick or Quick Suite still?"; Corey Quinn: "Two product states, four prior names, and counting."
- MCP client: remote-only, no VPC — limits enterprise patterns.

## 6. Market perception

- **Gartner MQ ABI 2025: AWS = Challenger** (praised: ecosystem, scalability, cost; cautioned: cloud-confined). **2026: reportedly moved to Leader** on the agentic-first redesign ("rebuilt Quick from the ground up as an agentic platform"), with cautions on AWS lock-in and a smaller practitioner community — *single secondary source (Concord); verify before quoting*. The framing: Leader placement "reflects ambition as much as maturity."
- **Constellation** bullish: Quick is "closer to how enterprises are actually putting AI agents to work than a traditional BI tool"; AWS pushing agentic at infra/platform/app levels.
- **Skeptics**: Corey Quinn's "application graveyard" thesis (WorkMail, WorkDocs, Honeycode, Chime dead; AWS brings "infrastructure-grade muscle to problems that demanded application-grade taste"); Forbes: "AWS kills the AI services it launched just two years ago" (July 2026 retirements).
- **HN on the Quick Suite launch**: "second-rate version of CoPilot… Seems like y'all are just throwing random things at the wall"; confusion vs Q Business; $250/mo fee read as corporate-only.
- **Review-site themes**: pros — cost, pay-as-you-go, serverless, AWS integration, SPICE speed. Cons — "minimum features compared to Power BI and Tableau," "overly simplistic visualizations," limited customization, config friction.
- **"Cheap but limited" in 2026?** Evolving: high AI-axis velocity + the MQ move, but the core viz/UX gap persists in reviews; reputation now ≈ "AWS's agentic workspace bet — promising but unproven, sensible only if all-in on AWS."
- **Scale**: "more than 100,000 customers" (AWS, Mar 2025); named: Docebo, GoDaddy, NFL, NBA, BMW, Siemens, Capital One, Best Western, Volvo. NFL Combine IQ = the flagship embedded-scale story (hundreds of thousands of *concurrent users*).

## 7. Embedded & developer experience (competitive read)

- **Embedded economics = QuickSight's strongest card**: registered readers $0.30/session capped $5/reader/mo; anonymous embedding via capacity pricing from $250/mo per 500 sessions ($0.50/extra, 30-min sessions). Serverless, no per-node licensing — undercuts Tableau/Power BI Embedded at high volume (the NFL story). Wins embed deals on price-at-scale, AWS-native auth/billing, zero ops, embeddable GenAI.
- Loses embed deals on: **pricing opacity** (their own forum is full of session-pricing confusion threads), visualization ceiling, theming limits, and DX.
- **DX = QuickSight's weakest card vs code-first tools**: dashboards are giant machine-generated JSON via `Describe/UpdateAnalysisDefinition`; "Assets-as-Code" exists only as AWS *samples*; no native git — AWS publishes a DIY "BIOps" blog series; Terraform/CFn coverage partial. Vs Looker (git-native LookML, pricier), Superset/Metabase (free, self-host ops). QuickSight has the worst as-code story of the set.

## 8. Rill vs QuickSight signal

- **No direct head-to-head content exists** (no comparison pages on rilldata.com, G2, or blogs). G2 pairs Rill's alternatives with Sigma/Hashboard, not QuickSight.
- Where they actually meet: (a) **embedded analytics for AWS-shop SaaS vendors** (session pricing vs Rill's performance + as-code); (b) **operational/event-scale analytics** where SPICE limits bite; (c) **"AI-ready BI" evaluations** (Quick's chat/MCP-client story vs Rill's MCP-server metrics layer).
- Expect Quick to appear as the "**default incumbent-by-EDP**" in AWS-committed accounts rather than a feature-matched rival.

## 9. AWS strategy context

- Quick is now AWS's **only** business-user AI surface (post Q Business/Kendra retirement) — high strategic priority, high stakes.
- **Procurement gravity**: Quick bills through the AWS console and is **EDP-eligible** — seats burn down committed AWS spend at negotiated discounts; also on AWS Marketplace. "Already paid for under our EDP" is the true competitor.
- Counters that work for independents: **AWS Marketplace listing** (so your spend also retires EDP), decisive performance/DX differentiation, and the stability argument (application graveyard + 4 renames in 18 months = risk for a multi-year BI standardization).
- The agentic-workspace pivot may *vacate* the classic BI conversation: accounts that want excellent dashboards are increasingly poorly served by Quick's direction.
