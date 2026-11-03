# Roadmap: Competing with (and beating) ThoughtSpot

Companion to `comparison.md`. Sizes S/M/L/XL; shared items ♻cube/♻omni/♻qs/♻looker/♻hex.

## 1. Strategic frame

ThoughtSpot owns the "AI analyst" narrative at the enterprise tier (Gartner Leader, Spotter everywhere) but is exposed on the three axes buyers now test: **accuracy** (54.67% public benchmark, no self-published evidence), **cost** (credits, $92.5K median, services), and **architecture** (live-query latency/cost that SpotCache tacitly concedes). Rill's pitch writes itself: *the accuracy comes from the architecture* — one governed metrics layer grounding both dashboards and agents, on an engine that's fast at fixed cost.

Where Rill actually meets ThoughtSpot: AI-BI evaluations (mid-market and AI-forward enterprise) and embedded deals (where TS's $12K/mo floor prices out exactly the SaaS vendors Rill serves at $500/mo).

### Deal-loss categories
1. "ThoughtSpot is a Gartner Leader; who's Rill?" (brand/analyst)
2. "Spotter does search, Slack, Salesforce, forecasting." (surface breadth)
3. "We need SCIM/audit/enterprise IT." (♻)

## 2. Phase 1 — Win on evidence (0–3 months)

- **Publish the accuracy benchmark** (M, ♻cube §2.3 / ♻hex): run the runQL-style methodology on Rill's metrics-layer chat, publish the number and the harness. TS is uniquely vulnerable — its bad number is public, its rebuttal is vocabulary ("consistency") not data. This single asset reframes every TS bake-off.
- **Verified answers + coaching parity** (S–M, ♻cube §2.2): certified queries = verified answers; `ai_instructions` UI = coaching. Match the *vocabulary* in docs so eval scorecards check the boxes.
- **Slack agent** (M, ♻) — Spotter-in-Slack is the demo that wins TS deals.
- **Proactive insight scans** (M): schedule the anomaly "Explain" machinery — periodic scans of key measures producing ranked insight digests (SpotIQ's job, agent-quality output). Mostly composition: alerts + Explain + AI reports.

## 3. Phase 2 — Embedded flank (3–6 months)

TS prices out mid-market embedded; Rill should own it:
- **Headless chat/agent API** (M, ♻cube §4.2): TS's `SpotterAgentEmbed` (bring-your-own-UI agent) is the reference; Rill's `CompleteStreaming` productization matches it.
- **React embed SDK** (S–M, ♻) + agent white-labeling (S, ♻).
- **Embedded pricing calculator** (S, GTM): TS $12K/mo floor + per-view usage vs Rill flat — publish it.
- **Forecasting agent tool** (M, ♻qs §3.3): removes "Spotter forecasts, Rill doesn't."

## 4. Phase 3 — Enterprise credibility (6–12 months)

- **SCIM + audit + custom roles** (♻) — table stakes at TS's tier.
- **Composite metrics views** (XL ♻) — TS models join; required for enterprise schemas.
- **Analyst relations** (ongoing, ♻): Gartner Agentic Analytics Market Guide first (winnable), ABI MQ long-term. TS's "only independent vendor among Leaders" line is available to Rill too — with "and open source" appended.
- **Salesforce/ServiceNow surface** — evaluate on deal signal only (non-goal until then).

## 5. GTM

1. **"Rill vs ThoughtSpot" page**: accuracy-by-architecture (with the benchmark), fixed-cost sub-second vs live-query + SpotCache patchwork, transparent seats vs credits + services, OSS/local vs Cloud-only.
2. **Target the credit-bill pain**: content on consumption-pricing surprise; case-study angle "same workload, predictable bill."
3. **StartupSpot counter**: Rill's OSS + $20 seats + flat embed *is* the startup offer — say so where StartupSpot markets.
4. **Search-UX concession**: don't chase tokenized search; position chat + Explore as the modern equivalent.

## 6. Non-goals

Tokenized search UX, Salesforce/ServiceNow embedded surfaces (signal-gated), Mode-style notebook suite (♻hex stance), on-prem software line, credit-pricing models.

**One-line strategy:** beat ThoughtSpot with published evidence (accuracy benchmark + verified answers), take the embedded mid-market its pricing abandons, and let the architecture story — grounding by construction on a fast fixed-cost engine — do the work its own SpotCache concession invites.
