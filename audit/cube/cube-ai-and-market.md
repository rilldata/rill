# Cube — AI Platform ("Agentic Analytics"), Market Position & Landscape

> Researched Aug 2026 from public web sources (cube.dev blog/docs/pricing, press, G2, third-party comparisons). Note: "Cube" (cube.dev, semantic layer) is often conflated with "Cube Software" (FP&A tool); the latter is excluded. Companion file `cube-features-core.md` covers the semantic layer/API/caching platform.

## 1. Cube D3 → "Cube Agentic Analytics"

### Timeline & framing
- **June 2, 2025**: "Cube Launches D3: The First Agentic Analytics Platform Built on a Universal Semantic Layer" (D3 = "data in cube"). Gradual access through summer 2025.
- **October 29, 2025**: GA — the "D3" name is dropped; semantic layer + agentic products unified under one UI and brand: **"Cube Agentic Analytics."**
- Homepage tagline (Aug 2026): *"The agentic analytics platform built on a semantic layer. AI answers your team — and your customers — can trust. Grounded in your semantic model, governed end to end."*
- Launch coverage was trade press (BigDATAwire, DBTA, insideAI News) + wire pickup; no dedicated TechCrunch/VentureBeat story found.

### Architecture
- Core thesis: LLMs over raw tables hallucinate; grounding in a governed semantic layer fixes it ("Without this governed, business-friendly layer, AI agents querying raw data directly are prone to errors, miscalculations, and 'hallucinations'").
- Analytics decomposed into **three AI-safeguarded layers** (compilation checks, linters, best-practice prompts, documentation, memory retrieval): (1) building the semantic layer; (2) building **Semantic SQL** queries on it; (3) creating visualizations from Semantic SQL results.
- **Semantic SQL** is the pivotal mechanism: the AI writes SQL against semantic definitions (not raw tables), can add ad-hoc calculations, and every query is reviewable/editable visually or as SQL.
- Governance constructs: **Agent Spaces** (agent groupings with role-based permissions), **Rules** (always-rules + agent-requested rules), **Certified Queries** (admin-approved SQL agents may reuse), **BYOM** (bring-your-own LLM, multi-LLM). Marketed "**AI Context Layer**" = semantic layer + business context + adaptive learning.
- Agent access protocols: **MCP** and **A2A** (agent-to-agent), incl. an "agent-to-agent capable via MCP" Analytics Chat API.

### Product surfaces at GA
1. **Analytics Chat** — conversational interface; multi-query analyses with summaries and links back to the semantic model.
2. **Workbooks** — AI-assisted exploration combining point-and-click, calculations, SQL, and **Python**.
3. **Dashboards** — presentation layer (tables, charts, KPI cards, custom HTML/markdown); added at GA.
4. **Semantic Model Agent ("AI Data Engineer")** — auto-builds/improves models from warehouse tables, **converts LookML projects**, manages measures/dimensions/joins/SQL transforms.
5. **AI Data Analyst** — NL querying producing Semantic SQL, visualizations, "data apps."
6. **Embedded Agentic Analytics** — chat, workbooks, dashboards embeddable via iframes; **Analytics Chat API**; **Creator Mode** (customers build their own workbooks/dashboards inside your app); white-labeled ("your colors, your branding, your agent name").

Demos: full-platform demo (youtube gvDy37p7AZ0), Dec 2025 walkthrough (n-U1FvHqE2Q), "warehouse to embedded AI in 39 minutes" founders demo (7ZQGGepDjUQ).

### Reception & reported limitations
- G2 snippet: "the AI analyst can be a little slow when retrieving information," though "the MCP makes up for it."
- Marquee proof point: **"Brex chose Cube over dbt Semantic Layer and LookML — the semantic layer is what makes the AI useful at scale."**
- Mitzu critique: models are hand-authored and need ongoing maintenance; because the LLM authors SQL, "methodology errors on funnels, retention and journeys are still possible"; semantic shape doesn't natively express product-analytics primitives.
- Luzmo (Aug 2026): Cube "sits firmly in the headless camp and has been adding pieces" of embedded analytics; flags the buyer risk of the semantic layer "becoming a separate platform project."
- Cube's own "Cube Cloud, Headless No More?" post (June 2025) pre-empts channel conflict with its BI partners (Tableau, Power BI, Superset…) — an acknowledged tension of shipping a first-party BI surface.
- Little organic Reddit/HN discussion of D3; reception evidence is mostly vendor-published + trade press + G2.

## 2. Pricing & packaging (cube.dev/pricing, Aug 2026)

