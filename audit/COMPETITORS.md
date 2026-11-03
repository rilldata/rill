# Rill Competitive Landscape — Audit Index & Triage (Aug 2026)

Full audits follow a common structure per competitor: research file(s), an honest current-state Rill inventory, a feature-by-feature comparison matrix (✅/🟡/❌/➕), and a phased competitive roadmap. Shared roadmap items are cross-marked (♻) so investments that win multiple fights are visible.

## Completed audits

| Directory | Competitor | Why they matter to Rill |
|---|---|---|
| [`omni/`](omni/) | **Omni** | Fastest-rising direct rival: Looker-lineage semantic BI + workbooks + AI ops suite + Explo-acquired embedded; $120M Series C at $1.5B (Apr 2026). The best-funded "governance + finished interface" competitor |
| [`cube/`](cube/) | **Cube** | Coined "agentic analytics" (D3 → GA Oct 2025); universal semantic layer + Semantic SQL + full AI-ops loop (skills/rules/evals); collides with Rill in embedded + agent-grounding deals |
| [`quicksight/`](quicksight/) | **AWS QuickSight / Amazon Quick** | The procurement-gravity incumbent in AWS shops ($3 readers, EDP, compliance); pivoted to an agentic workspace (Oct 2025), vacating the fast-governed-dashboards position |

| [`sigma/`](sigma/) | **Sigma** | The momentum leader ($200M ARR, $3B Series E May 2026); spreadsheet self-service + writeback/data apps + Sigma Agents; its live-warehouse architecture makes the customer's compute bill its Achilles' heel |
| [`looker/`](looker/) | **Google Looker** | Incumbent code-first semantic BI (LookML + git) + Gemini CA GA/CA API/managed MCP; simultaneously Rill's biggest migration *source* (Omni/Lightdash monetize the wave today) and a Google-gravity rival |
| [`thoughtspot/`](thoughtspot/) | **ThoughtSpot** | AI-BI incumbent (Gartner Leader 2025/2026, Spotter everywhere) with public wounds: 54.67% runQL accuracy, credit bill-surprise, $92.5K median contracts, SpotCache conceding live-query's limits |
| [`hex/`](hex/) | **Hex** | The most complete AI-ops stack shipping (Context Studio, Evals + public DataBench benchmark, MCP both ways, memory); no RLS engine, no data plane, Enterprise-gated embedding |
| [`metabase/`](metabase/) | **Metabase** | The default OSS/cheap embedded choice — now shipping MCP/Slack/BYO-model AI + git sync; no engine (scale is complaint #1), RLS/SSO paywalled. Rill = the graduation path |
| [`tinybird/`](tinybird/) | **Tinybird** | Most architecturally-kindred: managed ClickHouse + SQL-as-REST-APIs, local-first, MCP-native; no semantic layer or BI surface; exposes Rill's streaming-ingestion gap; squeezed by ClickHouse Inc. |

## Considered, deferred (revisit on deal-signal)

- **Superset / Preset** — overlaps Metabase's slot (OSS default) with a heavier ops burden; Metabase audit covers the objection pattern. Defer unless Preset shows up in deals.
- **Lightdash** — philosophically closest OSS "BI-as-code on dbt" rival, but small (~$11M raised); watch, and fold into a dbt-ecosystem note if it grows.
- **Grafana** — the incumbent for ops dashboards; competes in observability-adjacent operational analytics. Defer: different buyer center (SRE), meets Rill mostly in IoT/infra verticals.
- **Databricks AI/BI (Genie) & Snowflake (Cortex Analyst / Semantic Views)** — platform-native gravity, covered as landscape pressure in the cube audit §6; a dedicated "platform-native BI" audit is worth doing if warehouse-direct mode becomes a sales motion.
- **Holistics, GoodData, Luzmo, Embeddable, Explo (†Omni), Hashboard** — embedded/as-code niche players; individually below audit threshold.
- **Zenlytic, TextQL, Dot, Querio** — AI-analyst startups; collectively commoditize "chat with governed data" (see cube audit §6) but individually small.
- **Power BI / Tableau** — incumbent giants; Rill rarely competes head-on for enterprise-wide BI standardization. The relevant surface (Microsoft-estate interop) is addressed in the cube roadmap (DAX/Excel decisions).

## Cross-cutting themes (across all nine audits)

1. **Everyone converged on "agentic analytics"** — Cube coined it, Gartner blessed it (2026 Market Guide + Hype Cycle), AWS rebuilt around it, Omni/Hex/Sigma productized AI ops, ThoughtSpot ships four named agents, even Metabase has MCP + Slack + AI governance. Rill's agent *architecture* is competitive or ahead; its *packaging* lags everyone — the shared Phase-1 investment in every roadmap.
2. **The evidence race is on**: Hex published DataBench; Tinybird runs a public LLM-SQL benchmark; Cube ships CI-gateable evals; Google claims "errors cut two-thirds"; ThoughtSpot's public 54.67% shows what happens without evidence. **A published accuracy benchmark + customer evals is the single highest-leverage shared item** — it reframes bake-offs against six of nine competitors at once.
3. **Recurring Rill gaps** (frequency-ranked across audits): customer evals/benchmark (all 9) · Slack agent (8) · semantic joins/multi-fact (7) · SCIM/audit (7) · maps + chart breadth (6) · viewer-side calcs (5) · embed React SDK + chat API (5) · outbound MCP client (4) · Python/forecasting agent tool (4) · delivery destinations (4) · cloud SQL console (4) · Kafka/streaming ingestion (3, decisive vs Tinybird) · paginated/report layouts (3) · drill-to-records (3).
4. **Recurring Rill advantages**: full data plane (ingestion→OLAP→BI) at fixed cost — the direct counter to Sigma/Looker/ThoughtSpot/Metabase's warehouse-bound architectures; one governed semantic layer grounding humans and agents (vs Topics/coaching/glossary parallel artifacts); write-capable MCP + git-native agent buildability; OSS completeness (only Metabase/Tinybird/Cube are partially comparable, all gate more); pricing simplicity; per-recipient-security AI reports (no one matches).
5. **Migration/graduation plays** are the cheapest offense: LookML converter (Looker wave, contested by Omni/Lightdash), QuickSight JSON importer (rebrand churn + AI fee), Metabase importer (scale graduation), Cube-model importer (OSS users watching Cloud-only AI), Tinybird guide (Classic sunset Sept 15, 2026).
6. **Platform gravity is the common enemy**: AWS EDP, GCP commits, Snowflake/Databricks native semantics squeeze *every* independent — marketplace listings (AWS, GCP) and BYO-engine posture (ClickHouse/MotherDuck alignment) are Rill's counters, and they hurt no other fight.
