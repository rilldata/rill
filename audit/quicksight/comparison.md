# QuickSight (Amazon Quick) vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill has it (rough parity or better) · 🟡 partial/different equivalent · ❌ Rill does not have it · ➕ Rill advantage QuickSight lacks. QuickSight detail in `quicksight-features-core.md` / `quicksight-ai-and-market.md`; Rill side in `rill-feature-inventory.md`.

**Framing:** This is a different fight than Cube or Omni. Quick(Sight) is not a better BI tool — it's a *cheaper, AWS-native, procurement-advantaged* one that just pivoted its entire identity to an agentic workspace. Rill rarely loses to Quick on product quality; it loses on "it's already in our AWS bill" (EDP), "$3 readers," AWS-native compliance (HIPAA/FedRAMP), and now "AWS gave us an agentic suite for free-ish." The battleground: embedded analytics for AWS-shop SaaS vendors, operational/event-scale analytics where SPICE breaks down, and AI-BI evaluations.

---

## 1. AI & agentic features

| Feature | How it works in Quick(Sight) | Rill status | Notes |
|---|---|---|---|
| NL Q&A over data | Q/Generative Q&A grounded in **Topics** (curated synonyms/entities/verified answers); multi-visual answers; custom instructions | ✅➕ | Rill's chat is grounded in the *same* metrics layer that powers dashboards — no parallel Topics artifact to curate/drift; citations are live dashboard URLs |
| Semantic grounding architecture | **Two parallel systems**: Topics for structured Q&A + Quick Index/Spaces (RAG) for documents; datasets/analyses/Topics can drift; curation burden heavily documented (re:Post accuracy complaints) | ➕ | Rill: one code-defined semantic layer for humans and agents — the cleanest grounding story in the market. Watch: **multi-dataset topics (Aug 2026)** = AWS converging on this |
| NL dashboard authoring | Build-a-visual from NL; **Generate Analysis GA (May 2026)**: full dashboard from a prompt over ≤3 datasets with editable plan | ✅ | Rill: AI canvas generation, `GenerateMetricsViewFile`, developer agent `init_project` (whole projects incl. ingestion — broader scope than dashboards-only) |
| Executive summaries / narratives | Executive summaries + **data stories** (auto-generated narrative docs/slides) GA 2024 | 🟡 | Rill has AI report narratives + dashboard chat summaries; no "story/slide document" artifact |
| Anomaly & forecasting ML | **ML anomaly detection** (Random Cut Forest, priced per metric), point-and-click + NL **forecasting**, "why" questions | 🟡➕ | Rill's "Explain" (agentic root-cause on metric shifts, interactive, drill links) beats static anomaly narratives; but Rill has **no forecasting engine** at all |
| What-if analysis | **Scenarios** GA Mar 2025 — agentic scenario/what-if workspace | ❌ | Nothing comparable in Rill |
| Deep research agent | **Quick Research**: plans, gathers enterprise + web sources, cited reports | ❌ | Out of BI scope for Rill today; matters in "agentic workspace" evaluations |
| Workflow/process automation | **Quick Flows** (NL automations) + **Quick Automate** (multi-dept processes, UI agents navigating websites, HITL approvals, audit trails) + **Autonomous Agents** (2026: long-running, scheduled/triggered) | ❌🟡 | Rill's analog is narrower: scheduled AI reports + alerts. Rill shouldn't chase RPA; but "dashboard one-click actions → trigger workflow" is a real gap |
| Chat workspace + custom agents | Quick chat with custom agents (personas, scoped data, behavior guidelines), Spaces; Slack/Teams/M365/Chrome extensions | 🟡 | Rill: project/dashboard/developer chat, `@` context; no custom personas, no Slack/Teams surfaces |
| MCP | **Client only**: consumes remote MCP servers (Asana/Jira/Notion/Zapier…, Snowflake Cortex; VPC-MCP Jun 2026). **No MCP server** exposing BI data to external agents found | ✅➕ | Rill is the inverse and better half: a governed **MCP server** (metrics + write tools, free local, full OAuth). Gap: Rill lacks the *client* side (outbound connectors) — same item as Cube roadmap §3.4 |
| Agent-buildable BI | Agents can generate analyses; **dashboards are giant machine JSON** — external coding agents can't meaningfully author Quick assets; no git | ➕➕ | Rill's YAML/SQL-in-git + local MCP + `rill init --agent` makes it the best coding-agent target in BI. Quick's own AI can't be driven by Claude/Cursor |
| AI grounding accuracy ops | Verified answers (pinned, vote-tracked), custom instructions, admin AI on/off per role | 🟡 | Rill: `ai_instructions` + feedback attribution; no verified-answers workflow, no customer-facing evals (shared gap; see Cube roadmap Phase 1) |
| BYO LLM | No — transparent Bedrock model routing, no model choice | ➕ | Rill: BYO key/provider (OpenAI/Azure/Claude/Gemini) from Growth tier |
| Embedded AI | Embed Generative Q&A/summaries/NL authoring (GA Apr 2025); priced $250/mo per 500 questions + $0.50/extra | ✅➕ | Rill embeds full chat with per-end-user history; AI metered in tokens bundled per embedded seat — simpler and cheaper at scale |
| AI pricing | Pro seats ($20–40) + **$250/account/mo infra fee** + question packs + agent-hours ($3–6/hr) + index $5/GB | ➕ | Rill: token quotas per seat, transparent overage. Quick's AI pricing complexity is a real objection (HN flagged it) |

