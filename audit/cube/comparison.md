# Cube vs Rill — Feature-by-Feature Comparison

Legend: ✅ Rill has it (rough parity or better) · 🟡 partial/different equivalent · ❌ Rill does not have it · ➕ Rill advantage Cube lacks. Detail on the Cube side in `cube-features-core.md` and `cube-ai-and-market.md`; Rill side in `rill-feature-inventory.md`.

**Framing:** Cube and Rill converged on the same pitch from opposite ends. Cube: headless semantic layer → added a first-party BI surface + agents (Oct 2025 GA). Rill: full-stack BI-as-code → marketing its metrics layer + MCP to agents. The battleground categories are (1) AI/agentic features, (2) semantic-layer depth, (3) external interoperability (SQL API, BI tools), (4) embedded analytics, (5) performance economics.

---

## 1. AI & agentic features (the main event)

| Feature | How it works in Cube | Rill status | Notes |
|---|---|---|---|
| Conversational analytics chat | Analytics Chat: multi-query analyses, summaries, links to semantic model; stateless "Explore" mode; admin prompt suggestions | ✅ | Rill: ProjectChat/DashboardChat/DeveloperChat; dashboard-context-aware (knows filters/time range); in-chat charts; URL citations to live dashboards — arguably better grounding UX |
| AI writes governed queries | **Semantic SQL**: AI authors SQL over semantic definitions; ad-hoc calcs allowed; every query reviewable as SQL or visually | ✅🟡 | Rill's analyst agent queries via structured metrics tools (safer, less expressive); Metrics SQL exists but is not the agent's primary interface and rejects JOINs. Cube's "reviewable Semantic SQL" is a stronger transparency artifact |
| AI model authoring ("AI data engineer") | Semantic Model Agent: bootstraps model from warehouse schema, **converts LookML**, adds measures/joins/views conversationally; AI-assisted onboarding auto-generates a starter model with reviewable diffs | ✅🟡 | Rill's developer agent + `init_project` + `develop_file` covers scaffolding/editing incl. whole projects; no LookML converter, no query-pattern-driven model optimization |
| Agent skills | `agents/skills/*.md` in repo (frontmatter + instructions), chat buttons/slash menu/auto-match, versioned like code; **plus open-source Claude Code skills repo** (`cube-agent-skills`, 9 skills, plugin marketplace) | ❌ | Rill ships internal agent instructions and `rill init --agent claude/cursor` exports, but no user-defined skills surfaced in chat |
| Agent rules / certified queries | **Rules** (always + agent-requested markdown), **Certified Queries** (admin-approved SQL agents reuse), **Memories** (tenant/space/user isolation), **Agent Spaces** (RBAC groupings), multi-agent per deployment | 🟡 | Rill has `ai_instructions` (project + metrics view) — rules-lite. No certified-query library, no memories, no multi-agent config, no agent spaces |
| Customer-facing AI evals | **Cube Evals** (`agents/eval_questions.yml`: NL question + ground-truth SQL; sort-invariant numeric-tolerant comparison; per-branch; CI-gateable; AI Studio UI) | ❌ | Rill has an internal eval harness (`runtime/ai/evals/`, golden transcripts) — raw material exists, not productized |
| AI observability/admin | AI Studio (tune/observe/eval/ship), Chats History (admin view), Usage Analytics incl. AI usage, per-seat AI token grants + packages | 🟡 | Rill meters AI tokens for billing and has feedback attribution (unique!), but no admin AI-quality/usage surface |
| Scheduled AI tasks | **Scheduled Tasks**: NL prompt + cadence, headless under creator's security context; can query, search web, build/update reports & dashboards, email members | ✅ | Rill: AI-powered scheduled reports (`resolver: ai`) run the analyst agent per recipient **with each recipient's security policies** — per-recipient RLS is ahead of Cube; but can't build dashboards or search web |
| Slack agent | Cube Slack Agent (May 2026): DM the app, threads = chat sessions, email-matched identity, user's own RLS applies (Premium+) | ❌ | Rill's Slack driver is notifier-only. High-visibility gap |
| MCP server (inbound) | Hosted `cubecloud.dev/mcp`, OAuth+PKCE, role-scoped tools, **Premium/Enterprise only**; Claude Connectors Directory listing (Aug 2026) | ✅➕ | Rill: MCP per project incl. **free local product**, full OAuth stack, and **write tools** (file read/write with edit access) — agents can *build* in Rill, not just query. Cube gates MCP behind $80/dev/mo |
| MCP connectors (outbound) | Agent consumes external MCP tools (Notion, Linear, Sentry, Attio + custom); used for context, model metadata, and write-back (file a ticket, write a recap) | ❌ | Rill agents cannot reach external tools |
| A2A / agent-to-agent | Analytics Chat API wrappable as a tool for external agents; A2A protocol marketed | 🟡 | Rill's MCP serves the same job for MCP-speaking agents; no streaming chat API productized for embedding in other agents' loops |
| Python / advanced analytics | Sandboxed Python on workbook tabs (Prophet, stats/ML preinstalled); agent auto-invokes for forecasting/regression/clustering/anomaly questions | ❌🟡 | Rill has AI anomaly "Explain" (root-cause narratives on metric shifts — marketed, well received) but no general Python/forecasting execution |
| AI anomaly explanation | AI Widget (narrative summaries on dashboards, custom prompts) | ✅➕ | Rill's "Explain" (click a spike → root-cause summary + charts + drill links + synced dashboard) is more interactive than Cube's static AI Widget |
| BYO LLM | BYOM incl. Snowflake Cortex; Enterprise tier; pinnable model list; `runtime: plain|reasoning` | ✅ | Rill: `ai_connector` (OpenAI/Azure, Claude, Gemini) from Growth tier — more accessible than Cube's Enterprise gating |
| AI in embeds | Embedded Analytics Chat (iframe + Chat API), white-labeled agent name/branding | ✅🟡 | Rill embeds chat with per-end-user history (`external_user_id`) + postMessage AI pane control; no chat-only API, no agent white-labeling |
| Coding-agent DX | Rust CLI designed as agent interface; open-source Claude Code skills; dashboards-as-code API | ✅➕ | Rill: everything already YAML/SQL in git + `rill init --agent` + local MCP — "Claude Code builds your whole project" is Rill's best story; Cube is now investing here fast |
| Eval/accuracy marketing | LLM benchmark content, "3x more accurate" research, Gartner Agentic Analytics Market Guide listing | 🟡 | Rill blogs "golden completions, not vibes" but publishes no benchmark, no analyst listings |

