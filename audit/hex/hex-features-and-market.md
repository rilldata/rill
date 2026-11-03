# Hex — Platform, AI & Market Inventory

> Researched Aug 2026 from hex.tech (blog, changelog, pricing, docs at learn.hex.tech), press, G2. Hex Technologies: founded 2020 (ex-Palantir; Barry McCardel CEO), positioning "The AI Analytics Platform where trust meets insight" / "agentic analytics."

## 1. Core product

- **Projects = polyglot notebooks**: SQL + Python + R + no-code cells (pivot, filter, chart, map, KPI, input, Markdown). **Reactive DAG execution** (only dependents re-run) — their signature vs Jupyter. Chained SQL: cell output is a dataframe queryable by downstream SQL/Python. Jinja templating; schema browser with dbt metadata inline.
- **Compute**: hosted kernels — XS/S/M included; PAYG L→4XL ($0.32–$2.58/hr) + GPUs (Team+); custom Docker images; can run inside Snowflake (Snowpark Container Services).
- **Connections**: 16 warehouses in tiers — Tier 1: Snowflake, BigQuery, Databricks, ClickHouse; Tier 2: Redshift, Athena, Postgres, MySQL, SQL Server, MotherDuck, Trino, CloudSQL; Tier 3: AlloyDB, MariaDB, Presto, Starburst. CSV upload, SSH tunnels, OAuth passthrough, AWS IAM roles.
- **dbt**: dbt Cloud only (Discovery API) — docs/tests/freshness in schema browser.

### Semantic layer (major 2025–26 investment)
- **Semantic Model Sync** (public beta Mar 5, 2025): imports models from **dbt MetricFlow, Cube, LookML, Snowflake Semantic Views, Databricks metric views** — "no proprietary lock-in"; no-code Metrics Cell.
- **Semantic Authoring / Modeling Workbench** (Aug 20, 2025): author models natively (autocomplete, validation, Workbench Agent) or in git with CI/CD + sync API. **Semantic Modeling Agent** (Oct 2025): synthesizes model code from existing projects, diff view, semi-additive measures. OSI member. Semantic projects pullable via CLI (Aug 2026). Professional+ feature.