- **Free**: forever-free; ≤5 workbooks; 1,000 daily queries; limited agent requests with full feature access.
- **Starter**: **$40/developer/mo**; unlimited workbooks; 10K–50K daily queries; extended agent limits + premium LLM access; Cube Store caching; Semantic Layer Sync.
- **Premium**: **$80/developer/mo** + consumer seats — **Explorer $40/user/mo**, **Viewer $20/user/mo**; unlimited queries; embedded dashboards + analytics chat; multi-cluster $1.20/hr/cluster; 99.95% SLA.
- **Enterprise**: custom; 99.99% SLA; single-tenant/BYOC; custom LLM; SAML SSO; **DAX API for Power BI**.
- Principle stated at GA: seat-based for direct access; **embedded stays compute-based**; AI/agent usage bundled into tiers (no separate credits).

## 3. Company scale & funding

- Launch-PR metrics: installed on **90,000 servers**, **~5M users**, **20% of the Fortune 1000** as customers.
- Funding: $6.2M seed (2020) → $15.5M Series A (2021, Decibel) → **$25M (June 2024, Databricks Ventures joining)** = **$48M total**. 4x new-customer growth, 3x bookings 2023→2024. No 2025–2026 raise found.
- Contrast: Rill has raised ~$13.8M (seed-stage, latest Seed II July 2025).

## 4. Rill's public positioning (for reference)

