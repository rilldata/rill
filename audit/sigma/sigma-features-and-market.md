# Sigma — Platform, AI & Market Inventory

> Researched Aug 2026 from help.sigmacomputing.com (raw markdown docs), sigmacomputing.com announcements/blog, BusinessWire/TechTarget/SiliconANGLE, Vendr, G2/Peer Insights. Founded 2014 (Sutter Hill incubation); CEO Mike Palmer; SF HQ.

## 0. Positioning (2026)

**"The AI runtime for business"** — a marked shift from 2023–24's "spreadsheet BI on the cloud warehouse" toward **AI Data Apps** and agentic analytics (Discover → Build → Act → Scale). Architecture: spreadsheet-like workbooks compiling every interaction to **live SQL against the customer's warehouse**; Sigma stores no data (writeback lands in a Sigma-managed schema *inside* the customer's warehouse).

## 1. Core product

- **Warehouses** (11): Snowflake (most complete), Databricks, BigQuery, Redshift, PostgreSQL, AlloyDB, MySQL, Azure SQL, SQL Server 2022, Starburst Galaxy, **ClickHouse (Beta)**. Feature support uneven (input tables/Python/geo functions restricted on lesser platforms). Hosted on AWS/Azure/GCP regions; **no on-prem**.
- **Workbooks**: pages of tables/pivots/charts/controls/UI elements; live edit/draft/publish, real-time multiplayer, version history, custom views, explorations. Separate **Reports** doc type (pixel-perfect paginated PDF, GA late 2025).
- **Data models** (successor to datasets — datasets dead **Sept 15, 2026**): Sigma's semantic layer — reusable source elements, **relationships** (join-on-demand), **metrics** (reusable aggregates, recompute at any grouping; no window/join functions, per-source, don't propagate through joins), RLS/CLS, materialization, warehouse views, parameters, version tagging, ERD, lineage, an **"AI context" field**. **Data models as code** (YAML via REST API, GA Sept 2025) + Sigma CLI. Can query/extend **Snowflake Semantic Views** (Beta) and query **dbt Semantic Layer** metrics.
- **Calculations**: Excel-like formula language (aggregate/window/geo/join/JSON/system functions), pivot ad-hoc calcs, formula bar AI.
- **Input tables (writeback)**: empty/CSV-seeded/linked; validation, protected columns, file-upload columns, row-level audit; needs warehouse write access. **Sigma Tables** (private beta May 27, 2026): shared warehouse-native CRUD tables usable by multiple workbooks/apps/APIs/agents — pitched up to "migrating transactional SaaS apps into Sigma."
- **Compute/caching**: everything live in the warehouse; mitigation = **materialization** (write-back tables; Snowflake dynamic tables with incremental refresh, skip-on-unchanged, auto-pause) + warehouse views. **Python** elements on Snowflake/Databricks only.

## 2. Visualization, apps & delivery