**Net:** Rill wins on grounding architecture, agent-buildability, MCP-server story, embedded AI economics, and BYO LLM. Quick wins on breadth of the agentic *workspace* (research, flows, automation, scenarios, forecasting, stories, chat extensions in Slack/Teams) and on bundling it all into AWS procurement. Rill should not chase the workspace breadth — it should win "trustworthy, fast, embeddable analytics agents" decisively.

## 2. Data plane & performance

| Feature | Quick(Sight) | Rill status | Notes |
|---|---|---|---|
| In-memory acceleration | **SPICE**: per-region capacity purchased at $0.38/GB/mo; 2B rows/2TB per dataset cap; string-heavy data taxed (24B+len); 15-min minimum refresh; incremental refresh only for SQL sources with look-back windows | ➕➕ | Rill: real OLAP engines (DuckDB/ClickHouse/Druid/Pinot/StarRocks) — no per-GB capacity purchasing, no dataset row caps at Rill's ceiling, true incremental/partitioned ingestion, sub-minute-capable refresh, event-scale cardinality where SPICE falls over |
| Direct query | Broad source list; 2-minute visual timeout; complex-query lag is a common complaint | 🟡 | Rill warehouse-direct OLAP mode (BigQuery/Snowflake/Databricks) is secondary; Rill's answer is "land it in OLAP" |
| Transformation | None (dataset joins/calc fields only; heavier prep = Glue/dbt) | ➕➕ | Rill's SQL models, incremental + partitioned models, staging cross-connector pipelines, data tests |
| AWS-native sources | Athena, Redshift, S3(+Tables), Aurora, Glue, OpenSearch, Timestream, IoT… | 🟡 | Rill covers S3/Athena/Redshift/Postgres/MySQL well; no Glue Catalog, OpenSearch, Timestream, S3 Tables/Iceberg-native reads |
| SaaS connectors | Jira, ServiceNow, Adobe, GitHub, Salesforce (BI) + 20 chat connectors | 🟡 | Rill: Salesforce only; no SaaS connector library |
| Streaming | Via Kinesis→S3/Athena patterns; 15-min SPICE floor | ➕ | Rill: ClickHouse-based real-time patterns; still no Kafka connector (shared gap) |
| Real-time refresh economics | Programmatic refresh requires capacity pricing | ➕ | Rill refresh is just model refresh — no premium gate |

## 3. Semantic layer & governance

