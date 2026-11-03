# Roadmap: Competing with (and beating) Metabase

Companion to `comparison.md`. Sizes S/M/L/XL; shared items ♻cube/♻omni/♻qs/♻looker/♻hex/♻ts.

## 1. Strategic frame

Metabase isn't beaten head-on at the low end — free, familiar, five-minute setup is unassailable for "just give me dashboards over Postgres." The winnable moments are **graduations**: (1) *scale graduation* — dashboards slow down as data grows (Metabase's #1 complaint); (2) *embedded graduation* — authenticated embed users at $12/user/mo and live-query concurrency limits; (3) *governance graduation* — OSS users hitting the sandboxing/SSO paywall; (4) *AI graduation* — Metabot answers questions, but can't do governed AI reports, anomaly root-cause, or agent-built projects.

Also note the strategic signal: **even Metabase** now ships an MCP server, Slack agent, BYO-model AI, AI governance, and git sync — confirming these are table stakes, not differentiators. Rill's Phase-1 AI-ops items (♻) are validated and overdue.

**Positioning: "Where Metabase graduates to."** Not a Metabase killer — the next step when data, embeds, or governance get serious. (Luzmo already frames Rill as the ships-an-interface code-first middle path.)

## 2. Phase 1 — Own the graduation path (0–3 months)

- **Metabase importer** (M): serialization/Remote Sync exports are documented YAML — convert questions/dashboards/models to Rill models + metrics views + canvases via the developer agent (same pattern as the QuickSight/LookML importers ♻). "Bring your Metabase, keep your dashboards, gain the engine."
- **"Outgrowing Metabase" content** (S, GTM): the graduation story with a benchmark — same dashboard, 100M+ rows, Metabase-on-Postgres vs Rill — plus an embedded-pricing calculator ($12/authenticated user vs $5 effective).
- **OSS positioning** (S): Rill OSS includes RLS, SSO-ready policies, the engine, and AI — exactly what Metabase paywalls. Make the comparison table explicit on the site.
- **Chart breadth** (♻ maps/waterfall/etc., omni §2.1) — Metabase now exceeds Rill's native chart count; this gap is no longer just an Omni/Quick issue.

## 3. Phase 2 — Close the DX conveniences (3–6 months)

- **Cloud SQL console** (M, ♻omni §3.5) — Metabase's SQL editor is its most-loved surface; Rill's absence is felt precisely by Metabase graduates.
- **React embed SDK** (S–M, ♻) — Metabase's SDK is the reference at this price point.
- **Guest-embed ergonomics** (S): position/polish magic-token public URLs as the "guest embed" equivalent (docs, per-URL caps, usage stats).
- **Slack agent + AI admin** (♻ cube Phase 1) — even-Metabase-has-it items.
- **Quick-start over raw tables** (M, consider): `GenerateMetricsViewFile` already scaffolds from a table — package a "connect DB → auto-generate metrics views for chosen tables → explore in minutes" onboarding flow to blunt Metabase's setup-speed advantage without abandoning the modeled-metrics philosophy.

## 4. Phase 3 — Keep the moats compounding (6–12 months)

Composite metrics views (♻), evals/benchmark (♻), writeback evaluation (Metabase Actions + Sigma input tables both have it; assess demand for a minimal `actions:`/form-input story in canvas — M if pursued), SCIM/audit (♻).

## 5. GTM

1. **"Rill vs Metabase" page** anchored on the four graduations; honest about where Metabase is the right choice (small data, internal, free).
2. **Importer-led campaigns** in communities where "Metabase is slow" threads recur (r/dataengineering, HN).
3. **Embedded calculator** for SaaS vendors comparing authenticated-user costs.
4. **Don't out-cheap them** — no free-cloud-tier race; the OSS product is the free tier.

## 6. Non-goals

GUI query builder over arbitrary raw tables (philosophy conflict; quick-start flow covers the onboarding job), Documents, x-rays clone, 20-driver live-query breadth, price-floor matching.

**One-line strategy:** be the graduation path — importer + benchmark + pricing calculator to catch teams outgrowing Metabase on scale, embeds, governance, or AI — while closing the two conveniences its users will miss (SQL console, embed SDK) and conceding the low end gracefully.
