# Roadmap: Competing with (and beating) QuickSight / Amazon Quick

Companion to `comparison.md`. Effort sizes: S (≤1 eng-month), M (1–3), L (3–6), XL (6+). Items shared with other competitive roadmaps marked ♻cube / ♻omni — build once, win all three fights.

## 1. Strategic frame

### This fight is different

Cube and Omni are feature fights. Quick is a **distribution fight**. Quick(Sight) rarely beats Rill on product — its dashboards are static, its AI grounding needs hand-curated Topics, its as-code story is machine JSON, and it has no data plane. It wins because it's **in the AWS bill** (EDP-eligible), readers cost $3, compliance is inherited from AWS, and in Oct 2025 AWS made it the default "agentic workspace" for every AWS-committed enterprise.

Quick's pivot also creates the opening: AWS turned its BI product into an agent workspace, **vacating the "excellent, fast, governed dashboards" position** — and its July 2026 service retirements (Q Business, Kendra, classic Bedrock Agents) plus four renames in 18 months give every buyer a stability objection. Meanwhile its weaknesses are precisely Rill's strengths: grounding-by-construction vs Topics curation; sub-second OLAP vs SPICE caps and 15-minute refresh floors; git-native YAML vs BIOps JSON; MCP *server* vs MCP client-only.

**Positioning: "The analytics your agents — and your customers — actually trust and feel. Fast, governed, code-first; runs great on AWS."** Don't out-suite Amazon. Win three specific battlegrounds:
1. **Embedded analytics for SaaS vendors on AWS** (Rill's commercial wedge vs Quick's session-priced static dashboards)
2. **Operational/event-scale analytics** where SPICE and direct-query timeouts break
3. **AI-grounded BI evaluations** where accuracy, auditability, and agent-buildability decide

### The four deal-loss categories to eliminate
1. **"It's already in our AWS bill."** (procurement/Marketplace/EDP)
2. **"We're HIPAA/FedRAMP/regulated — AWS is pre-approved."** (compliance/enterprise IT)
3. **"Readers are $3; you want $20."** (passive-viewer economics)
4. **"Quick does maps/paginated reports/mobile/forecasting."** (checklist items)

### Deliberate non-goals (§6)
The agentic workspace (research agents, RPA/flows, UI-navigating agents, app builder), native mobile apps, paginated-report parity beyond basics, 30-chart breadth parity, $3-reader price matching.

---

## 2. Phase 1 — Kill the procurement & checklist losses (0–3 months)

### 2.1 AWS Marketplace listing + AWS-native GTM (S–M, mostly BD/ops)
The single highest-ROI move against Quick. List Rill Cloud on **AWS Marketplace** so customer spend retires EDP commitments — this neutralizes "already paid for under our EDP" without touching product. Pair with: a "Rill on AWS" reference architecture page (S3/Athena/Redshift → Rill managed OLAP; private connectivity options), co-sell registration (ISV Accelerate), and case-study content for AWS-shop SaaS embedders. Also evaluate **BYOC-on-AWS** demand signal (XL if built — defer, but Marketplace + VPC peering covers most of it today).

### 2.2 Viewer-tier pricing answer (S, pricing/GTM decision)
Quick's $3 reader is the sharpest objection in internal-BI deals. Options to evaluate (decision, not eng): a low-cost **viewer seat** (consume dashboards/bookmarks only, no explore/AI) or positioning discipline (Rill seats = self-service explore + AI, not passive viewing — "you're comparing an analyst seat to a viewer seat"). For embedded, the flat $500/100-seats already wins; publish a **cost-comparison calculator** vs session pricing (sessions × $0.30–0.50 + Q questions × $0.50 adds up fast for engaged users — Quick is only cheap for *rarely-viewed* dashboards).

### 2.3 Compliance & enterprise IT pack (M–L) ♻cube ♻omni
Same items as the other roadmaps, but Quick makes them existential in regulated deals: **audit logs** (M), **SCIM** (M), SSO self-serve config (S–M), then a compliance roadmap decision (SOC 2 Type II if not already → HIPAA BAA capability → defer FedRAMP unless public-sector pipeline appears). Add **PrivateLink-style private connectivity** and **CMK for managed storage** to the Enterprise backlog with demand tracking.

### 2.4 Visualization checklist: maps first (M) ♻omni
Quick has 3 map types, waterfall, gauge, treemap, radar, word cloud, histogram, small multiples. From the omni roadmap §2.1, the priority order vs Quick: **maps (choropleth + point)** → gauge/bullet (pairs with KPI) → waterfall → histogram → small multiples. Vega-Lite renderer makes each a template + config schema. Keep the AI-generated custom-chart escape hatch as the long-tail answer.

