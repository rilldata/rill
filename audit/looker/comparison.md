# Looker vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill has it (parity or better) · 🟡 partial/different · ❌ absent · ➕ Rill advantage Looker lacks. Detail in `looker-features-and-market.md` and `rill-feature-inventory.md`.

**Framing:** Looker is the product Rill's philosophy descends from — git-versioned semantic models feeding governed self-service — now Google-owned, Gartner-Leader-listed, AI-accelerating, and simultaneously bleeding customers to migration (Omni, Lightdash). Rill fights Looker two ways: **as a rival** in code-first/embedded evals (win on speed, DX, price) and **as a destination** for the migration wave (compete with Omni for LookML refugees).

## 1. AI & agentic features

| Feature | Looker | Rill status | Notes |
|---|---|---|---|
| Conversational analytics | Gemini CA GA (Nov 2025), free for all platform users; "How was this calculated?" transparency; Dashboard Agents (Preview '26) | ✅ | Rill chat is dashboard-context-aware with live-URL citations; parity on capability, Rill better on grounding transparency (same layer as dashboards) |
| Grounding | LookML semantic layer; claims errors cut "up to two-thirds"; **verified/golden queries** (Preview 2026) | ✅🟡 | Rill grounds in metrics views by construction; lacks a verified-queries workflow (♻ certified queries, cube roadmap §2.2) |
| Embeddable AI API | **Conversational Analytics API GA (Jun 2026)** + A2A protocol; embedded CA GA | 🟡 | Rill embeds chat in iframes with per-user history; no standalone chat API yet (♻ cube §4.2) |
| Python/advanced analysis | **Code Interpreter/"Advanced Analytics" GA 2026** (NL→executed Python: forecasting, anomaly) | ❌ | Shared gap (♻ quicksight §3.3 agent compute tool) |
| Agentic monitoring | **Agentic Workflows** (Preview: agents monitor metrics, spot anomalies, notify, recommend) | 🟡 | Rill alerts + AI reports + Explain cover parts; no NL-created monitoring agents |
| LookML/code assist | LookML Assistant, Expression Assistant, VS Code extension with AI + MCP (names Claude Code!) | ✅➕ | Rill's developer agent *writes whole projects*; `rill init --agent` + write-capable MCP is stronger than assist-in-IDE |
| MCP | Managed MCP server **Preview May 2026** (read/query; admin allowlists; audit-logged) + MCP Toolbox OSS | ✅➕ | Rill MCP is GA-equivalent, includes **write tools**, free local. Looker's is read-oriented and preview |
| Slides/report generation | Automatic Slide Generation (live-linked Google Slides) | ❌ | Google-ecosystem strength; low priority for Rill |
| AI observability | CA observability (token usage, success rates, feedback in System Activity) | 🟡 | ♻ AI admin console (cube §2.6) |
| AI pricing | Data-token pools per instance; **overage billing starts Oct 1, 2026** ($3/1M in, $20/1M out) | ➕ | Rill: transparent per-seat token quotas + BYO LLM from Growth. Looker has no BYO LLM (Gemini only) |

## 2. Semantic layer

| Feature | Looker | Rill status | Notes |
|---|---|---|---|
| Joins + symmetric aggregates | LookML explores' signature capability | ❌ | ♻ The recurring #1 gap (composite metrics views planned) |
| PDTs / aggregate awareness | Mature: triggers, incremental PDTs, `aggregate_table` routing | 🟡 | Rill rollups early; but Rill's OLAP engine makes most of it unnecessary |
| Extends/refinements | Model layering/inheritance | 🟡 | Derived metrics views only (subset, no overrides) |
| Access grants + user attributes | Named grants; attributes drive RLS + Liquid | ✅🟡 | Rill security policies = parity on RLS/field access; no named reusable grants |
| Liquid templating | Parameterized SQL/filters/links | 🟡 | Go templating covers env/user-attrs; no dashboard-filter→SQL binding |
| Git workflow | Browser IDE, branches, PRs, **Looker CI GA**, VS Code extension | ✅➕ | Rill: git *is* the source, local binary, reconciler validation — simpler and stronger; but Looker CI's SQL/content testing is ahead of Rill's CI story |
| Semantic layer for other tools | Connected Sheets, Studio connector, Open SQL Interface (JDBC); Modeler shelved | ❌ | ♻ Postgres-wire Metrics SQL (cube §3.1) answers this |
| dbt | No first-party integration (ecosystem tools) | ❌ | Even — both lack it; ♻ cube §3.5 dbt read would leapfrog |
| Data plane | None (live-query only; PDTs materialize in customer warehouse) | ➕➕ | Rill's ingestion + incremental models + managed OLAP |

## 3. BI product & consumption

| Feature | Looker | Rill status | Notes |
|---|---|---|---|
| Governed self-service explore | Explores: field picker over join trees, pivots, table calcs, custom fields, merged results | 🟡 | Rill Explore is faster but modeled-fields-only; no viewer calcs (♻), no cross-model exploration |
| Dashboard speed | Warehouse-bound; ~38% of G2 reviewers flag slowness; guardrails not engine fixes | ➕➕ | Rill's core wedge — sub-second by architecture |
| Deliveries | Email/Slack/S3/SFTP/webhooks via Action Hub; per-tile alerts + Gemini key-driver insights | 🟡 | Rill: email/Slack only (♻ destinations, omni §2.4); alerts stronger on per-user RLS |
| Paginated reports | GA Next '26 | ❌ | ♻ report-layout canvas (quicksight §2.5) |
| Tabbed dashboards, boards, marketplace viz | Mature content ecosystem | 🟡 | Canvas tabs exist; no marketplace/custom-viz ecosystem |
| Self-service CSV/Excel blend | **Self-service Explores GA '26** (personal uploads + governed data, auto-LookML) | ❌ | Notable new Looker capability; Rill uploads are developer-flow |
| Mobile | Thin apps | ❌ | Even, effectively |

## 4. Embedded

| Feature | Looker | Rill status | Notes |
|---|---|---|---|
| Embed mechanics | Signed/SSO embed, user attributes → RLS, Embed SDK, cookieless, extension framework | ✅ | Rill iframe + postMessage + attributes = parity on mechanics; no SDK wrapper (♻) |
| Embedded AI | Embedded CA GA + CA API | 🟡 | Rill: embedded chat with per-user history; no API |
| White-label | Private labeling (Embed edition) | ✅ | Themes + custom domains |
| Economics | **~$180–198K/yr list** platform fee + ~$30/viewer/mo; "most embedded buyers skip Looker on price" | ➕➕ | Rill $500/mo per 100 embedded seats. The single easiest competitive displacement in Rill's market |

## 5. Enterprise & pricing

| Dimension | Looker | Rill | Notes |
|---|---|---|---|
| Governance/audit | System Activity + Cloud Audit Logs, CMEK, VPC-SC, model sets/roles | ❌🟡 | Rill lacks audit logs/SCIM (♻); Looker enterprise-mature |
| Procurement | Billed through GCP → burns GCP committed spend; BigQuery bundling | ➕/❌ | Same dynamic as AWS/Quick: platform gravity. Rill counter: GCP Marketplace listing (mirror of quicksight roadmap §2.1) |
| Pricing | ~$60K–200K/yr platform + per-seat ($125 dev, ~$30–66 user/mo); call-sales opacity | ➕ | Rill: $20–30/seat transparent, OSS free tier. Realistic Looker entry $50–100K+/yr |
| Deployment | GCP-hosted (core); customer-hosted legacy only | ✅ | Rill local + cloud; OSS self-host |

## 6. Where Rill wins / loses

**Rill structural advantages:** sub-second OLAP vs the slow-dashboard reputation; full data plane; radically simpler code-first DX (local binary, YAML, git-as-source); write-capable MCP + agent-buildable projects; embedded economics (~100x cheaper entry); transparent pricing + OSS; BYO LLM.

**Looker structural advantages:** joins/symmetric aggregates + 10 years of semantic-layer maturity; Gartner Leader ×3 + Google distribution (GCP spend, BigQuery bundling, Gemini brand); enterprise governance depth (audit, CMEK, VPC-SC); Action Hub/marketplace ecosystem; CA API + A2A embedding story; paginated reports; the "free CA for all users" giveaway (until Oct 2026 overage billing).

**Looker's exploitable weaknesses:** slow dashboards (38% of reviewers) with no engine answer; migration wave already in motion (Omni/Lightdash prove demand — Rill should be on the shortlist); LookML learning curve vs Rill YAML; price (platform fees + seats + AI overage starting Oct 2026); post-acquisition support decay; strategy churn (Modeler shelved, Studio-in-Looker killed, Studio renamed twice); customer-hosted abandoned; Gemini-only AI (no BYO LLM).
