# Google Looker — Platform, AI & Market Inventory

> Researched Aug 2026 from cloud.google.com/looker docs + pricing + release notes, Next '25/'26 coverage, G2, third-party pricing analyses. Naming note: **Looker Studio was renamed back to "Data Studio" in April 2026**; "Looker (Google Cloud core)" and "Looker (original)" both still exist.

## 1. Semantic layer (LookML)

- Declarative, git-versioned modeling: models/views/explores; joins with **symmetric aggregates** (fan-out-safe); derived tables/**PDTs** (trigger-based, incremental) + **aggregate awareness** (`aggregate_table` transparent routing); `extends` + refinements for layering; `access_grant`/`required_access_grants`; **Liquid templating** (`_user_attributes`, `{% condition %}`, parameterized SQL).
- **Git-native workflow**: browser IDE, per-dev branches, PR integration, LookML + content validators. **Looker CI GA at Next '26** (built from the **Spectacles acquisition**; standalone Spectacles dies Nov 30, 2026). **VS Code extension (Preview 2026)**: local LookML dev in VS Code/Cursor — Google's release note explicitly names **Claude Code** — with AI-assisted LookML via the Looker MCP server.
- **"Universal semantic layer" status**: **Looker Modeler (2023) was never launched and appears shelved.** What survived: Connected Sheets against Explores, a Looker Studio/Data Studio connector, Power BI/Tableau via Open SQL Interface (JDBC/Avatica), and now the **managed MCP server** as the third-party-AI consumption path. **In-database semantic models (Preview 2026)**: integration with BigQuery Graph and Snowflake Semantic Views. Google joined **OSI** as a member in the later expansion; LookML itself is not open-sourced.
- Headline accuracy claim since Next '25: LookML grounding cuts gen-AI NL-query errors **"by up to two-thirds"** vs raw-schema access.

## 2. BI product

- **Explores** (governed self-service: field picker, pivots, table calcs, custom fields, drill, merged results), Looks, dashboards (user + LookML), boards, folders; custom viz (Marketplace or self-hosted JS); cross-filtering, cascading filters, drill overlays.
- Deliveries: schedules to email, Slack, **S3, SFTP, webhooks** (Action Hub); threshold alerts; **June 2026**: Gemini "key driver analysis" insights inside alert notifications.
- **Next '26 UI wave**: tabbed dashboards GA, **paginated reporting GA**, modernized AI-driven drag-drop Explore UI (Preview), **self-service Explores GA** (blend personal CSV/Excel uploads with governed data — LookML auto-generated). Dashboard tile limits (perf guardrail) GA Aug 2026. Mobile apps exist but thin.
- **Data Studio (ex-Looker Studio)**: free, 800+ connectors, 1.2M+ MAU (end-2024). **Pro $9/user/month *per project*** (per-project multiplier is a known agency gotcha) — adds workspaces, governance, Gemini features. **April 2026 rebrand back to Data Studio**; positioning re-split: Data Studio = personal/ad-hoc, Looker = governed enterprise BI. The "Studio-in-Looker"/Looker reports unification was **killed July 13, 2026**; complimentary Pro licenses with Looker instances ended Aug 1, 2026.

## 3. Performance / caching

Live-query only against the warehouse; caching via datagroups; PDTs + incremental PDTs + aggregate awareness are the acceleration answer. **The "slow dashboards" reputation persists** — ~38% of G2 reviewers flag performance (slow dashboards, timeouts, slow renders on large data). Google's 2026 response is admin guardrails (Performance Center, tile limits), not an engine change.

## 4. AI features (Gemini in Looker)

| Date | Milestone |
|---|---|
| Apr 2024 | Gemini in Looker announced (Next '24) |
| 2024–25 | Preview wave: Conversational Analytics, Formula Assistant, **Automatic Slide Generation** (live-linked Google Slides summaries), LookML Assistant (NL→LookML), Visualization Assistant |
| Apr 2025 (Next '25) | Conversational Analytics opened to **all platform users at no additional cost**; "How was this calculated?" transparency |
| Aug 2025 | **Conversational Analytics API** public preview (embed chat-with-data in customer apps) |
| Nov 2025 | **Conversational Analytics GA** (Looker 25.18+) |
| 2025–26 | **Code Interpreter** (NL → generated-and-executed Python: forecasting, anomaly detection) → renamed "Advanced Analytics," **GA 2026** |
| May 2026 | **Looker-managed MCP server Preview** — fully managed, admin-controlled tool allowlists, usage in System Activity + Cloud Audit Logs; positioned as "Open BI" |
| Jun 2026 | **CA API GA** (BigQuery + Looker) + **A2A protocol support (Preview)**; CMEK/VPC-SC previews |
| Apr 2026 (Next '26) | **Dashboard Agents** (Preview: Q&A + summaries in dashboards), **Agentic Workflows** (Preview: NL-created agents that monitor metrics, spot anomalies, notify, recommend actions), Expression Assistant, Insight Assistant, Visualization Assistant GA, AI Quick Starts |
| Mid-2026 | **Verified/golden queries** (Preview: curated NL→Explore-query pairs), **CA observability** (Preview: engagement, token usage, success rates, feedback in System Activity) |

- Also: MCP Toolbox for Databases gained Looker support Jul 2025 (open-source path for Claude/Cursor/Gemini CLI). **Gemini Enterprise** (ex-Agentspace, folded Oct 2025): Looker CA agents publishable into Gemini Enterprise chat (Preview, 26.8).
- **AI pricing**: metered in **"data tokens"** pooled per instance — included: Standard 60M in/1.2M out, Enterprise 300M/6M, Embed 1.2B/24M monthly. **Overage $3/1M input, $20/1M output — billing starts Oct 1, 2026** (free fair-use until Sep 30, 2026). Data Studio Gemini features gated behind Pro.

## 5. Embedded analytics

Private embedding, **SSO/signed embedding** (user attributes → RLS), Embed SDK (JS, two-way events), cookieless embed, API-only backends; **Embed edition** = signed embed + white-label + 500K query-API calls/mo. **Embedded Conversational Analytics GA 2026** (iframe or CA API). Extension framework (full JS apps in Looker). Pricing reputation: **~$180K–198K/yr list for Embed** — "most companies with embedded use cases skip Looker due to price" (Holistics, Toucan).

## 6. APIs & ecosystem

Looker API 4.0 (SDKs: Python/TS/Kotlin/Swift/R/Go/C#); Action Hub (custom actions); Marketplace + Blocks (customized via refinements); **no first-party dbt integration** (ecosystem fills it: dbt2looker, vendor sync layers); BigQuery depth = stated #1 strategy pillar at Next '26.

## 7. Enterprise / governance

Roles = permission sets × model sets; user attributes for RLS (incl. SSO embed); **System Activity** audit explores + Cloud Audit Logs (core); SAML/OIDC/LDAP/Google IAM; CMEK, VPC-SC, private IP (core). **Deployment**: core is GCP-hosted only; customer-hosted survives only on Looker (original), unloved; consolidation still incomplete.

## 8. Pricing

Official page is "call sales"; verified structure: **Standard** (<50 users, 1K API calls/mo), **Enterprise** (100K), **Embed** (500K); each includes 10 Standard + 2 Developer users; 1–3-year terms. Third-party figures: Standard ~$60–67K/yr, Enterprise ~$132K/yr, Embed ~$180–198K/yr platform fees; Developer ~$125/user/mo, Standard user ~$800/user/yr, Viewer ~$400/user/yr. Data Studio free; Pro $9/user/project/mo. **Billed through GCP → counts toward GCP committed spend** (EA/marketplace private offers; BigQuery bundling standard motion). Realistic enterprise entry ~$50K–100K+/yr.

## 9. Market position & sentiment

- **Gartner MQ ABI: Leader 2024, 2025, 2026** (third consecutive) — cited: "strongest technical semantic model in BI," API-first composability, Gemini. 2026 Leaders: Microsoft, Salesforce/Tableau, Google, Qlik, ThoughtSpot.
- Market share ~3.7% of BI category (~18,000 companies); dwarfed by Studio's free tier by user count.
- Customers publicized 2026: YouTube, Telenor, Allo Fiber (agents in production); historical: Deliveroo, Lyft, Etsy.
- **Post-acquisition narrative** (2024–26, from migrating customers + rivals): slowed product development, degraded support, rigid modeling workflow, premium pricing. **Migration wave is real**: Omni (ex-Looker CPO; LookML converters; BuzzFeed migrated in <3 months after 8 years; acquired Explo Oct 2025), Lightdash (OSS LookML successor on dbt), ThoughtSpot/Sigma/Hex on alternatives lists.
- Complaints (G2 2025–26): LookML learning curve, slow dashboards (~38% of reviewers), limited viz/calc flexibility vs Tableau/PowerBI, cost for small teams.
- **Momentum counter-narrative**: AI cadence accelerated sharply (CA GA → CA API GA + A2A → managed MCP → Next '26 agentic wave), while Google **pruned non-AI bets** (Looker reports killed, Modeler shelved, Studio re-split as Data Studio). Portfolio now: Looker = governed agentic BI on LookML; Data Studio = free ad-hoc. **AI monetization (token overage) begins Oct 1, 2026.**