### 2.5 Scheduled delivery & PDF/PNG polish (S–M) ♻omni
Quick's email snapshots + paginated reports set expectations. Ship: **PNG/PDF snapshot rendering of explore + canvas** (extend the existing PDF pipeline), scheduled dashboard-image email delivery, and a "report layout" canvas mode (page-break hints, header/footer) — deliberately *not* full pixel-perfect pagination. Covers ~80% of "email me the dashboard" and unblocks the Slack agent (♻cube §2.4).

### 2.6 QuickSight migration tooling (M)
AWS ships Tableau/Power BI → Quick migration agents; Rill should harvest the same estates *plus QuickSight itself*. QuickSight's `DescribeDashboardDefinition`/asset-bundle APIs export complete JSON — write a converter (datasets → models/metrics views; analyses → canvas/explore YAML; RLS rule datasets → security policies) driven by the **developer agent** (JSON → YAML is exactly what LLM agents are good at, with the reconciler validating output). Market as "escape the JSON: your QuickSight dashboards as reviewable code, in an afternoon." Targets: teams burned by the rebrand churn and the $250/account AI fee.

**Phase 1 exit criteria:** Rill is purchasable through AWS Marketplace, has a credible answer sheet for compliance and viewer pricing, renders maps, emails dashboard snapshots, and demos a QuickSight-project import.

---

## 3. Phase 2 — Win the AI-BI evaluation (3–6 months)

Mostly shared with the Cube roadmap Phase 1 (♻cube) — the same AI-ops investments beat both; deltas specific to Quick:

### 3.1 The grounding bake-off story (S–M, product marketing + evals)
Quick's known weakness is Topics curation and accuracy complaints. Weaponize Rill's architecture: publish a **grounding benchmark** — same dataset, NL question suite, Rill metrics-layer chat vs Q&A over an uncurated *and* a curated Topic — measuring accuracy and *time-to-curate*. Requires the customer-facing evals harness (♻cube §2.3). The claim to own: "zero-curation accuracy — the semantic layer you already built for dashboards is the grounding."

### 3.2 AI-ops parity items (♻cube Phase 1, unchanged)
Skills, rules/certified queries, customer evals + CI, AI admin console, Slack agent. Quick-specific angle for the **Slack agent**: Quick ships Slack/Teams extensions — a Rill Slack agent is required to not look surface-poor; **Teams** support should follow (M) for enterprise deals where Quick leads with M365 integration.

### 3.3 Forecasting & anomaly detection as agent tools (M)
Quick has ML forecasting/anomaly detection; Cube added Python analysis. Rill's cheapest credible answer: a **sandboxed compute tool for the analyst agent** (server-side, stats library incl. seasonal forecasting + changepoint/anomaly detection) invoked on metrics-query results — surfacing as "forecast this metric" in chat and as an optional forecast overlay on time series. Skip a Scenarios-style what-if workspace (non-goal §6) but note demand.

### 3.4 Dashboard actions → workflows (S–M)
Quick markets one-click actions from dashboards into tickets/workflows. Rill's outbound MCP connectors (♻cube §3.4) generalize this: canvas/explore "action" buttons + chat-invoked actions that call approved external MCP tools (file a Linear/Jira ticket with the current filter context). Small surface, kills a demo-differentiator.

### 3.5 Embedded v2 (♻cube §4.2, re-prioritized for AWS SaaS deals)
Priority order vs Quick: **React embed SDK** (S–M — Quick has a JS SDK; table stakes for SaaS evals) → **Chat API** (M) → anonymous-embed ergonomics (public URL embeds with locked filters already exist — document as the "anonymous embedding" answer, add per-URL usage caps) → creator mode (L, later). Keep hammering flat-price embedded AI vs $0.50/question.

---

## 4. Phase 3 — Structural offense (6–12 months)

### 4.1 Semantic joins / composite metrics views (XL) ♻cube ♻omni
Unchanged from the other roadmaps — and now urgent from a third direction: Quick's **multi-dataset topics** (Aug 2026) added runtime joins with one semantic model for dashboards + Q&A. Every major competitor now joins; Rill's post-aggregation composite design remains the differentiated, correctness-first answer.

### 4.2 Event-scale benchmark vs SPICE (M, content + harness)
The performance claim needs public proof aimed at Quick: 1–10B-row event dataset — Rill managed ClickHouse vs QuickSight SPICE (capacity cost at 2TB cap, string tax, refresh latency) and vs Athena direct query (2-min timeouts). Measure: interactive-explore latency, freshness floor, monthly cost. Publish with methodology; reuse for the ClickHouse co-marketing lane.

