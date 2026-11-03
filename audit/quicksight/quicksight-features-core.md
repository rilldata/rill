# Amazon QuickSight / Quick — Core Platform Inventory

> Researched Aug 2026 from docs.aws.amazon.com/quick + /quicksight, aws.amazon.com/quick(sight)/pricing, AWS What's New feed, re:Invent 2025 coverage, G2. **Naming**: QuickSight (GA Nov 2016) → **Amazon Quick Suite** (Oct 9, 2025) → **Amazon Quick** (2026 H1); BI module lives on as "Quick Sight"; APIs/CloudFormation keep the `QuickSight` namespace. Companion file `quicksight-ai-and-market.md` covers AI features and market perception.

## 0. Product structure (2026)

Components of Amazon Quick: **Quick Sight** (the BI product: dashboards, analyses, datasets, embedding) · **Quick Research** (agentic deep research, cited reports) · **Quick Flows** (NL workflow automations) · **Quick Automate** (multi-step agentic process automation) · **Quick Index** (org knowledge base grounding AI answers; SharePoint/S3/Confluence sources) · **Apps in Quick** (NL-built web apps, preview) · **Quick chat** (unifying chat with custom agents/personas). Surfaces: web, desktop app, Chrome/Slack/Teams/M365 extensions. External connectivity via MCP (client) + OpenAPI.

Transition mechanics for QuickSight customers: no migration; dashboards/datasets/permissions/APIs unchanged; compliance carried over (SOC, **HIPAA**, ISO 2700x, GDPR, **FedRAMP**); agentic features initially in only 4 regions; embed footer became "Powered by Amazon Quick Suite" (removable via brand customization / annual capacity commitment); admin custom permissions to disable AI features per account/role/user.

## 1. Data & connectivity

