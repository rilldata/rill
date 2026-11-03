# Roadmap: Competing with (and beating) Sigma

Companion to `comparison.md`. Sizes S/M/L/XL; shared items ♻cube/♻omni/♻qs/♻looker/♻hex/♻ts/♻mb.

## 1. Strategic frame

Sigma is the best-funded, fastest-growing rival in this audit set and has productized the exact agent bundle (user-defined agents + tools + scheduling + MCP) Rill's roadmaps plan. But its architecture is a tax it cannot repeal: **every viewer click, embed render, and AI query is a live warehouse query on the customer's bill** — the #1 complaint in its own reviews, and it compounds precisely where Rill is strongest (embedded concurrency, operational data, AI workloads). Sigma's governance is also optional-by-design (spreadsheet freedom → metric divergence), while Rill is governed by construction.

**Positioning: "Self-service that doesn't run up your warehouse bill — governed by construction, fast by architecture."** Win embedded and operational deals on cost predictability + speed; achieve *sufficient* self-service so Sigma's spreadsheet demo doesn't end evals; match the agent packaging.

### Deal-loss categories
1. "Analysts live in spreadsheets; Sigma lets them keep doing that." (self-service depth)
2. "Sigma has agents/writeback/data apps shipping today." (AI + app packaging)
3. "Sigma is the safe momentum pick." (brand/capital)

## 2. Phase 1 — Cost & momentum counters (0–3 months)

- **TCO calculator + benchmark** (S–M, GTM-heavy): model the *all-in* cost — Sigma license (median $68K) + warehouse compute per active viewer/embed/AI query — vs Rill seats + fixed CU-hr. Publish with a reproducible workload. Sigma cannot rebut architecturally; their own AI-cost-monitoring templates concede the problem.
- **Agent packaging** (♻cube Phase 1, urgency ↑): skills + scheduled agent tasks + outbound MCP tools = Sigma Agents parity from existing runtime pieces. Sigma charging agents as a **premium SKU** is the opening: Rill bundles agent capability in seats.
- **Editable-plan transparency** (S–M): adopt Sigma's best UX idea — show the analyst agent's query plan (metrics view, filters, measures) as an editable/inspectable breakdown before/alongside results. Rill's structured tool calls make this mostly a frontend rendering task on data that already exists in the conversation tree.
- **Accuracy benchmark** (♻) — neither has one; first mover owns the claim.

## 3. Phase 2 — Sufficient self-service (3–6 months)

- **Ad-hoc measures/calcs in explore** (L, ♻omni §3.1) — the keystone against the spreadsheet demo; promotion-through-code-review as the governed counter to metric divergence.
- **Report-layout canvas + PDF/PNG** (♻qs §2.5) — answers Sigma Reports.
- **Chart breadth** (♻ maps/gauge/waterfall/funnel-parity) — Sigma's native set is a checklist standard now.
- **Dashboard actions via MCP tools** (♻qs §3.4) — a slice of Sigma's Actions framework (trigger workflows from data) without the app-builder surface.
- **Embedded self-service** (♻cube §4.2 creator-mode-lite): saved views exist; scoped canvas composition for embed users answers "embedded explorations."

## 4. Phase 3 — Structural & deliberate bets (6–12 months)

- **Composite metrics views** (XL ♻) — Sigma's relationship-based joins are a real modeling advantage today.
- **Writeback decision** (M–L if pursued): Sigma's input tables/Sigma Tables are its boldest bet. Rill options: (a) minimal `form`/input component in canvas writing through a custom API to a connector table (covers annotations, targets, approvals — the 80% case), or (b) concede the app-platform category explicitly. Recommend (a) scoped narrowly + explicit non-goal on the app platform; revisit on deal evidence. (♻mb Actions overlap.)
- **Tenant packaging** (M): productize org/project + attribute tokens into a named "Tenants"-style API for SaaS embedders (provisioning, delegated admin, per-tenant usage) — mostly exists, needs packaging (♻ per-tenant AI caps, cube §2.6).
- **SCIM/audit/localization** (♻).

## 5. GTM

1. **"Rill vs Sigma" page**: all-in TCO, governed-by-construction vs metric divergence, OSS/local vs warehouse-only SaaS, bundled AI vs premium-SKU agents + pass-through costs.
2. **Embedded flank**: SaaS vendors evaluating Sigma embed face per-click warehouse costs at scale — the calculator is the wedge; Persona/G2-class logos are the target profile.
3. **Operational/event-scale**: Sigma on ClickHouse is Beta and limited; Rill's ClickHouse depth + BYO is the counter for the workloads Sigma's warehouse-only model handles worst.
4. **Momentum narrative**: don't fight the $3B story head-on; niche it — "Sigma is Excel for the warehouse; Rill is the operational metrics layer for products and agents."

## 6. Non-goals

Full spreadsheet paradigm/Excel function library, data-app platform (forms/modals/app builder beyond the minimal input slice), warehouse-agent federation (signal-gated), transactional Sigma-Tables-style CRUD, Salesforce-replacement ambitions.

**One-line strategy:** make Sigma's warehouse bill the headline (calculator + benchmark), match its agent packaging from Rill's existing runtime, ship *sufficient* self-service (ad-hoc calcs, reports, actions) so the spreadsheet demo doesn't end evals — and win where its architecture can't follow: fixed-cost sub-second embedded and operational analytics, governed by construction.