**Net:** Rill's agent *architecture* (multi-agent runtime, MCP-native, write-capable, per-recipient-security AI reports, feedback attribution) is competitive or ahead. Cube's agent *operationalization* (skills, rules, certified queries, memories, evals, AI Studio, Slack, outbound MCP, Python) and *packaging/marketing* are well ahead — Cube productized the loop Rill only runs internally.

## 2. Semantic layer

| Feature | Cube | Rill status | Notes |
|---|---|---|---|
| Joins / relationship graph | Cubes declare joins (cardinality-aware, Dijkstra path resolution); views pin join paths | ❌ | **Rill's #1 structural gap** (same vs Omni): metrics views bind to one model/table; Metrics SQL rejects JOINs; escape valves = lookup dims, upstream models |
| Multi-fact analysis | Multi-fact views on shared dimensions, fan-out-safe (Tesseract, GA v1.7) | ❌ | No cross-metrics-view stitching |
| Multi-stage calcs | PoP, % of total, running, YTD/QTD via Tesseract `grain`/`filter` directives | ✅ | Rill: derived/window/time-comparison measures — near parity, arguably cleaner YAML |
| Custom calendars | **Calendar cubes** (4-5-4 retail, fiscal, time shifts) | ❌ | Rill: `first_month_of_year` only; no fiscal calendar (also flagged vs Omni) |
| Modeling languages | YAML **+ Jinja/Python**, or JavaScript; dynamic per-tenant model generation | 🟡 | Rill: YAML + Go templating (`.env`, user attrs, dev/prod) — no programmatic model generation, no per-tenant dynamic models |
| Security | Data Access Policies: row-level, member-level, **masking**; user attributes/groups | ✅🟡 | Rill: access/row filters/field access, templated on JWT attrs — parity except **no masking mode** |
| Organization | Views, view groups, hierarchies, nested folders, links in model | 🟡 | Rill: explores select fields; no folders/hierarchies metadata for large models |
| Drill members | `drill_members` + client `drillDown()` | ❌ | No drill-to-records path (same gap vs Omni) |
| Catalog / lineage | **Semantic Catalog**: search across model + upstream warehouse + downstream BI content; Data Graph lineage | 🟡 | Rill has project DAG + resource lineage UI in Developer, not a catalog product |
| dbt integration | One-way dbt→Cube sync with CI/webhooks + review branches; `cube_dbt` | ❌ | None |
| Ingestion/transformation | **None — Cube models over existing warehouse tables** | ➕➕ | Rill's full ELT: sources, SQL models, **incremental + partitioned models**, staging cross-connector flows, data-quality tests. Cube depends on warehouse + dbt |
| Consistency guarantee | Model compiled per request over live warehouse | ➕ | Rill reconciler continuously validates whole project DAG; parse errors surface immediately |