- Homepage: **"The fastest business intelligence tool for humans and agents"**; banner "Build fast agentic analytics with Claude Code and Rill."
- Pillars: define-once semantic layer; branded embedded analytics; NL Q&A with "visual, verifiable" answers (URL citations — "the URL an agent cites is the same URL a human opens"); **AI anomaly detection / "Explain"** (root-cause summaries on metric shifts); BI-as-code with Claude/Cursor; Canvas dashboards.
- Pricing (Aug 2026): Starter **$20/seat/mo** (≤20 seats, 3 projects, embed previews only, 1M AI tokens/seat/mo, ≤10GB managed DB, $0.15/CU-hr) · Growth **$30/seat/mo** (≤100 seats, 10 projects, **embedded $500/mo up to 100 active embedded seats**, 2M tokens/seat + 200k/embedded seat, BYO AI model, ≤1TB, $0.20/CU-hr) · Enterprise custom (unlimited embedded seats, VPC peering). $250 starting usage credit on all plans.
- Packaging contrast vs Cube: Rill meters **AI as token quotas per seat**; lets customers **BYO database (ClickHouse, MotherDuck, Snowflake…) and BYO AI model** below enterprise tier; flat embedded pricing vs Cube's compute-based embedded.
- 2025 ships (year-in-review): AI anomaly detection, AI Chat + MCP, Canvas, BYO ClickHouse/MotherDuck (zero-copy), embed iframe API, public API docs. Agentic-architecture blog (May 2026) describes the router/analyst/developer agent design, "golden completions, not vibes" evals, MCP-native runtime.
- ClickHouse relationship: no formal partnership found, but deep co-marketing (ClickHouse's own "Rill and ClickHouse: real-time operational BI for a metered world" post, joint meetups, dlt+ClickHouse+Rill content). Similar with MotherDuck.
- Heritage: founded 2020 by Michael Driscoll; team previously built Metamarkets (→ Snap; database open-sourced as Apache Druid).

## 5. Head-to-head perception

- **Categorization gap**: Cube appears in essentially every 2026 semantic-layer roundup (Bruin, Holistics, Atlan, Promethium, Dawiso). **Rill is absent from most of them**, categorized instead as BI / operational BI / BI-as-code.
- Clearest third-party framing (Luzmo, Aug 2026): Rill is the middle path — "BI-as-code… keeps the code-first workflow that draws people to Cube, **but ships an interface rather than expecting you to build one**."
- **Cube does not acknowledge Rill**: its "Best BI Tools 2026" article lists Looker, Power BI, Tableau, Sigma, ThoughtSpot, Hex, Metabase, Sisense, GoodData, Databricks, Snowflake — no Rill.
- Convergence: since GA, Cube ships chat/workbooks/dashboards/embedded — a direct competitor for the analytics *surface* including embedded (Rill's commercial wedge); Rill markets its metrics layer + MCP as a semantic layer for agents — entering Cube's narrative from the other side. Near-identical language on both sites.
- Sharpest technical contrast in public materials: Cube = warehouse-agnostic middleware with caching/pre-aggregations (Cube Store); Rill = semantic layer coupled to a fast OLAP engine (DuckDB embedded, live ClickHouse/MotherDuck/Druid/Pinot) for sub-second dashboards.
- Community signal: Rill praise centers on speed ("even though we have Looker, analysts prefer to do everything in Rill because it's fast" — Liftoff); Cube's on AI grounding (Brex; Delphi's 100%-accuracy runs over Cube's layer). Cube has substantial G2 presence; Rill's is thin (named testimonials/case studies: Emodo, Liftoff, Angel Studios, Syft, NOCD, Sabio).

## 6. Broader AI-BI landscape (pressure on both)

- **dbt Labs**: Coalesce Oct 2025 — Fusion engine, four AI agents (incl. Analyst agent in now-GA **dbt Insights**), MCP server, **MetricFlow open-sourced (Apache 2.0)** co-maintained with Snowflake/Salesforce; **dbt + Fivetran merger** bundles semantic layer + consumption into the default pipeline tool.
- **Snowflake**: **Semantic Views as native DB objects** (SQL querying GA Mar 2026), Cortex Analyst (~90%+ claimed accuracy), Snowflake Intelligence, verified queries. Makes third-party semantic layers look redundant for Snowflake-committed shops — existential squeeze on Cube's middleware position.
- **Databricks**: AI/BI Genie (84.67% in runQL's 2025 benchmark), **Metric Views GA early 2026**. (Databricks Ventures is a Cube investor — hedged.)
- **Google Looker**: Conversational Analytics GA grounded in LookML + a **Conversational Analytics API** (GA for BigQuery/Looker) aimed at embedding governed NL agents — attacks both Cube's Analytics Chat API and Rill's embed story.
- **ThoughtSpot Spotter**: strong brand; scored **54.67%** in runQL's benchmark — the proof point both Cube and Rill use that a chat UI ≠ trustworthy answers.
- **Omni**: BI + built-in semantic layer + AI; **acquired Explo (embedded) Oct 2025**; **$120M Series C at $1.5B (April 2026)** on ~$30M ARR. Packages Cube's governance story + Rill's finished-interface story + embedded, with far more capital.
- **Hex**: $70M Series C (May 2025); notebook-first AI workspace (Reddit, Anthropic, Figma, NBA).
- **AtScale**: the other pure universal semantic layer; 2025 GigaOm Radar Leader; strong in Power BI/Excel/DAX estates.
- **Select Star**: catalog repositioning as "context for AI agents"; OSI co-leader.
- **Startups**: Zenlytic, TextQL, Dot, Querio, Honeydew, Mitzu, runQL, Promethium — collectively commoditizing the "chat with governed data" pitch.
- **Open Semantic Interchange (OSI / "Apache Ossie")**: launched Sept 2025 by Snowflake, Salesforce, dbt + **Cube, Omni, Hex, Sigma, ThoughtSpot, Select Star, Alation, Atlan, BlackRock, Mistral AI**; v0.1 spec, incubating at Apache. **Rill is not a listed member** — a standards-visibility gap.

## 7. What buyers care about (2026 evaluation criteria)

1. **AI grounding/governance** — #1 criterion: do AI answers rest on governed metrics or raw tables (Luzmo; Gartner Agentic Analytics Market Guide: guardrails now "explicit requirements").
2. **NL accuracy with evidence** — cited numbers: raw text-to-SQL ~16.7% vs ~83% with a semantic layer (Cube-commissioned research); Delphi 100%-accuracy runs; runQL 2025 (Cortex 88.67%, Genie 84.67%, Spotter 54.67%). Buyers increasingly ask for **eval harnesses**.
3. **Cost legibility** — dev vs viewer vs consumption seats; AI metering (Rill tokens vs Cube bundling); embedded per-tenant economics; warehouse compute generated by agents.
4. **Latency/interactivity** — caching/pre-agg strategy "makes or breaks an embedded experience"; Rill's founding pitch; the one recurring Cube AI complaint (slow AI analyst).
5. **Embedding & frontend ownership** — build vs configure; multi-tenant RLS; white-label; embeddable chat/agents.
6. **Self-hosting/BYO/data gravity** — OSS core (both), BYOC (Cube Enterprise), BYO DB + LLM (Rill Growth+), zero-copy vs data movement.
7. **Interoperability/lock-in** — OSI membership, MetricFlow Apache 2.0, LookML conversion, MCP support (now table stakes).

## 8. Bottom line

- **Cube** converted semantic-layer credibility into a full-stack agentic analytics platform (Oct 2025 GA), accepting collision with its former BI partners. Moats: install base (90k servers, 20% of F1000, Brex), Semantic SQL governance machinery. Exposure: platform-native semantics (Snowflake/Databricks), dbt+Fivetran bundling, and better-funded surface rivals (Omni).
- **Rill** attacks from the opposite end: open-source, OLAP-fast, BI-as-code "for humans and agents," MCP-native multi-agent runtime, transparent pricing, ClickHouse/MotherDuck alignment. Weaknesses: ~$13.8M funding, thin analyst/review presence, absence from semantic-layer roundups and OSI, and now direct overlap with a larger Cube in embedded, chat, and agent access.