### 4.3 Level-aware / viewer-side calculations (L) ♻omni §3.1
Quick's LAC-A/LAC-W and table calcs are its semantic-depth bright spot. Rill's ad-hoc measures plan (expression builder over existing measures, window functions server-side, promotion-through-code-review) answers it — pull forward if analyst-led evals recur.

### 4.4 AWS-native depth (M)
Connector gaps that matter in AWS shops: **Glue Data Catalog / Iceberg & S3 Tables reads** (DuckDB's Iceberg support makes this tractable), OpenSearch, Timestream (demand-gated). Plus IAM-role-based auth for AWS sources everywhere (mostly exists) and documented Kinesis→S3→incremental-model streaming patterns.

---

## 5. GTM moves (parallel, not eng-gated)

1. **Publish "Rill vs QuickSight/Amazon Quick"**: grounded-by-construction AI vs Topics curation; sub-second explore vs static dashboards; YAML-in-git vs BIOps JSON; flat embedded pricing vs session+question metering; stability vs 4 renames + service retirements. No third-party comparison exists — write the canonical one and own the SEO.
2. **"Escape the rebrand" campaign** targeting QuickSight customers: migration tool (§2.6) + the stability narrative (Corey Quinn's "application graveyard," Q Business/Kendra retirements) + the $250/account AI-fee objection. Time it around re:Invent 2026.
3. **AWS Marketplace + co-sell** (§2.1) so "buy through AWS" stops being Quick-exclusive.
4. **Embedded SaaS play**: target AWS-shop B2B SaaS vendors comparing QuickSight embedded — lead with embedded *explore* + per-user AI chat + predictable pricing + white-label without annual commitments; publish the cost calculator (§2.2).
5. **Don't fight the suite** — concede the agent-workspace category to AWS/Microsoft/Google in messaging and reframe: "Quick is a workspace with dashboards inside; Rill is the analytics your agents and customers actually trust and feel." Sell Rill *alongside* Quick accounts as the operational/embedded/AI-grounded layer.

---

## 6. Deliberate non-goals (and why)

| Quick feature | Why we skip it |
|---|---|
| Agentic workspace breadth (Quick Research, Flows, Automate, UI agents, app builder) | Competing with AWS/Microsoft/Google on horizontal agent workspaces is unwinnable for a BI vendor; Rill's outbound MCP connectors let customers wire Rill into whichever workspace wins |
| Scenarios-style what-if workspace | §3.3's forecasting agent tool covers the analytical core; a full scenario workspace is a different product |
| Native mobile apps | Responsive web + Slack/email delivery covers Rill's operational buyer; revisit on demand evidence |
| Pixel-perfect paginated reports | §2.5's report-layout canvas + PDF/PNG covers most demand; full pagination engines are a multi-year sink |
| $3 passive-reader price matching | Race to the bottom against AWS's margin structure; win on seat *value* (explore + AI) and embedded flat pricing instead |
| 30+ chart-type parity, Highcharts-style code charts | Maps/gauge/waterfall/histogram close the real objections; Vega-Lite + AI generation is the long tail |
| FedRAMP (near-term) | Only with a real public-sector pipeline; sequence SOC 2 Type II → HIPAA first |

---

## 7. Sequencing summary

| Phase | Theme | Items |
|---|---|---|
| 1 (0–3 mo) | Procurement & checklist | AWS Marketplace + co-sell, viewer-pricing answer + embed cost calculator, audit/SCIM/compliance pack ♻, maps + gauge/waterfall/histogram ♻, PNG/email snapshots + report layout, QuickSight importer |
| 2 (3–6 mo) | Win AI-BI evals | Grounding benchmark vs Topics, AI-ops suite ♻cube (skills/rules/evals/admin), Slack + Teams agent, forecasting/anomaly agent tool, dashboard actions via MCP, React embed SDK + Chat API |
| 3 (6–12 mo) | Structural offense | Composite metrics views ♻, SPICE benchmark content, viewer-side calcs ♻, Glue/Iceberg + AWS-native connector depth |
| GTM (now) | Positioning | vs-Quick page, "escape the rebrand" migration campaign, embedded SaaS play + calculator, sell-alongside-the-suite framing |

**The one-line strategy:** neutralize AWS's procurement gravity (Marketplace, compliance, pricing answers), turn Quick's agentic pivot into Rill's opening — own the "fast, governed, code-first analytics that agents and customers trust" position Quick vacated — and prove it with grounding and SPICE-scale benchmarks Quick structurally cannot answer.
