# Sigma vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill parity or better · 🟡 partial/different · ❌ absent · ➕ Rill advantage. Detail in `sigma-features-and-market.md` and `rill-feature-inventory.md`.

**Framing:** Sigma is the momentum machine — $200M ARR doubling yearly, $3B valuation, first sold-out user conference, and a 2026 pivot from "spreadsheet BI" to "the AI runtime for business" (data apps, writeback, agents). Its identity is *self-service depth on the live warehouse*; its Achilles' heel is the same architecture: **every click is a warehouse query and a warehouse bill**. Rill meets Sigma in modern-BI evals and embedded deals; Rill's engine + code-first governance vs Sigma's spreadsheet UX + app platform.

## 1. AI & agentic features

| Feature | Sigma | Rill status | Notes |
|---|---|---|---|
| NL analyst | **Sigma Assistant/Ask Sigma**: editable multi-step plan ("Analysis breakdown"), explore-into-workbook, thumbs→admin dashboard | ✅ | Rill parity+ (dashboard context, citations, persistence/sharing). Sigma's *editable plan* transparency UX is best-in-class — worth emulating |
| Grounding | Endorsed data models/metrics/relationships + "AI context" field; curated or auto-indexed sources; RLS/CLS respected | ✅ | Same philosophy as Rill (`ai_instructions` ≈ AI context; endorsed ≈ certified ♻); Sigma executes the curation UX better |
| Agents | **Sigma Agents** (beta, premium SKU): instructions + sources + tools (deterministic actions, MCP servers, warehouse agents); chat or **scheduled/autonomous**; "fastest-adopted product ever" | 🟡 | Rill has the runtime (router/analyst/developer) but not user-defined agents with tools (♻ skills + outbound MCP + scheduled tasks packaging, cube Phase 1/2 — Sigma just productized that exact bundle) |
| Warehouse agents | First-class Cortex Agents + Genie Spaces invocation | ❌ | Interesting partner-not-fight move; Rill analog = federating to warehouse AI where warehouse-direct mode is used (low priority) |
| AI in formulas/columns | AI Query GA (warehouse LLM functions as columns), formula assistant, input-table AI fill | ❌🟡 | Rill `ai` resolver is an analog for pipelines, not for viewer columns |
| NL app/dashboard building | AI Builder (private beta), Assistant build mode | ✅ | Rill AI canvas/project generation comparable-or-ahead (ingestion included) |
| MCP server | Live; Claude Connectors Directory, ChatGPT plugin, Codex, Cursor; user-permission-inherited; warehouse-tagged queries | ✅➕ | Parity; Rill adds write tools + OSS/local. Sigma's cost-attribution tagging is a nice touch |
| MCP client | Announced, coming | ❌ | ♻ outbound MCP — both racing |
| Proactive insights | Ask Sigma Discover (announced) | 🟡 | ♻ insight scans (ts roadmap §2) |
| Accuracy evidence | None published (transparency + endorsement narrative) | ➕ potential | Neither has a benchmark — first mover wins the claim (♻ evals) |
| AI pricing | Assistant tier-gated; **agents premium SKU**; LLM + warehouse-AI costs pass through to customer accounts | ➕ | Rill: AI tokens bundled per seat, managed or BYO. Sigma's pass-through means AI usage inflates the *warehouse* bill too |

## 2. Self-service & apps (Sigma's identity)

| Feature | Sigma | Rill status | Notes |
|---|---|---|---|
| Spreadsheet calcs for viewers | Excel-like formulas anywhere, pivot ad-hoc calcs | ❌ | ♻ ad-hoc measures (omni §3.1) — the shared #1 self-service gap |
| Input tables / writeback | Deep: validation, audit, linked tables, file columns; **Sigma Tables** (shared CRUD, app substrate) | ❌ | Rill has nothing; Sigma is furthest ahead of anyone here |
| Data apps / actions | Actions framework (sequences, if/else, API/stored-proc calls, agent calls), modals/forms/containers | ❌ | Rill canvas variables are a thin analog; ♻ dashboard actions (qs §3.4) covers a slice |
| Explorations from any element | Ad-hoc exploration + custom views + saved views | ✅ | Rill Explore + bookmarks comparable for modeled data |
| Paginated reports | Reports doc type (pixel-perfect PDF) | ❌ | ♻ report-layout canvas (qs §2.5) |
| Python | Notebook elements (Snowflake/Databricks only) | ❌ | ♻ agent Python tool |
| Viz breadth | Gauge, funnel, Sankey, waterfall, box, maps native | 🟡 | ♻ chart gaps again |