- **Sources**: Athena, Aurora, Glue Data Catalog, OpenSearch, Redshift (+Spectrum), S3 (+S3 Tables, S3 Analytics), Spark, Impala, IoT Analytics, Databricks, Exasol, BigQuery, MariaDB, SQL Server, MySQL, Oracle, PostgreSQL, Presto, Snowflake, Starburst, Trino, Teradata, Timestream; on-prem via VPC connections. Files: CSV/TSV, logs, JSON (auto-flatten), XLSX. SaaS (BI side): Jira, ServiceNow, Adobe Analytics, GitHub, Salesforce.
- **Quick-chat/agent connectors (2026)**: Slack, M365, Google, QuickBooks + 16 added Aug 2026 (Adobe, Figma, Shopify, Smartsheet, Snowflake, WhatsApp, Zapier, ZoomInfo, Moody's, D&B…); Snowflake Cortex via managed MCP (Jun 2026); New Relic agents (May 2026); VPC connectivity for private MCP servers (Jun 2026).
- **SPICE** (in-memory engine): per-region account-shared capacity, $0.38/GB/mo (10 GB included per Author); import vs direct-query per dataset. Quotas: Enterprise **2B rows / 2TB per dataset** (docs; pricing page still says 1B/1TB), 2,000 columns; strings cost 24 bytes + length (heavy string tax). Refresh: full + **incremental** (SQL sources only, look-back window); schedules down to **every 15 minutes**; ingestion APIs. Direct query: 2-minute timeout.
- Dataset-as-source (child datasets, RLS inherited); cross-account Athena via role chaining (May 2026); **Agentic Catalog Experience** preview (Jul 2026) over Glue Data Catalog + Databricks Unity Catalog.
- **No transformation layer**: data prep = joins/calculated fields/filters at dataset level; anything heavier is Glue/dbt/warehouse territory.

## 2. Semantic layer / modeling

- **Datasets** are the semantic unit: cross-source joins, calculated fields, filters, field folders, friendly names, geo roles, RLS/CLS.
- **Calculated fields**: large function library — aggregates (+`...If` conditional variants, `periodToDate*` family), date/numeric/string, table calculations (running/window/rank/percentile).
- **Level-aware calculations**: LAC-A (aggregate at explicit group-by level, `$visualDimensions` system parameter) and LAC-W (window at level).
- **Parameters**: analysis parameters with controls; **dataset parameters** pushed into direct queries (per-tenant slicing).
- **Topics** (NL semantic layer for Q): curated synonyms/named entities/formatting/relevance — maintained separately from dashboards. **Multi-dataset topics (Aug 2026)**: relationships across datasets, automated runtime joins, one semantic model for dashboards + NL Q&A, RLS/CLS-aware — AWS converging toward a real semantic layer.

## 3. Visualization & dashboards

- **30+ visual types**: AutoGraph, bar (all variants), box plot, combo, custom visual content (web/image/video embed), donut, funnel, **gauge**, heat map, **Highcharts visuals** (code-defined custom charts), histogram, KPI (sparklines), **layer maps / points / filled maps / geospatial heatmap**, line, **small multiples**, pie, pivot table, **radar**, **Sankey**, scatter, table, text box, **tree map**, **waterfall**, **word cloud**.
- Limits: 50 datasets/dashboard, 20 sheets, 30 visuals/sheet; per-visual data-point caps (10k bar/line points, "other" bucketing).
- **Analyses** (editable) → published **Dashboards** (read-only interactive); **Stories** (AI narrative docs). Layouts: tiled grid, **free-form** (absolute positioning, overlap, conditional visibility), classic. Themes (custom, API-managed, runtime for embeds).
- Interactivity: filter/navigation/URL **actions**, drill-down hierarchies, cascading filters, cross-sheet filters (2024), custom tooltips, conditional formatting, frozen columns, custom totals, bookmarks.
- **Paginated reports** (pixel-perfect, multi-page; Nov 2022): $500/mo per 500 report units (unit = 100 pages/100MB), annual $24,000/yr for 4,000 units/mo.
- Exports: PDF/CSV/Excel incl. programmatic Snapshot Export APIs.
- **Generate Analysis GA (May 2026)**: full dashboards from an NL prompt over ≤3 datasets, editable plan, generated calc fields.

## 4. Consumption & collaboration

Email report subscriptions; **threshold alerts** (KPI/gauge/table) + **ML anomaly detection** (priced per metric evaluated) + built-in ML forecasting; **native iOS/Android apps**; folders (personal/shared/**restricted**, Contributor role); global search, favorites, pinned metrics on home.

## 5. Embedded analytics

- Embeddable: dashboards, individual visuals, **full authoring console**, Q/Generative Q&A bar, Quick agentic experiences (2025+).
- APIs: `GenerateEmbedUrlForRegisteredUser` / `ForAnonymousUser`; JS **embedding SDK** (events/messaging); **1-click embed** incl. public embedding.
- **Tag-based RLS for anonymous embedding** (pass tag values at embed-URL time, OR conditions); user-based RLS for registered.
- Runtime filtering + runtime theming; state persistence + bookmarks in embeds; embedded callback actions; **namespaces** (Enterprise) for multi-tenant isolation; scaled shared folders; brand customization API; footer removal with annual capacity commitment.
- Anonymous embedding requires **session capacity pricing** (below).

## 6. APIs & BI-as-code

- Full `quicksight` API namespace (SDK/CLI): users/groups, namespaces, folders, data sources, datasets (incl. RLS/CLS), ingestions, refresh schedules, templates, themes, topics, VPC connections, dashboards, analyses.
- **Definitions as JSON**: `Describe/Create/UpdateDashboardDefinition` — dashboards fully expressible as declarative JSON (machine-generated, huge).
- **Asset bundle APIs** (BIOps): export/import jobs for cross-account promotion/backup with import-time overrides.
- CloudFormation `AWS::QuickSight::*` (Analysis, Dashboard, DataSet, Theme, Topic, RefreshSchedule…) + Terraform `aws_quicksight_*` (partial coverage); EventBridge asset events for CI/CD. **No native git integration** — BIOps is DIY pipelines.
- **AWS Transform BI migration agents** (May 2026, with Wavicle): convert **Tableau and Power BI** dashboards to Quick Sight.

## 7. Governance & enterprise

- **RLS**: user/group permission datasets (max 999 rules/user; strings only); tag-based RLS for anonymous embeds; inherited by child datasets. **CLS** (Enterprise).
- Identity: IAM, **IAM Identity Center** (GA 2023), AD + SAML 2.0 federation, MFA; custom permission profiles; programmatic group-role assignment.
- Network: VPC connections (Multi-AZ), **PrivateLink** (2024), VPC for MCP (2026). Encryption: SPICE at rest, KMS CMK for SPICE + Quick Research (2026).
- Audit: **CloudTrail**, EventBridge events, per-user cost/usage reporting, user tagging for cost allocation.
- Compliance: SOC, HIPAA, ISO 27001/18/19/9001, GDPR, **FedRAMP**.

## 8. Pricing (Aug 2026)

**Amazon Quick tiers**: Free $0 (no AWS account; chat/research/Flows/apps) · Plus $20/user/mo · **Professional $20/user/mo + $250/account/mo infra fee** (dashboard *viewing*, scenarios, RBAC/SSO, 25 GB index/user) · **Enterprise $40/user/mo + $250 fee** (dashboard/report/dataset *authoring*, Automate authoring, 50 GB/user). Agent hours: Pro 2+2, Ent 4+4 per user/mo; overage $3/agent-hr, $6/research-hr; index overage $5/GB/mo; 30-day trial.

**Quick Sight BI pricing**: Author $24/user/mo · Author Pro $40 (was $50) · **Reader $3/user/mo** · Reader Pro $20 · **$250/account/mo infra fee** when any Pro user or Q&A enabled. **Session capacity** (required for anonymous embedding): 500 sessions/mo = $250 ($0.50 extra); annual tiers 50K/yr $20,000 down to $0.16/session at 1.6M/yr; 30-min sessions; annual unlocks white-label. **Q question capacity**: 500/mo $250; 60K/yr $18,000. SPICE $0.38/GB/mo. Paginated reports $500/mo per 500 units. Anomaly detection $0.05–$0.50 per 1,000 metrics evaluated. Standard Edition ($9–12/user/mo) survives via API only.

## 9. Scale & market position

- **"Over 100,000 customers"**; named: NFL (Next Gen Stats, Combine IQ — hundreds of thousands of concurrent embedded users), Best Western, Siemens, Volvo, Rio Tinto, AutoTrader, MLB, BMW, Docebo, GoDaddy, NBA; Amazon internal deployment "across thousands of users."
- **Gartner MQ ABI**: Challenger 2024 + 2025; reportedly **Leader 2026** (single secondary source — verify). G2 ~4.3/5 (~690 reviews).
- Complaints: limited visual customization vs Tableau/Power BI, admin/setup clunkiness, direct-query performance lags, learning curve, AWS lock-in, SPICE string tax. Strengths: cost ($3 Readers, session capacity), serverless, AWS-native integration, embedded economics, SPICE speed.

## 10. Momentum mid-2025 → mid-2026

Oct 9, 2025: **Quick Suite launch** (Research/Flows/Automate/Index/chat/custom agents; Author Pro $50→$40; purple rebrand; 4 agentic regions) → re:Invent 2025: full Quick Suite session track ("agentic teammate," multi-tenant ISV patterns, embedded revenue streams) → May 2026: **Generate Analysis GA**, **Tableau/Power BI migration agents**, New Relic integration, cross-account Athena → Jun 2026: **autonomous agents**, multi-dataset analytics, Snowflake Cortex MCP, VPC for private MCP, CMK for Research → Jul 2026: **Agentic Catalog Experience** preview (Glue + Unity Catalog) → Aug 2026: **multi-dataset topics** (runtime joins, one semantic model for dashboards + Q&A), 16 new connectors, custom no-data messages. Rename to **Amazon Quick**; desktop + mobile apps (Jun 2026); Quick Autonomous Agents (Summit 2026).