| Feature | Quick(Sight) | Rill status | Notes |
|---|---|---|---|
| Semantic model | Datasets (joins, calc fields, folders) + separate **Topics** for NL; **multi-dataset topics** (Aug 2026) add runtime joins + one model for dashboards & Q&A | 🟡 | Rill metrics views are the single governed layer, but single-table (no joins/multi-fact — the recurring structural gap); QuickSight datasets *do* join across sources |
| Calculated fields (viewer-side) | Rich function library incl. **LAC-A/LAC-W level-aware calcs**, table calcs, `periodToDate*`, parameters | ❌🟡 | Rill measures (derived/window/comparison) are author-defined; **no viewer-side calc fields** (shared gap with Omni analysis §3.1) |
| RLS | User/group rule datasets (999 rules/user, strings only) + **tag-based RLS for anonymous embeds**; CLS | ✅ | Rill: templated policies on JWT attributes — more expressive (any attribute, SQL row filters, field-level conditions); no anonymous-tag shortcut but attributes cover it |
| Identity | IAM, **IAM Identity Center**, AD, SAML 2.0, MFA | 🟡 | Rill: OIDC/Auth0 (SAML via enterprise connections); no IAM-native story, no SCIM |
| Audit/compliance | CloudTrail, EventBridge, per-user cost reporting; SOC/**HIPAA**/ISO/GDPR/**FedRAMP** | ❌ | Rill: no audit logs; compliance surface far thinner. Blocks regulated/public-sector deals outright |
| Network | VPC connections, PrivateLink, CMK encryption | 🟡 | Rill: VPC peering at Enterprise; no PrivateLink/CMK |

## 4. Visualization & consumption

| Feature | Quick(Sight) | Rill status | Notes |
|---|---|---|---|
| Chart breadth | 30+ types: **maps (3 kinds), radar, Sankey, treemap, waterfall, word cloud, gauge, box plot, histogram, small multiples, Highcharts custom code** | 🟡 | Rill: 11 native types + Vega-Lite escape hatch; missing maps/waterfall/gauge/treemap/etc. (tracked in omni roadmap §2.1) |
| Interactive explore | Dashboards are click-through; analyses are author-only; no slice-and-dice surface for readers | ➕➕ | Rill Explore (leaderboards, TDD, pivot, comparisons, URL state) at sub-second speed is a categorically better consumption experience than a Quick dashboard |
| Layout | Tiled + **free-form** (absolute, overlap, conditional visibility); 20 sheets; cross-sheet filters | 🟡 | Canvas is rows/tabs; no free-form, no conditional visibility, no multi-sheet w/ per-sheet filters |
| Actions | Filter/navigation/URL actions; **one-click actions into workflows** (tickets etc.) | ❌ | No cross-widget actions or click-to-filter in canvas (shared gap) |
| Paginated reports | Pixel-perfect multi-page, scheduled, $500/mo+ | ❌ | Rill: canvas PDF only. Matters in ops/finance deals |
| Alerts | Threshold alerts + ML anomaly alerts (priced per metric) | ✅➕ | Rill alerts: multi-criteria, comparison-aware, **per-user RLS evaluation**, email/Slack — more governed; no anomaly-driven alerting though |
| Scheduled delivery | Email snapshots, report subscriptions | ✅ | Email + Slack, CSV/XLSX/Parquet + AI narratives |
| Mobile | **Native iOS/Android apps** + responsive web | ❌ | None |
| Bookmarks/sharing | Bookmarks, folders (personal/shared/restricted), favorites | ✅ | Bookmarks, public URLs (magic tokens — stronger), custom domains |

## 5. Embedded analytics (the head-to-head market)

| Feature | Quick(Sight) | Rill status | Notes |
|---|---|---|---|
| Embed surfaces | Dashboards, single visuals, **full authoring console**, Q&A bar, agentic experiences | 🟡 | Rill: explore/canvas/component embeds + AI pane; no authoring-console embed ("creator mode" gap, shared with Cube/Omni) |
| Anonymous embedding | 1-click public embed + tag-based RLS + session capacity pricing | 🟡 | Rill public URLs (magic tokens w/ locked filters) cover the read-only case; embeds otherwise want per-user tokens |
| Multi-tenancy | **Namespaces**, scaled shared folders, per-tenant dataset parameters | 🟡 | Rill: org/project isolation + attribute-scoped tokens; no namespace-style tenant asset distribution |
| Runtime control | Runtime filtering/theming, state persistence, callback actions, embedding SDK (JS) | ✅ | postMessage API (state get/set, nav, theme, AI pane); no npm SDK wrapper |
| Embed economics | Sessions: $0.50→$0.16/session at volume ($20k–258k/yr commitments); Q questions $0.30–0.50 each; white-label needs annual commit | ➕ | Rill: flat $500/mo per 100 embedded seats, AI tokens bundled per seat, unlimited at Enterprise. **Simpler and predictable** — but at NFL-scale concurrency, Quick's session pricing is genuinely cheap; Rill wins mid-market SaaS, must sharpen the story at extreme scale |
| Embed performance | SPICE/direct-query latency; concurrency proven (NFL) | ➕ | Sub-second OLAP interactivity is Rill's wedge — an embedded *explore*, not a static dashboard |

## 6. Dev workflow & APIs

| Feature | Quick(Sight) | Rill status | Notes |
|---|---|---|---|
| BI-as-code | JSON definitions via API, asset bundles, CFn/Terraform (partial), EventBridge CI/CD; **no git**; AWS's own "BIOps" is a DIY blog series | ➕➕ | Rill: authored YAML/SQL source in git, branch deployments, cloud edit sessions, local dev binary — the workflow QuickSight cannot offer. This is *the* DX wedge |
| Local development | None (console only) | ➕➕ | `rill start` single binary, full product offline |
| Open source | None | ➕➕ | Entire Rill product OSS |
| Query APIs | SDK/CLI over account+asset ops; no ad-hoc query API over datasets for external tools; no SQL endpoint | ✅ | Rill runtime queries + custom APIs w/ OpenAPI + Metrics SQL + MCP. (Both lack a Postgres-wire SQL API; Rill's is planned — Cube roadmap §3.1) |
| Migration tooling | **AWS Transform agents convert Tableau + Power BI dashboards** (May 2026) | ❌ | Rill has no importer from anything (Cube roadmap proposes Cube/LookML importers; add QuickSight JSON→Rill YAML to the list — their asset-bundle export makes this feasible) |

## 7. Pricing & procurement

| Dimension | Quick(Sight) | Rill | Notes |
|---|---|---|---|
| Viewer cost | **$3/reader/mo** (interactive dashboards) | $20–30/seat | Quick wins raw internal-viewer cost; Rill's counter is seats include explore-level self-service + AI, not passive viewing |
| Author cost | $24–40/author + $250/account AI fee | $20–30/seat, no account fee | Rill simpler/cheaper for author-heavy teams |
| Embedded | Session/question capacity ($20k+/yr commitments for volume + white-label) | Flat $500/mo per 100 seats; Enterprise unlimited | Rill predictable; Quick cheap at extreme concurrency |
| Compute | SPICE $/GB + underlying warehouse (Athena/Redshift) costs | CU-hr usage on managed OLAP; BYO ClickHouse/MotherDuck | Rill's cost includes the engine; Quick's hides warehouse spend elsewhere in the AWS bill |
| Procurement | **In the AWS bill, EDP-eligible, Marketplace** — "already paid for" | Independent vendor | Quick's single biggest advantage; Rill's counter = AWS Marketplace listing (EDP-retiring) |
| Trials/free | Free tier (chat/research, no BI), 30-day trial | OSS product free forever; $250 credit | Rill's free tier is a *complete BI product*; Quick's free tier has no dashboards |

## 8. Where Rill is structurally ahead

1. **One semantic layer for humans and agents** — no Topics/dataset drift, no curation tax; NL accuracy is grounded by construction, with URL citations.
2. **Speed as product**: sub-second explore on event-scale data vs SPICE caps, string taxes, 15-min refresh floors, and 2-minute direct-query timeouts.
3. **Full data plane**: ingestion + incremental/partitioned transformation built in; Quick has none.
4. **BI-as-code + local dev + OSS**: authored source in git vs machine JSON and DIY BIOps; the best coding-agent target vs a closed console.
5. **MCP server with write tools** — external agents can query *and build*; Quick is MCP-client-only.
6. **Embedded product depth**: embedded *explore* + per-user AI chat + predictable pricing vs static dashboards + session/question metering.
7. **Governed AI delivery**: per-recipient-security AI reports, feedback attribution, BYO LLM below enterprise tier.

## 9. Quick's structural advantages

1. **Procurement gravity**: AWS bill, EDP burn-down, Marketplace, IAM-native, "nobody gets fired for buying AWS."
2. **$3 readers + session pricing at extreme concurrency** (NFL-scale) — the cheapest passive-viewer economics in the market.
3. **Compliance breadth**: HIPAA, FedRAMP, ISO set, CloudTrail, PrivateLink, CMK — regulated-industry table stakes Rill lacks.
4. **Agentic workspace breadth**: research/flows/automation/scenarios/forecasting/stories/Slack/Teams/mobile — a surface Rill shouldn't chase but will be demo-compared against.
5. **Chart breadth** (maps, waterfall, Highcharts custom code) and paginated reports.
6. **Migration agents** actively converting Tableau/Power BI estates into Quick.
7. **Ecosystem defaults**: Athena/Redshift/S3/Glue users get connectivity and auth "for free."

## 10. Quick's exploitable weaknesses

1. **Stability/trust**: 4 names in 18 months; Q Business/Kendra/Bedrock-Agents killed July 2026; AWS "application graveyard" reputation — a multi-year BI standardization risk buyers feel.
2. **Curation-heavy AI accuracy** (Topics) with documented complaints; two parallel grounding systems.
3. **Worst as-code story of any major BI tool**; no git; no local dev; closed source.
4. **Pricing opacity** (sessions + questions + agent-hours + index GB + infra fee) — their own forums are full of confusion.
5. **Dashboard-viewer ceiling**: no self-service slice-and-dice for readers; analysts still export to Excel.
6. **Region/language gaps** for AI features; English-optimal.
7. **The pivot vacates classic BI**: accounts that want *great dashboards* rather than an agent workspace are underserved by Quick's direction — exactly Rill's buyer.