### Explore & data apps
- **Explore** (Oct 30, 2024): drag-drop no-code slice/dice surface for business users, reads governed measures/dims/joins from semantic models; also an Explore cell in notebooks. Drill-down is Team+.
- **App Builder**: multi-tab published data apps, input cells, themes (custom themes Team+, used for embed white-labeling). **Generative Apps** (2026): prompt-to-dashboard (Apr 30), prompt-driven creation (May 12), signed embedding of generative apps (May 21), direct code editing (Jul 21), `design.md` brand guidelines (Jul 8).
- **Scheduled runs** (hourly max; Team+ for real use), notifications with PNG/PDF screenshots + CSV/**Google Sheets** exports (Sheets one-click + scheduled Mar 2026); Airflow/Dagster/dbt Cloud triggers; REST APIs.

## 2. AI features (timeline)

- **Hex Magic** (Feb 2023 beta): Generate/Edit/Fix/Explain/Complete in SQL/Python cells.
- **Notebook Agent** (Aug 27, 2025, public beta): the copilot→agent pivot — plans analyses, searches warehouse, creates + chains cells, builds charts, writes summaries; cell-by-cell review, auto-versions. Launched on Claude Sonnet 4. Rebranded 2026 as the general **"Hex Agent"** across surfaces.
- **Threads** (Oct 1, 2025, Fall launch "Agents, for analytics, for teams"): chat self-serve Q&A grounded in workspace context (connections, semantic models, endorsed tables); Explore cells in answers; thread→notebook conversion; Explorers+ on Team/Enterprise. Evolution: CSV upload + user-level instructions (Nov 2025), **Endorsed mode default** + **Sensitive mode** (Jan 2026), Python execution in Threads (Mar 2026), chat↔notebook toggle + build-app-from-conversation (Aug 2026).
- **Agent everywhere**: **Slack agent + remote MCP server** at app.hex.tech/mcp (Oct 29, 2025; Claude/ChatGPT/Cursor/Codex/Glean); interactive Claude Connector + Cursor plugin (Mar 2026); **MCP client support** (May 28, 2026 — Hex agent calls external MCP tools); Hex in Codex (Jun 2026); **Hex CLI** (Apr 2026) with full agent via CLI/API (Jul 2026); **Agent Tasks** (Jun 16, 2026 — scheduled agent questions delivered to Slack/email); agent **web search** (Jun 2026, via Parallel); **automatic user memory** (Apr 2026).
- **Models**: picker (Jun 2026), workspace default, "Auto" mode; roster spans Claude Sonnet 4→4.5, **Opus 5**, **Fable 5**, Kimi K2.7, GPT-5.6 family; **Fast mode** (Aug 2026, ~2.5x faster at higher credits); Enterprise BYOK.
- **Context Studio** (Jan 28, 2026): curated agent-context workspace — Semantic Models, **Guides** (markdown business logic, GitHub-syncable), **Endorsed Tables**, **Evals**, **Review Agent** (auto-flags confusion/missing context); observability across Hex/Slack/MCP/CLI with full audit trails.
- **Evals** (Aug 4, 2026): CLI eval suites — test questions with expected outcomes, deterministic/LLM-judged/hybrid rubrics, fork-context testing, production drift runs. **DataBench** (2026): public agentic-analytics benchmark (100 tasks, synthetic business, judged with native Evals).
- **AI pricing**: credit-based per-seat monthly grants by plan; add-on credits, workspace pools, auto top-ups + per-user limits; Fast mode/larger models burn more.

## 3. Collaboration, BI surfaces, embedded

- **Converging on BI**: Explore + semantic layer + Threads + Generative Apps + the **Hashboard acquisition (Apr 30, 2025)** (lightweight semantic-BI product; team joined to accelerate self-serve).
- Version control: built-in history (unlimited Team+) + **Git export** (each version = commit, published = main; Team+). Real-time multiplayer, cell comments. Component library.
- **Embedded**: public/private/**signed embedding** (`createPresignedUrl` single-use iframe URLs; **RLS via embed user attributes**); PDF/CSV export for embed users. **Enterprise add-on**, priced opaque, cited as expensive. Notion embeds.

## 4. Enterprise / governance

Roles: Admin/Manager/Editor (paid) + **Explorer** (separate paid seat, Team/Ent) + Viewer/Guest (free). OIDC SSO (Enterprise), enforced SSO; SCIM thin publicly. Audit logging Team+, AI audit logs, Observability API (Ent), Thread Inspector. **No self-hosted product**: Enterprise single-tenant VPC or EU multi-tenant; in-Snowflake option; HIPAA add-on; SOC 2 Type II. **No warehouse-style RLS engine** — RLS via embed attributes/parameterized queries/OAuth passthrough (noted gap vs governed BI).

## 5. Pricing (hex.tech/pricing, mid-2026)

Community free (agent trial, Small compute, 3 components) · **Professional $36/editor/mo** (Standard credits, 5 published apps, semantic models) · **Team $75/editor/mo** (Extended credits, unlimited apps, Explore drill-down, schedules/alerts, APIs, audit logs; **Explorer seats = paid add-on**, price unpublished) · Enterprise custom (Premium credits, OIDC, BYOK, single-tenant/EU, HIPAA, **embedded add-on**, Observability API). Viewers free. Compute PAYG per §1.

## 6. Company / market

- **$172M raised**: Series B $52M (2022) + $28M ext (2023, Sequoia) + **Series C $70M (May 28, 2025, Avra; a16z, Sequoia, Snowflake Ventures)** — framed explicitly as agentic analytics. Valuation undisclosed.
- 1,500+ teams; headcount ~160–270 (trackers disagree); ~$19.8M ARR (2024, Latka, unverified).
- Customers: Reddit, Anthropic, Figma, NBA, HubSpot, Cisco, StubHub, Rivian, Duolingo, Notion, Brex, ClickUp, PandaDoc, Candidly (embedded).
- **Not in the Gartner ABI MQ** (2025 or 2026). G2 4.5/5 (~385 reviews). Snowflake "Product Partner of the Year — Agentic Analytics" 2026.
- Complaints: performance on large datasets/70+-cell projects; Professional→Team price cliff; Explorer seats sales-gated; embedding Enterprise-only/expensive; agent emits raw SQL/Python not always via a semantic layer (wrong answers hard for non-technical users to trace); semantic layer thinner than Omni/Looker-class (join support); per-editor pricing adds up.
- Positioning: vs notebooks — "Jupyter is infrastructure, Hex is a product"; vs BI — rivals (esp. Omni) argue Hex complements rather than replaces governed BI; Hex counters with Semantic Sync (no lock-in), OSI, Context Studio + Evals/DataBench as the trust story.
- **Momentum**: extremely high cadence — Series C + Hashboard (May 2025) → Notebook Agent (Aug) → Threads + Semantic Modeling Agent (Oct) → Slack + MCP (Oct) → Context Studio (Jan 2026) → CLI, MCP client, Generative Apps, Agent Tasks, web search, memory (spring 2026) → Evals, DataBench, Fast mode (summer 2026). Arc: AI notebook → full agentic analytics platform with its own context layer and agent distribution into Slack/IDE/CLI/MCP.