## 3. Architecture & platform

| Feature | Sigma | Rill status | Notes |
|---|---|---|---|
| Engine/cost model | **Live warehouse queries for everything**; materialization as mitigation; complaint #1 = warehouse bill unpredictability; warehouse-only sources; no on-prem/OSS | ➕➕ | Rill's managed OLAP = predictable cost + sub-second regardless of warehouse tuning; BYO ClickHouse for zero-copy. The sharpest structural contrast |
| Data plane | None (upstream ETL assumed) | ➕➕ | Ingestion + incremental models |
| Semantic layer | Data models: relationships (join-on-demand!), metrics (limited: no window fns, per-source), RLS/CLS, as-code YAML via API | 🟡 | Sigma has joins (♻ Rill gap); Rill measures are deeper (window/comparison/derived); Rill as-code is native not export-shaped |
| Governance risk | Spreadsheet freedom → "metric divergence within six months" without discipline | ➕ | Rill's governed-by-construction story is the direct counter |
| Version control | Tags + approval flows; API-based as-code; no git | ➕ | Rill git-native + local dev + OSS |
| dbt / warehouse semantics | dbt SL query integration; Snowflake Semantic Views | ❌ | ♻ dbt read (cube §3.5) |
| Multi-tenancy | **Sigma Tenants** (isolated workspaces, API provisioning, delegated admin) | 🟡 | Rill org/project + attribute tokens; no tenant-workspace productization |
| Enterprise IT | SAML + **SCIM**, audit-log connection + export, OAuth warehouse passthrough | 🟡❌ | ♻ SCIM/audit gaps again |
| Localization | UI + content translation files | ❌ | ♻ i18n (omni §4.4) |

## 4. Embedded

| Feature | Sigma | Rill status | Notes |
|---|---|---|---|
| Mechanics | JWT one-time URLs, JS Embed API, **React SDK (OSS)**, custom domains, theming/localization | 🟡 | Rill lacks SDK (♻); otherwise parity |
| Embedded self-service | Explorations, saved views, **embedded writeback**, **embedded Ask Sigma** | 🟡 | Rill: embedded explore + AI chat with per-user history; no writeback/creator surfaces |
| Multi-tenant | Tenants + user attributes + dynamic warehouse roles | ✅ | Attribute-based RLS parity; Tenants packaging is ahead |
| Economics | Negotiated book; underlying warehouse compute per embed click passes to the customer | ➕ | Rill flat $500/100 seats with engine included — at embed concurrency, Sigma's warehouse passthrough is the hidden cost bomb |
| Named | Persona, G2, DoorDash, Affirm, Blackstone | — | Sigma's embed logos are strong |

## 5. Where each wins

**Rill structural advantages:** engine economics (fixed-cost sub-second vs per-click warehouse bills — Sigma's #1 complaint); full data plane; governed-by-construction metrics (vs metric-divergence risk); git-native BI-as-code + local dev + OSS; embedded cost predictability; AI tokens bundled vs pass-through; write-capable MCP.

**Sigma structural advantages:** self-service depth (spreadsheet calcs, writeback, input tables, data apps) — the market's best; momentum + capital ($200M ARR, $3B, strategic investors incl. Databricks/ServiceNow/Workday); Sigma Agents productization (tools + scheduling shipped); warehouse-native trust (data never leaves); Tenants; localization; enterprise IT (SCIM/audit); embed logos.

**Sigma's exploitable weaknesses:** warehouse-compute passthrough (cost + performance tied to the customer's warehouse tuning — worst at embed/AI scale); warehouse-only sources, no on-prem/OSS/local; governance optionality; opaque sales-led pricing (median $68K/yr); metrics limitations (no window functions, per-source scoping); no published AI accuracy evidence; viz depth still trails Tableau.
