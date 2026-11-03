# Roadmap: Competing with (and beating) Hex

Companion to `comparison.md`. Sizes S/M/L/XL; shared items ♻cube/♻omni/♻qs/♻looker.

## 1. Strategic frame

Hex is the **pace-setter on AI-ops packaging** — Context Studio, customer Evals, a *published benchmark* (DataBench), Slack, MCP in both directions, model picker, user memory. It is not, however, a governed BI product: no RLS engine, no metrics-native semantic source of truth, Enterprise-gated embedding, no engine of its own, and it's absent from the Gartner MQ. Hex and Rill often aren't substitutes — Hex serves data teams doing analysis; Rill serves operational/embedded metrics consumption — but they now compete for the same "agentic analytics platform" budget and mindshare.

**Positioning: "Governed, fast, and shippable — the agentic analytics your whole company (and your customers) use, not just your data team."** Concede the notebook; win on governance, speed, embedded, and price.

### Deal-loss categories
1. "Hex's AI stack is more mature." (Context Studio/Evals/Slack/memory)
2. "Our analysts need Python/notebooks." (workflow)
3. "Hex demos better to AI-forward buyers." (benchmark + narrative)

## 2. Phase 1 — AI-ops parity, raised bar (0–3 months)

The ♻cube Phase-1 suite (skills, rules/certified queries, evals + CI, Slack agent, AI admin) — with Hex-specific adjustments:
- **Evals must ship with a public benchmark** (M): Hex's DataBench turned evals into marketing. Rill's version: publish the grounding benchmark (♻qs §3.1) *on a DataBench-style public dataset*, run via the customer-facing evals harness, with the metrics-layer-vs-raw-SQL comparison as the headline. One launch, three fights (Hex, Cube, QuickSight).
- **Context curation surface** (M): fold rules/certified-queries/`ai_instructions` into a coherent **"AI context" page** in web-admin (what the agent knows, per metrics view; feedback-attribution insights as suggested improvements). Not Context Studio's breadth — its *legibility*. Rill's feedback agent already classifies failures; surfacing that closes most of the Review Agent gap cheaply.
- **User-level memory + instructions** (S–M): per-user agent instructions (Hex shipped Nov 2025) then opt-in distilled memory (♻cube §2.2).

## 3. Phase 2 — The analysis-depth answer (3–6 months)

- **Sandboxed Python agent tool** (M) ♻qs §3.3: forecasting/anomaly/stats invoked by the analyst agent on metrics-query results. This is the honest answer to "our analysts need Python" for the *consumption* side — without building a notebook.
- **Cloud SQL console** (M) ♻omni §3.5: Metrics SQL + OLAP SQL surface for editor+ roles; hands results to the agent ("make this a model/metrics view"). Covers the SQL-analyst job without the document paradigm.
- **Chat/agent API** (M) ♻cube §4.2: Hex exposes its agent via CLI/API; Rill's `CompleteStreaming` productization covers embedded + programmatic use.
- **Outbound MCP + web search** (M) ♻cube §3.4: match Hex's client-side MCP; web search is a config flag on the same machinery.

## 4. Phase 3 — Press the wedge (6–12 months)

- **Governance offensive** (content + product ♻): SCIM/audit (♻) plus a "governed agentic analytics" comparison — every Rill AI answer passes RLS, per-recipient AI reports, embedded AI under the same policies. Hex cannot follow without building an RLS engine.
- **Embedded agentic analytics v2** (♻cube §4.2): chat API + white-label + React SDK — Hex's Enterprise-gated embedding is soft; Rill should be the obvious embedded choice for AI-forward SaaS.
- **Semantic import breadth** (M): Hex imports MetricFlow/Cube/LookML — Rill's migration tooling (♻looker `rill migrate lookml`, ♻cube importer, dbt read ♻cube §3.5) doubles as the same interop story, but into a *governed* layer.
- **Composite metrics views** (XL ♻) — keeps the semantic-source-of-truth claim credible as Hex's authoring matures.

## 5. GTM

1. **Don't fight the notebook** — sell alongside: "your data team keeps Hex; your company and customers run on Rill" is a viable land in Hex accounts (different budget lines). Then compete for the agent/self-serve layer.
2. **Governance-first messaging** in AI evals: "an agent without row-level security is a demo, not a deployment."
3. **Benchmark presence** (Phase-1 launch) so Hex's DataBench isn't the only public eval story.
4. **Embedded flank**: target AI-forward SaaS that outgrew Hex's Enterprise embed pricing.
5. **Gartner/analyst work** (♻): Hex isn't in the MQ either — the "agentic analytics" Market Guide is the winnable list for both; Rill should be in it first.

## 6. Non-goals

Notebook/document paradigm, hosted Python kernels as a product (agent tool only), reactive-DAG multiplayer editing, GPU compute, per-model picker UI beyond BYO connector (revisit on demand).

**One-line strategy:** match Hex's AI-ops packaging where it's cheap (context page, evals + public benchmark, Slack, memory), answer analysis depth with an agent Python tool and a SQL console rather than a notebook, and win on what Hex structurally lacks — governance, speed, embedded, and price.
