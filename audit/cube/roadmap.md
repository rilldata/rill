# Roadmap: Competing with (and beating) Cube

Companion to `comparison.md`. Effort sizes: S (≤1 eng-month), M (1–3), L (3–6), XL (6+ across teams). Items shared with the Omni roadmap (`../omni/roadmap.md`) are marked ♻ — build once, win both fights.

## 1. Strategic frame

### Where the fight actually is

Cube spent 2025–26 converting semantic-layer credibility into "agentic analytics" — and it now owns that category's narrative (coined the term, got the Gartner Market Guide listing, ships an AI/agent feature every ~2 weeks). But its position has three structural cracks Rill is built to exploit:

1. **Cube has no data plane.** No ingestion, no transformation, no storage. Every un-pre-aggregated query hits the customer's warehouse; its performance answer is pre-aggregations — the single most complained-about surface in Cube's own community (build/refresh complexity, multi-tenant cache headaches, "slow AI analyst" on G2). Rill's managed OLAP makes speed the default, not an engineering project.
2. **Cube's AI platform is 100% Cloud-gated and seat-priced.** MCP requires Premium ($80/dev/mo); BYOM, SSO, DAX are Enterprise; the OSS project explicitly gets none of the agentic frontend (and HN already notices Core "falling behind"). Rill's OSS product is the whole product — local AI chat, write-capable MCP, dashboards — free on a laptop.
3. **Cube is squeezed from above.** Snowflake Semantic Views, Databricks Metric Views, dbt+Fivetran/MetricFlow-OSS, and Looker's Conversational Analytics API all commoditize "semantic layer middleware." Cube must defend the middleware position while pivoting to BI. Rill never depended on that position: its wedge is the full stack (data → OLAP → BI → agents).

Cube's real strengths to respect: Semantic SQL over Postgres wire (universal interop), joins/multi-fact modeling, the AI operationalization suite (skills/rules/certified queries/evals/AI Studio), embedded breadth (Chat API, Creator Mode), and install-base/analyst mindshare.

**Positioning: "The agentic BI platform that's fast by default and open all the way down."** Cube grounds agents in a semantic layer over your slow warehouse; Rill grounds agents in a semantic layer over a sub-second OLAP engine it also manages — so both humans *and agents* get instant answers, and agents can *build*, not just ask. Win operational/embedded/event-scale deals outright; achieve sufficient AI-ops and interop parity so Rill stops losing "agentic analytics" evaluations on checklists.

