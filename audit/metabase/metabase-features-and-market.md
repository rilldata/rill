# Metabase — Platform, AI & Market Inventory

> Researched Aug 2026 from metabase.com (pricing, releases v55–v63, paid-features, embedded, AI pages — fetched directly), docs, GitHub, press. Metabase, Inc.: founded 2014 out of Expa, launched 2015; Sameer Al-Sakran CEO; ~120–130 employees; ~$51M raised (Series B $30M, Insight, Aug 2021 — nothing since); claims profitable-ish PLG, "8-figure ARR," site says "trusted by 100,000+ companies"; ~49k GitHub stars.

## 1. Core product

- **Question builder** (GUI notebook editor): click-based filtering/summarizing/joins/custom expressions, drill-through everywhere (signature feature), command palette.
- **Native SQL editor**: variables + **field filters** (dashboard widgets against native SQL), snippets (parameterizable v57), reference saved questions (`{{#123}}`), AI generation/debugging.
- **Models** (curated datasets, v42 2022; persistence/caching into the warehouse) and **Metrics** (relaunched v51 Nov 2024; metrics explorer v60, metrics math v61).
- **Data Studio** (v59, Mar 2026) — the semantic-layer workbench: Library of published tables/metrics/snippets, measures, segments, **glossary** (business definitions used by AI), dependency graph + diagnostics, interactive ERD (v62), **transforms (SQL + Python)**, git sync of the semantic library. Metabase now markets a "Semantic layer" nav item.
- **~20+ DB drivers** (Athena, BigQuery, ClickHouse, Databricks, Druid, MongoDB, MySQL, Oracle, Postgres, Redshift, Snowflake, SQL Server, Trino/Starburst, Vertica…); CSV upload, Google Sheets sync, Cloud Storage ($2/1M rows).
- **Caching**: per-DB/dashboard/question policies (granular = paid), preemptive cache warming (v53), model persistence. No engine of its own — live queries to the source.
- **Dashboards**: filters (linked, custom sources, per-card), click behavior/cross-filtering, tabs, iframe/link cards with variables, **Actions** (DB writeback, v46), data visualizer (v55), subscriptions (email/Slack, PDF attachments v63), **Documents** (v57 — notebook-style narrative docs, all plans).
- **Viz**: standard set + waterfall, Sankey (v52), box plot (v59), treemap (v63), pie/sunburst, maps; **custom visualizations plugin SDK** (v62 — React chart types via `@metabase/custom-viz`, works in embeds; paid). X-rays (auto-insight dashboards, 2017 vintage). Alerts (goal/results, cron-level v54, webhooks).

## 2. Open source vs paid

- **Core AGPL**, unlimited users, self-hosted; paid features license-token-gated in the same repo (Metabase Commercial License). Pro and Enterprise are **feature-identical** ("Feature-wise, zero" difference).
- OSS includes: full builder/SQL/viz/dashboards/models/metrics/alerts, public sharing + basic static embedding, and (2026) **open-sourced AI** — AI SQL generation + Metabot exploration with BYO key.
- **Paid-only**: **sandboxing (RLS/CLS)**, impersonation, database routing (v55), **Tenants** (v58), SSO (SAML/JWT/OIDC/LDAP) + **SCIM**, audit/usage analytics, whitelabeling (logo/fonts/domain/UI), modular embedding SDK/JS + embedded AI, **serialization** (YAML export/import) + **Remote Sync** (v57 git integration), Data Studio, official collections/verification, granular caching, AI governance (per-group access, token limits, system prompts, usage analytics), 2FA, air-gap (Enterprise +fee).

## 3. Pricing (fetched 2026-08-29)

| Plan | Price | Users | Extra users |
|---|---|---|---|
| Open Source | Free | Unlimited | — |
| Starter | $100/mo ($90 annual) | 5 | +$6/user/mo |
| **Pro** | **$575/mo** ($517.50 annual ≈ $6.9K/yr) | 10 | **+$12/user/mo** |
| Enterprise | From **$20,000/yr** | custom | custom (incl. embedding packages) |

- **Billing user definition**: anyone who signs in — **including authenticated embed end-users** (+$12/user/mo on Pro); guest/static embeds don't count (unlimited viewers).
- Add-ons: AI service $3.75/1M tokens (first 1M free; BYO key = free), transforms $0.01–0.02/run past included, storage $2/1M rows, dev instances $100/mo. No charges for query volume/dashboards/sources. SOC 2 Type II.