- Charts: bar/line/area/scatter/pie/combo/KPI/**gauge/funnel/Sankey/waterfall/box/maps** (point/region); reference marks, custom tooltips, legends-as-filters.
- **Actions framework**: condition + trigger + effect sequences with if/else; effects incl. input-table CRUD, stored-procedure and **API calls**, modals/drawers, **call/reset a Sigma agent**; scheduled actions previewed. UI elements (containers, modals, forms, tabs, plugins) = the **AI Data Apps** substrate.
- **Exports/schedules** — destinations: email (incl. **per-recipient bursts**), Slack, Teams, SharePoint, Google Sheets/Drive, S3/Azure/GCS, webhooks; formats PDF/Excel/CSV/JSON/PNG. **Alerts = conditional exports** (Act tier+).
- Versioning: history + restore, **version tags** (protected, approval flows, tag-based dev→staging→prod, per-tag materialization). No native git for workbooks; as-code via API/YAML + CLI + published **agent skills**; docs expose llms.txt and an MCP server.

## 3. AI features (chronology & taxonomy)

- **Dec 12, 2024 — "Ask Sigma"** (private beta): NL "agentic AI analyst" showing its editable multi-step plan (source→filters→formulas), opens into a workbook.
- **May 20, 2025**: embeddable Ask Sigma; AI usage/admin dashboard; Snowflake/Databricks-hosted models; embedded writeback.
- **Sept 10, 2025**: **AI Builder** (NL app creation), Ask Sigma Discover (proactive insights), Workbook Suggestions (answer with vetted existing workbooks), **Chat Element** (custom in-workbook agents), MCP extensibility — mostly "coming soon" at announce.
- **Dec 2025**: **AI Query GA** (warehouse LLM functions — `AI_COMPLETE`/`ai_query` — as Sigma formulas/columns), AI Builder private beta, MCP Client (coming), MCP Server (coming), Cortex Agents in Ask Sigma (coming), **Sigma Tenants GA**.
- **Mar 5, 2026 — Workflow 2026** (first user conference, sold out): previews of conversational + autonomous **Sigma Agents**, Sigma MCP Server, **Sigma Public** (public AI-data-apps playground), scheduled actions, NL→data models. Internal demo claim: "clear path to fully replacing Salesforce internally."
- **Apr 2, 2026**: **Sigma Agents ship (beta)** — "fastest-adopted product in company history" per the Series E release.
- Recognized: **Sample Vendor, 2026 Gartner Hype Cycle for Agentic AI**.

**Current taxonomy (docs):**
- **Sigma Assistant** (the Ask Sigma surface): from Home or workbooks; conversational; **Analysis breakdown** (step-by-step decision logic); "Explore" any answer into a workbook; thumbs feedback → admin AI-usage dashboard. Grounding: prefers **endorsed/badged data models, metrics, descriptions, relationships, usage frequency**; admin-curated or auto-indexed (top-2,000 sources, Beta). Respects RLS/CLS. Limits: relationship-defined joins only; published model versions only.
- **Assistant in workbook** (Beta): analyze + build modes (build dashboards/apps from prompts).
- **Sigma Agents** (Beta; **premium SKU post-beta**): workbook-scoped agents = instructions + data sources + tools (deterministic **actions**, **MCP servers**, **warehouse agents**); Chat Element UX or **scheduled/autonomous** runs. Use cases: approval workflows, sales copilots drafting Slack follow-ups, churn notifications, anomaly detection, scenario analysis via Python/stored procedures.
- **Warehouse agents** (Beta): first-class invocation of **Snowflake Cortex Agents** and **Databricks Genie Spaces** — governed warehouse-side; note Sigma RLS does *not* apply to warehouse-agent answers.
- **AI Query/columns**, formula assistant, "Explain this chart," input-table AI fill, file columns for unstructured AI workflows.
- **Sigma MCP Server** (live): OAuth per-org endpoint; listed in **Claude Connectors Directory**, ChatGPT marketplace plugin, Codex, Cursor, Cortex Code; tools for search/describe/query with user-inherited permissions; MCP queries warehouse-tagged for cost attribution. **MCP Client** announced (external systems into Assistant/agents; mTLS documented).
- **Models (BYO-provider)**: warehouse-hosted (Cortex: Claude Sonnet 5 / GPT-5.1; Databricks: Sonnet 4.5; BigQuery: Gemini 3.6 Flash) or external (OpenAI GPT-5.4, Azure OpenAI, Anthropic, Gemini, Bedrock); per-provider embeddings; strict data-handling stance; per-feature permission gating; admin usage + feedback dashboards.
- **Accuracy story**: no published benchmarks — rests on transparency (editable plan), endorsed-model grounding, deterministic action tools, feedback loops.
- **AI pricing**: Assistant gated by license tier; **agents premium-priced**; LLM/warehouse-AI costs pass through to the customer's own accounts (hence Sigma's own **AI cost-monitoring templates** for Claude/OpenAI/Cortex spend).

## 4. Embedded analytics

Public embeds + **secure embeds** (server-generated one-time JWT URLs); JS Embed API + open-source **React Embed SDK**. Multi-tenancy: teams + user attributes in JWT claims (RLS, dynamic warehouse roles) and **Sigma Tenants** (GA late 2025: fully isolated workspaces, API provisioning, delegated admin, per-tenant usage/audit). White-label: full theming, custom fonts, **localization** (UI + content translation files), custom domains. **Embedded self-service**: explorations, saved views, **embedded Ask Sigma**, **embedded writeback** — "embed an app, not just a dashboard." Version tags as embed environments. Named: Persona ("8× adoption"), G2, DoorDash, Affirm, Workato, Blackstone. Embedded pricing = separate negotiated book.

## 5. Enterprise & governance

RLS (user attributes on data models) + CLS grants, both enforced for AI and writeback; granular account types × license tiers; endorsed badges; data catalog. SAML SSO, **SCIM**, OAuth **connection-level passthrough** to Snowflake/Databricks/BigQuery (users inherit warehouse policies), service accounts, IP allowlists. **Audit-logs connection** (queryable in Sigma, export to S3/Azure/GCS). dbt: metadata integration + dbt SL query integration. Deep Snowflake (Partner Connect, Cortex, Semantic Views, dynamic tables) and Databricks ties (**Databricks Ventures invested in the Series E**, as did ServiceNow Ventures and Workday Ventures — app-platform ambitions).

## 6. Company & market

- **Funding ≈ $742M**: Series C $300M (Dec 2021, ~$1B) → Series D $200M (May 2024, ~$1.5B) → **Series E $80M (May 18, 2026, Princeville; Databricks/ServiceNow/Workday Ventures) at $3B**. J.P. Morgan equity investor at the $100M-ARR mark.
- **$200M ARR (Apr 2026), 100%+ YoY**; 1.1M+ new active users/yr; 2,000+ customers incl. Fortune 10; 2M+ workbooks/apps. Customers: JPMorgan Chase, AMD, Duolingo, Colgate, DraftKings, DoorDash, Affirm, Blackstone, Persona, G2.
- **Gartner MQ ABI**: first inclusion 2025 as **Niche Player** ("highest ability-to-execute debut in 15 years"); included again 2026 (Sigma avoids naming the quadrant). Peer Insights 4.8/5. G2 ~4.4–4.7 (~525 reviews). Hype Cycle Agentic AI Sample Vendor.
- **Pricing** (sales-led): Vendr median ≈ **$68K/yr** ($20–60K small → $150–400K+ enterprise); license model **View / Act / Analyze / Build**; viewers cheap/free in some deals; creators ~$1.2–5K/yr; embedded + agents priced separately.
- **Complaints**: **#1 warehouse compute passthrough** ("Sigma's subscription is predictable, your warehouse bill is not" — compute can rival the license); spreadsheet-paradigm limits on messy schemas + optional governance ("metric divergence within six months"); warehouse-only (no APIs/Mongo/on-prem); performance = customer's warehouse tuning; opaque pricing; viz depth trails Tableau.
- **Momentum**: quarterly launches (May/Sept/Dec 2025, Mar–May 2026) + weekly release notes; $100M→$200M ARR in ~a year; first user conference sold out.
- **Posture**: "legacy BI" framing vs Tableau/Power BI/Looker/Qlik; vs Looker "no LookML needed"; vs Tableau "live warehouse, no extracts"; closest architectural rival per 2026 guides: **Omni** (Sigma = spreadsheet UX + writeback + apps; Omni = Looker-style modeling).

## Key URLs
docs: help.sigmacomputing.com (llms.txt, `.md` suffix) · Series E: sigmacomputing.com/resources/announcements/series-e · $200M ARR: /resources/announcements/200m-arr · Dec 2025 launch: /blog/december-2025-product-launch · Workflow 2026: /blog/workflow-2026-insights-from-sigmas-first-user-conference · Sigma Tables: /blog/introducing-sigma-tables · MCP docs: help.sigmacomputing.com/docs/use-sigma-mcp-server · Embed SDK: github.com/sigmacomputing/embed-sdk · Vendr: vendr.com/marketplace/sigma