### The four deal-loss categories to eliminate
1. **"Cube plugs into my existing BI stack; Rill doesn't."** (SQL API, Power BI/Excel, BI sync)
2. **"Cube's agents are tunable, testable, and everywhere; Rill's chat is a black box."** (skills, rules, evals, Slack, outbound tools)
3. **"Rill can't model my schema."** (joins/multi-fact — same as Omni loss #1)
4. **"It's missing enterprise table stakes."** (SCIM, audit logs, BYOC, custom roles)

### Deliberate non-goals (§7)
Warehouse-agnostic middleware play, 20-engine live-query breadth, DAX/MDX full fidelity, workbook/Python document paradigm, Cube Store-style pre-agg infrastructure.

---

## 2. Phase 1 — AI operationalization parity (0–3 months) — THE PRIORITY

Cube's biggest genuine lead is that it productized the agent trust/tuning loop. Rill's `runtime/ai/` architecture already contains most of the raw material; this phase is packaging, not research. This is also the phase that feeds the "agentic analytics" evaluation checklists Gartner just legitimized.

### 2.1 Agent Skills (S–M) ♻
**Cube:** markdown skills in `agents/skills/` in the repo — chat buttons, slash menu, auto-matching; versioned/reviewed like code; plus an open-source Claude Code skills repo.
**Plan:** Rill is *better* positioned here — projects are already git repos parsed by `runtime/parser`:
- New resource kind `skill` (`skills/*.md` or YAML frontmatter + markdown body): name, description, inputs, instructions. Parser + reconciler are trivial; skills surface as (a) one-click buttons/slash commands in all three chat surfaces (`web-common/src/features/chat/`), (b) **MCP prompts** (the MCP spec has prompts; almost nobody ships them — instant differentiation for Claude/Cursor users), (c) invocable by the router agent via auto-matching.
- Ship a starter skill library in `rill init` templates + docs gallery.
- **Mirror Cube's Aug 2026 move**: publish `rilldata/rill-agent-skills` for Claude Code (explore-metrics, build-model, build-dashboard, deploy, configure-embed…) installable via the plugin marketplace. Rill already ships `rill init --agent claude` (`cli/cmd/initialize/init.go`) — extract + publicize.

### 2.2 Rules, certified queries, and memories (M)
**Cube:** Rules (always/agent-requested markdown), Certified Queries (admin-approved SQL agents reuse as trusted examples), Memories (space/user isolation), multi-agent config in `agents/config.yml`.
**Plan:** generalize `ai_instructions` into a first-class AI config block (project-level `ai:` in `rill.yaml` + per-metrics-view):
- `rules:` list (markdown files or inline; always vs on-request) injected into agent prompts (`runtime/ai/instructions/` already does templated prompt assembly).
- `certified_queries:` named metrics queries / Metrics SQL with NL descriptions — served to the analyst agent as few-shot examples and exposed via MCP as example resources; doubles as the eval ground truth (§2.3).
- `memory:` opt-in conversation-derived memory per user/project (catalog store already persists sessions; add a distillation step in the completion loop). Scope: user-level only at first; skip Cube's space isolation until multi-agent exists.
- BYO model config already exists (`ai_connector`) — advertise it against Cube's Enterprise-gated BYOM.

### 2.3 Customer-facing evals, CI-gateable (M) ♻ — highest leverage/effort ratio
**Cube:** `agents/eval_questions.yml` (NL question + ground-truth SQL), sort-invariant numeric-tolerant result comparison, per-branch runs, AI Studio UI.
**Plan:** productize `runtime/ai/evals/`:
- `evals:` resource: prompt + expected result (certified query ref, Metrics SQL, or golden values). Runner executes the real analyst agent against any branch deployment (branch-scoped deployments + MCP already exist) and compares result sets (reuse Cube's tolerant-comparison semantics — they're right).
- Surfaces: `rill ai eval` CLI (CI-gateable — beat Cube to "evals in GitHub Actions" as the default workflow, it fits Rill's git-first story better than their IDE), plus a simple results page in web-admin.
- Marketing: "golden completions, not vibes" is already Rill's blog line — make it a product screenshot and publish accuracy numbers (buyers now cite runQL/Delphi benchmarks; absence of numbers reads as weakness).

### 2.4 Slack agent (M) ♻
**Cube:** DM the Cube app; threads = chat sessions; email-matched identity; user's own RLS applies. Premium+.
**Plan:** extend the `slack` driver beyond notifier: Slack app with DM + @mention → router/analyst agent, thread = conversation (persistence exists), email→user mapping → JWT attributes so **security policies apply per asker**, chart images in replies (needs PNG render — §4.3), link back to explore with filters applied. This is the highest-visibility AI demo feature in every 2026 eval; Rill's per-user security story makes its version *more* enterprise-credible than Cube's.

### 2.5 Scheduled agent tasks, loudly (S)
**Cube:** Scheduled Tasks (NL prompt + cadence, can build dashboards + email members) — shipped Jul 2026 as a headline feature.
**Rill already has this** (`resolver: ai` reports with per-recipient security — ahead of Cube on governance). Gaps to close: let the scheduled agent optionally use developer tools on a branch (create/update a canvas, open a PR — never direct-to-prod), add Slack delivery of narrative + chart, and **name it** ("AI Reports" → productized "Routines/Tasks" page in web-admin). Mostly packaging + docs + a launch post.

### 2.6 AI admin & usage surface (M)
**Cube:** AI Studio, Chats History (admin), usage analytics incl. AI, per-seat token grants.
**Plan:** web-admin org/project **AI page**: conversations volume, thumbs up/down rates, **feedback-attribution breakdown** (the `feedback_agent` already classifies failures as Rill's/project's/user's fault — no competitor has this; make it a dashboard), token spend per user/embed org, per-org and per-embed AI caps (billing already meters tokens). Admin conversation browser gated by an org setting (privacy-flagged).

**Phase 1 exit criteria:** In an "agentic analytics" bake-off, Rill demonstrates skills, rules, certified queries, evals-in-CI, Slack agent, scheduled AI tasks, and an AI admin console — while being the only vendor whose agent also *builds* the project and runs free locally.

---

## 3. Phase 2 — Interop & semantic depth (3–6 months)

### 3.1 Postgres-wire SQL API over metrics views (L) — the big interop bet
**Cube:** Semantic SQL over Postgres protocol is its most-used API and now "the primary API going forward"; it is how Tableau/Metabase/Superset/psql/pandas/every LLM tool connect.
**Plan:** a `pgwire` endpoint on the runtime speaking Postgres protocol, translating to **Metrics SQL** (parser exists: `runtime/metricsview/metricssql/`) with metrics views exposed as tables and measures as `MEASURE()`-style columns (adopt Cube's convention — it's becoming a de-facto standard and eases migration *from* Cube). Read-only, JWT/service-token auth mapped to existing security claims so RLS holds. Scope v1 to the aggregation-query shape Metrics SQL supports; punt full pushdown.
Payoff: every "can my analysts use their existing tools" objection dies; Rill becomes usable as a *semantic layer* by teams not ready to switch BI — Cube's own wedge, reversed. Also instantly makes Rill consumable from any tool that speaks Postgres, including AI notebooks and LLM data tools.

### 3.2 Semantic joins: shared dimensions → composite metrics views (XL) ♻
Identical to the Omni roadmap §4.1 (shared/conformed dimensions → composite metrics views stitched by FULL OUTER JOIN post-aggregation → many-to-one dimension-table joins later). Against Cube specifically, the pitch is stronger: Cube joins tables and needs Tesseract machinery to avoid fan-out; Rill joins *after aggregation*, correct by construction. Required to stop losing enterprise-schema deals; also unlocks §3.1's usefulness on real schemas. Includes fiscal/custom calendars (Cube's calendar cubes) as part of the time-model work (S once designed) ♻.

### 3.3 Excel + Google Sheets, the pragmatic slice (M)
**Cube:** MDX API + native add-ins (Cloud), DAX API (Enterprise).
**Plan:** do NOT build MDX/DAX. Ship (a) **Google Sheets add-on** ("connected metrics": pick metrics view/measures/dims/time range → live-refreshing range; reuse existing REST APIs + OAuth), and (b) **Excel via the Postgres SQL API** (§3.1) with a documented ODBC path + templates. Covers ~80% of "get governed numbers into a spreadsheet" at ~20% of the cost. Power BI stays a documented SQL-API connection, not a DAX endpoint (revisit only on repeated enterprise loss evidence).

### 3.4 Outbound MCP connectors for Rill agents (M)
**Cube:** agent consumes Notion/Linear/Sentry/Attio + custom MCP tools for context and write-back.
**Plan:** `runtime/ai` already speaks MCP natively (go-sdk) — add client-side support: project-level `ai.connectors:` (remote MCP endpoints + allowed tools + per-user OAuth where needed), tools surfaced to analyst/developer agents behind admin enablement, nothing on by default. The approved external-MCP-connector design on branch `mcp_external` (see memory) is the starting point. Differentiator to market: Rill agents can pull a Linear ticket for context *and* open a PR on your metrics repo — a full loop Cube's model-frozen scheduled tasks can't do.

### 3.5 dbt read-integration (S–M)
**Cube:** one-way dbt→Cube sync with CI/webhooks, `cube_dbt`.
**Plan:** read dbt `manifest.json`/`semantic_manifest.json`: import models as external tables/sources with lineage + docs; optional MetricFlow-metric → metrics-view conversion (MetricFlow is now Apache 2.0 — legally clean). Positions Rill as complementary to the dbt-standard stack instead of ignoring it.

### 3.6 Standards & catalog visibility (S, non-eng-heavy)
- **Join OSI / Apache Ossie.** Cube, Omni, Hex, Sigma, ThoughtSpot are members; Rill's absence reads as irrelevance in every analyst deck. Implement OSI import/export for metrics views when the v0.x spec stabilizes; being the "OSI + open source + fast engine" vendor is cheap, on-brand positioning.
- **Model organization metadata** (S): `folder:`/`group:` + synonyms on metrics-view fields ♻ — feeds search, MCP tool descriptions, and any future catalog surface. A full Semantic-Catalog product is deliberately deferred (§7).

---

## 4. Phase 3 — Press the structural advantages (6–12 months)

### 4.1 Own "fast agentic analytics" with proof (M, ongoing)
- **Publish the benchmark Cube can't**: agent-answer latency + cost on a 1–10B-row event dataset, Rill (managed ClickHouse/DuckDB) vs Cube-over-Snowflake with and without pre-aggs. Include the pre-agg engineering time as a cost line. Cube's G2 complaint ("AI analyst can be a little slow") is the opening; agents make latency *worse* for Cube (multi-query plans × warehouse latency), *better* for Rill.
- **Accuracy benchmark** on the runQL pattern using the §2.3 eval harness, published with methodology.

### 4.2 Embedded agentic analytics v2 (L)
Cube's embedded suite (Chat API, Creator Mode, React SDK) targets Rill's core commercial wedge; respond on Rill's terms:
- **Chat API** (M): streaming HTTP conversational endpoint (thin productization of `CompleteStreaming`) with embed-token auth (`external_user_id` history + per-user security already work) → customers build fully custom chat UX; also wrappable agent-to-agent.
- **Agent white-labeling** (S): name/avatar/intro/suggested prompts per embed org.
- **Embed creator mode, Rill-shape** (L) ♻: scoped canvas-composition for embed users over allowed metrics views (no SQL, no model editing, RLS always on) — matches Omni roadmap §3.4; sell against Cube's Creator Mode with "governed by construction."
- **React embed SDK** (S–M): typed wrapper over iframe + postMessage APIs; npm package. Cheap, removes a checklist line.
- Keep the pricing story loud: flat $500/mo embedded vs Cube's compute-priced embeds.

### 4.3 Platform table stakes (M each) ♻
Same list as Omni roadmap Phase 1/2 — they close Cube checklists too: **SCIM v2 + audit logs + custom-roles review**, **PNG rendering** (prereq for Slack agent charts + email bodies), drill-to-records, maps + reference lines + small multiples in canvas, delivery destinations (webhook/S3/Sheets). Add Cube-specific: **query history / usage analytics surface** for customer admins (Rill has the telemetry; expose it), and evaluate **BYOC** demand before building (single-tenant + VPC peering may suffice; BYOC is XL and only Enterprise-deal-gated).

### 4.4 Background agent loops — beat Cube to its own roadmap (L)
Cube's Aug 2026 roadmap post telegraphs: sandboxed agent session environments, model unit tests, event-triggered background loops (dbt changes, dashboard errors, query signals). Rill can ship this *first* because the primitives exist: branch deployments = sandboxed environments; the reconciler emits resource-error events = triggers; the developer agent writes files = the actor; PRs = the merge gate. Ship: "on model error / on parse failure / on upstream schema change → developer agent proposes a fix PR (with eval run attached)." This is the "self-healing BI project" demo — a category-defining move that leverages exactly the full-stack + git-first advantages Cube lacks.

---

## 5. GTM moves (parallel, not eng-gated)

1. **Publish a Rill vs Cube page**: fast-by-default vs pre-agg engineering; open-source-complete vs Cloud-gated AI; agents that build vs agents that ask; flat embedded pricing vs compute-priced; free local MCP vs Premium-gated MCP. Cube ignores Rill publicly — force the comparison where Rill wins.
2. **Cube-migration play** (S eng): importer from Cube YAML models (cubes/views → models + metrics views; joins flagged until §3.2 lands; `MEASURE()` SQL-API compat from §3.1 eases client cutover). Target Cube OSS users watching the AI platform stay Cloud-only, and Cube Cloud customers hit by the Oct 2025 seat-pricing shift.
3. **Ride the squeeze**: for Snowflake/Databricks-committed shops choosing native semantic views, position Rill as the *consumption + agent* layer on top (warehouse-direct OLAP mode + §3.5 dbt read), not competing middleware. Cube must fight the platforms; Rill can partner-position.
4. **Analyst & standards presence**: brief for the next Gartner Agentic Analytics Market Guide cycle and GigaOm semantic-layer radar; join OSI (§3.6); get Rill into the 2027 semantic-layer roundups it's currently absent from.
5. **Keep maxing the ClickHouse/MotherDuck co-marketing lane** — "real-time operational BI for a metered world" is a category Cube structurally can't enter, and ClickHouse's own blog is already selling it.
6. **Lead AI messaging with what's unique**: write-capable agents + per-recipient-security AI reports + failure attribution + evals-in-CI + free local MCP. "Your agents can build in Rill, not just ask."

---

## 6. Sequencing summary

| Phase | Theme | Items |
|---|---|---|
| 1 (0–3 mo) | AI-ops parity | Skills (+ Claude Code skills repo), rules/certified queries/memories, customer evals + CI, Slack agent, scheduled-tasks packaging, AI admin console |
| 2 (3–6 mo) | Interop & semantics | Postgres-wire Metrics SQL API, joins/composite metrics views (start), Sheets add-on + Excel-via-SQL, outbound MCP connectors, dbt read, OSI membership, field metadata |
| 3 (6–12 mo) | Structural offense | Latency + accuracy benchmarks, embedded chat API/white-label/creator-mode/React SDK, SCIM/audit/PNG/maps/drill ♻, background agent loops ("self-healing BI") |
| GTM (now) | Positioning | vs-Cube page, Cube importer, platform-squeeze partnering, analyst/OSI presence, ClickHouse lane, "agents that build" messaging |

## 7. Deliberate non-goals (and why)

| Cube feature | Why we skip it |
|---|---|
| Warehouse-agnostic middleware breadth (20+ live engines, Fabric, per-user warehouse OAuth) | Rill's story is "bring data to a fast engine"; invest connector breadth in *ingestion*. The middleware position is the one being commoditized by Snowflake/Databricks/dbt |
| Cube Store-style pre-aggregation infrastructure | Rill's OLAP engines are the acceleration layer; metrics-view `rollup:` covers the residual need |
| Full DAX/MDX APIs | Multi-year fidelity chase; §3.3's Sheets add-on + Postgres SQL API captures most demand. Revisit on enterprise loss evidence only |
| Workbooks/Python document paradigm | Same stance as vs Omni: ad-hoc calcs in explore (♻ Omni §3.1) + AI chat cover the jobs; a sandboxed-Python *agent tool* for forecasting is a cheaper future add than a document surface |
| Semantic Catalog as a product | Field metadata + search + MCP discovery (§3.6) capture the AI-grounding value; a standalone catalog fights Atlan/Select Star, not Cube |
| GraphQL API | No demand signal; REST/OpenAPI + SQL API suffice |
| Coining a counter-category | Don't fight "agentic analytics" as a term — win it: Gartner blessed the category, Rill should be listed in it |

**The one-line strategy:** match Cube's AI operationalization fast (Phase 1 — it's packaging work on an agent runtime Rill already has), open the interop doors Cube walks through today (Phase 2 — SQL API, spreadsheets, joins), then press the three advantages Cube cannot copy — a data plane that makes agents fast, an open-source product that includes the whole platform, and agents that build and repair projects instead of only querying them (Phase 3 + GTM).