## 3. Performance & caching

| Feature | Cube | Rill status | Notes |
|---|---|---|---|
| Pre-aggregations | Declared in model, **aggregate awareness** matching, refresh workers, partitioned/incremental builds, lambda (streaming merge), Cube Store (distributed Rust/Parquet store) | 🟡 | Rill's answer is architectural: queries hit a real OLAP engine (DuckDB/ClickHouse/Druid/Pinot/StarRocks). New metrics-view `rollup:` is early. Different philosophy: Cube accelerates a slow warehouse; Rill replaces it |
| Result caching | In-memory + `cache` control per query | ✅ | Runtime LRU + `cache.key_sql`/`key_ttl` |
| Sub-second at scale | Only via pre-aggs (build/refresh complexity is Cube's #1 user complaint) | ➕➕ | Rill's core wedge: managed OLAP, no pre-agg modeling burden, real-time-ish via live ClickHouse |
| Streaming | ksqlDB/Materialize/RisingWave sources + lambda pre-aggs | 🟡 | Rill: no Kafka connector; Kafka via ClickHouse only |
| Warehouse-direct | The default mode (Snowflake, BigQuery, Databricks, Fabric, 20+ engines) | 🟡 | Rill has warehouse-direct OLAP mode (BigQuery/Snowflake/Databricks) but it's secondary and less battle-tested |

## 4. APIs & interoperability

| Feature | Cube | Rill status | Notes |
|---|---|---|---|
| SQL API (Postgres wire) | **Semantic SQL over Postgres protocol** — any BI tool/psql client connects; pushdown GA; declared the primary API | ❌ | **Rill's #1 interop gap.** No Postgres wire endpoint; REST/gRPC/custom APIs/MCP only |
| Power BI / Excel | **DAX API** (native Power BI), **MDX API** (Excel pivots), Excel + Sheets add-ins | ❌ | Nothing. Blocks every Microsoft-estate deal |
| BI tool sync | Semantic Layer Sync (Superset/Preset/Metabase/Tableau) | ❌ | None — deliberate historically ("Rill is the BI tool"), but now blocks "semantic layer for your existing stack" deals |
| GraphQL | Yes | ❌ | Low priority; REST/OpenAPI covers embedded devs |
| REST APIs | JSON API + Control Plane API (deployments, users, dashboards-as-code, SCIM, embed sessions; OpenAPI) | ✅ | Runtime + admin APIs (168 RPCs) + custom APIs w/ OpenAPI; parity in capability |
| Custom API endpoints | Not a feature (you build an app on the APIs) | ➕ | Rill custom APIs: YAML endpoint + resolver + security + OpenAPI — nice embedded/back-end story |
| Orchestration hooks | Airflow/Dagster/Prefect providers for refresh | 🟡 | Rill: refresh via CLI/API cron; no orchestrator provider packages |
| CLI | New Rust `cube` CLI, full API coverage, agent-oriented | ✅ | `rill` CLI mature and broader (local dev server included) |

## 5. Consumption surfaces & embedded

| Feature | Cube | Rill status | Notes |
|---|---|---|---|
| Dashboards | First-party dashboards (GA Oct 2025): tables/charts/KPI/custom HTML/markdown; dashboards-as-code API | ✅➕ | Rill explore + canvas are years more mature (TDD, pivots, bookmarks, public URLs, themes, dark mode) |
| Slice-and-dice explore | Workbooks (point-and-click + calcs + SQL + Python) | ✅➕ | Rill Explore is best-in-class for speed; but workbooks include ad-hoc calcs/SQL/Python that Rill viewers lack |
| Viz breadth | bar/line/area/scatter/pie/heatmap/boxplot/table/KPI/**maps (point+choropleth)**/custom HTML/Vega-Lite; small multiples, dual axes, reference lines | 🟡 | Rill lacks maps, boxplot, small multiples, reference lines (same gaps vs Omni); has funnel/donut/stacked-100% Cube lacks natively |
| Embedded dashboards/chat | iframes (theming, SSO, signed auth, postMessage, localization), **Creator Mode**, Chat API, React Embed SDK (preview) | ✅🟡 | Rill embed (iframe + postMessage + RLS + per-user state + AI history) is strong; **no creator mode, no React SDK, no chat-only API, no localization** |
| Alerts / reports | Scheduled Tasks (AI) + reports; no classic threshold alerting found | ➕ | Rill: full alerting (criteria, per-user RLS eval, email/Slack) + scheduled exports (CSV/XLSX/Parquet) + AI reports |
| Excel/Sheets consumption | Native add-ins + MDX pivots | ❌ | None |

## 6. Cloud platform & enterprise

| Feature | Cube | Rill status | Notes |
|---|---|---|---|
| SSO / SCIM | SAML (Google/Entra/Okta) + SCIM 2.0 + LDAP (Enterprise) | 🟡 | Rill: OIDC/Auth0 SSO; **no SCIM** (also flagged vs Omni) |
| Audit log | Enterprise audit log + Chats History | ❌ | None |
| Custom roles | Enterprise | 🟡 | Fixed org/project roles + groups |
| BYOC / private networking | **BYOC on AWS/Azure/GCP**; VPC peering/PrivateLink/PSC | 🟡 | Rill: VPC peering (Enterprise); no BYOC data plane |
| Multi-tenant infra | Multi-cluster routing by tenant (>100 tenants), per-tenant models/data sources | 🟡 | Rill: org/project isolation + provisioner (k8s runtimes, managed ClickHouse); no per-tenant dynamic model story |
| Observability for admins | Query History, Performance Insights, usage analytics, monitoring integrations | 🟡 | Project logs/status + billing usage; no customer-facing query-history/BI-usage analytics |
| Git workflow | IDE branch dev mode, GitHub/GitLab/Bitbucket, staging envs | ✅➕ | Rill: git **is** the source of truth; branch deployments; cloud edit sessions; local single binary — better DX |
| Open source | Cube Core 20.7k stars, Apache/MIT; but the entire AI/BI frontend is Cloud-only | ✅➕ | Rill OSS includes the full product (dashboards, AI chat with BYO key, MCP) locally — stronger OSS story where it counts |
| Pricing transparency | $40–$80/dev + $20–$40 consumer seats + compute; MCP Premium+; DAX/BYOM/SSO Enterprise | ➕ | Rill $20–$30/seat + usage; embedded flat $500/mo per 100 seats; BYO DB + BYO LLM at Growth. Simpler, cheaper story |

## 7. Where Rill is structurally ahead

1. **Full data plane**: ingestion → incremental/partitioned models → OLAP → dashboards in one binary/platform. Cube has no ingestion or transformation; it inherits warehouse cost and latency.
2. **Speed economics**: sub-second interactivity without pre-aggregation engineering (Cube's most complained-about surface). At event scale, Rill's cost/latency floor is Cube's ceiling.
3. **Finished BI surface**: mature explore/canvas/alerts/reports/public URLs vs Cube's one-year-old dashboards. Luzmo's framing: Rill "ships an interface rather than expecting you to build one."
4. **MCP + write-capable agents everywhere, including free/local**: Cube gates MCP at Premium and its agents can't edit the model from scheduled runs; Rill's developer agent writes project files, and coding agents build whole projects via `rill init --agent`.
5. **Per-recipient security on AI reports** and feedback attribution — governance depth Cube doesn't match.
6. **Pricing simplicity** and BYO-infrastructure (DB + LLM) below enterprise tier.
7. **OSS completeness**: Rill's open-source product is the whole product; Cube's AI platform is 100% Cloud-gated.

## 8. Cube's structural advantages

1. **Universal interoperability**: Postgres-wire Semantic SQL + DAX/MDX/Sheets/Excel + Semantic Layer Sync — Cube plugs into every existing BI estate; Rill requires adopting Rill's UI.
2. **Joins/multi-fact semantic model** — models real enterprise schemas; Rill's single-table metrics views can't.
3. **AI operationalization suite**: skills, rules, certified queries, memories, evals-in-CI, AI Studio, Slack agent, outbound MCP connectors, Python analysis — a productized trust/tuning loop.
4. **Install base + analyst mindshare**: 20.7k stars, "20% of F1000," Dresner #1, GigaOm Leader, Gartner Agentic Analytics listing, OSI founding member. Rill is absent from all these lists.
5. **Warehouse-native credibility**: 20+ engines incl. Fabric/streaming DBs; per-user warehouse OAuth; BYOC.
6. **Embedded agentic breadth**: Chat API + Creator Mode + React SDK + localization.
7. **Ship velocity + narrative ownership**: a launch every ~2 weeks and the "agentic analytics" category name (they coined it and got Gartner to bless it).