## 4. Embedded analytics (the strategic surface)

Modes: **static/signed embeds** (OSS; translation dictionaries v56) → **Guest Embeds** (v58, view-only, no auth, auto token renewal) → **full-app embedding** (paid, SSO, drill-through, whitelabel) → **Embedded Analytics SDK for React** (GA Feb 28, 2025; now "Modular embedding SDK": component-level — dashboards, questions, query builder, **AI chat/Metabot component** — theming, plugins) → **Modular embedding JS** (v56, framework-agnostic library on the SDK). Sandboxing/impersonation/DB routing/**Tenants** for multi-tenant RLS; per-tenant AI token limits. Agent skills teaching LLMs to implement Metabase embedding. Named embedded customers: CourseKey, Marketfuel, Rechat, Sincera, Faros AI, Cal.com, Lime.

**Embed economics**: no separate SKU — Pro $575/mo floor; authenticated embed users billed per seat; guest embeds free-viewer; Enterprise custom packages when per-seat doesn't fit.

## 5. AI features (Metabot arc)

- Prehistory: Slack Metabot (retired), GPT "Metabot 2.0" 2023 (deprecated).
- **v58 (Jan 2026)**: Metabot **out of beta**, add-on on all Cloud plans — NL questions, chart/SQL generation, query fixing, summaries.
- **v59 (Mar 2026)**: AI SQL generation in **Open Source**; Data Studio as grounding; AI transform codegen.
- **v60 (Apr 2026)**: **"AI just went open source"** — AI tools open-sourced; **official MCP server** (Claude/ChatGPT/Cursor/VS Code); self-hosted Metabot with **BYO model**; **Metabot in Slack**; "Metabase is becoming a platform AI agents can read and write to."
- **v61 (May 2026)**: **AI governance** — per-group access, token/message limits (user/group/tenant), custom system prompts + branding, usage analytics; **dashboards-as-code from AI coding terminals**.
- **v62 (Jun 2026)**: MCP returns **interactive charts inline**; **Metabase CLI** "built for coding agents."
- **v63 (Jul 2026)**: providers expand — OpenAI, Bedrock, Azure join Anthropic.
- Roadmap: "Explorations" AI surface, **Workspaces** (per-agent isolated warehouse spaces), data apps on the semantic layer, catalog with lineage/PII.
- Grounding story: "every AI answer is inspectable" (links to the query), respects permissions, grounded in Data Studio (models/metrics/**glossary**), surfaces existing dashboards before generating.
- AI pricing: BYO key free, or $3.75/1M tokens (20% markup, 1M free).

## 6. BI-as-code

Serialization (YAML export/import, paid) + published Representation Format spec + validation package; **Remote Sync** (v57): push/pull content to git from the UI, each dashboard/question a reviewable YAML file; dashboards-as-code via MCP/CLI (v61–62); config-file init; REST API + API keys; Agent/Workspace API; community (not official) Terraform providers; dev instances + dependency checks.

## 7. Market & competitive posture

- G2 ~4.4/5 but only ~147 reviews; praised for setup speed, non-tech usability, price. Complaints: performance on large data (live queries, no acceleration), historically thin modeling (improving via Data Studio), permissions complexity at scale, **feature gating** (SSO/sandboxing/serialization paywalled; $100→$575 Starter→Pro cliff), content sprawl.
- Typical buyer: startups/SMB internal BI (PLG) + embedded SaaS; loses enterprise-primary-BI to Looker/Power BI/Tableau.
- vs Superset/Preset: wins time-to-value/UX, loses viz breadth/scale. vs Redash: won (maintenance mode). vs Lightdash: dbt-native segment vs Metabase's own as-code semantic story. In embedded bake-offs: the default cheap/fast option; weaknesses = query performance at concurrency, per-authenticated-user pricing at consumer scale, iframe modes less native (SDK addresses).
- **Momentum v55→v63 (monthly majors)**: data visualizer/DB routing (v55) → embedded JS (v56) → dark mode, Remote Sync, Documents (v57) → Tenants, Guest Embeds, Metabot GA (v58) → **Data Studio** (v59) → **AI open-sourced + MCP + Slack + BYO model** (v60) → AI governance + dashboards-as-code (v61) → custom viz SDK + CLI (v62) → treemap/2FA/PDF/providers (v63). Clear repositioning: "easy dashboards" → "AI analytics platform grounded in a semantic layer."
