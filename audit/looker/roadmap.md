# Roadmap: Competing with (and beating) Looker

Companion to `comparison.md`. Effort sizes: S/M/L/XL. Shared items marked ♻cube/♻omni/♻qs.

## 1. Strategic frame

Looker is two opportunities in one:

1. **A migration source.** The post-acquisition wave is real and already monetized by Omni (BuzzFeed in <3 months) and Lightdash. Looker customers leave over: slow dashboards, LookML burden, platform fees ($60K–200K/yr), support decay, and strategy churn (Modeler shelved, Studio renamed twice, customer-hosted abandoned). Rill is philosophically the *closest* destination (git-native semantic BI) but isn't on the shortlist — that's a marketing failure, not a product one.
2. **A rival with Google gravity.** Where Looker is actively bought (GCP-committed shops, BigQuery estates), it wins on procurement + Gartner Leader status + free-until-Oct-2026 Gemini CA. Rill's counters mirror the Quick playbook: marketplace listing, speed proof, DX proof.

**Positioning: "Everything you wanted Looker to be — sub-second, local-first, agent-native, and a tenth of the price."** Rill is the modernized expression of Looker's own founding idea (governed semantic models in git) without the warehouse-latency ceiling or the LookML tax.

### Deal-loss categories to eliminate
1. "Rill can't model our LookML join graph." (♻ semantic joins)
2. "We need CA-API-style embeddable AI / paginated reports / delivery destinations." (checklist)
3. "It's in our GCP commit." (procurement)
4. "Rill isn't on the migration shortlist." (GTM absence)

## 2. Phase 1 — The migration play (0–3 months) — highest ROI

- **`rill migrate lookml` (M)** ♻cube-GTM: LookML → Rill converter driven by the developer agent. Views/explores → models + metrics views; dimensions/measures/filters mechanical; joins flagged with upstream-model workarounds until composite metrics views land; PDTs → incremental models (often a strict upgrade); Liquid user-attribute filters → security policies. Ship as a skill + CLI command; publish a conversion guide with a real project.
- **Migration marketing (S)**: "Looker migration" landing page + honest capability-mapping docs (what maps, what doesn't) + a benchmark ("same model, Looker vs Rill: dashboard latency + annual cost"). Target the exact search terms Omni/Lightdash own today.
- **PDT-pain counter-story (S)**: content showing incremental models + OLAP replacing PDT/aggregate-awareness engineering.
- **Delivery destinations (M)** ♻omni §2.4: S3/GCS/webhook/SFTP + PNG — Looker's Action Hub is the standard ex-Looker users expect; this is the most-cited practical gap in migrations.

## 3. Phase 2 — Neutralize Google's AI story (3–6 months)

- **Chat/Conversational API (M)** ♻cube §4.2: productize `CompleteStreaming` as an embeddable chat API — Looker's CA API GA is now the reference point in embedded-AI evals.
- **Verified/golden queries (S–M)** ♻cube §2.2: Looker shipped these in Preview; Rill's certified-queries item covers it — match the vocabulary ("verified answers") in docs.
- **Evals + published accuracy (M)** ♻cube §2.3: Google claims "errors cut by two-thirds"; Rill needs its own number. Same harness, Looker-comparable methodology.
- **Agentic monitoring (M)**: Looker's Agentic Workflows preview (agents that watch metrics and notify) maps to Rill alerts + AI reports — package "AI-monitored metrics" (alert triggers analyst-agent root-cause, delivers narrative) as a named feature. Mostly composition of existing pieces (alerts → `resolver: ai` → Slack once the Slack agent ♻ exists).
- **Python/advanced analytics agent tool (M)** ♻qs §3.3: answers Code Interpreter.
- **BYO-LLM marketing (S)**: Looker is Gemini-only with overage billing starting Oct 1, 2026 — "your model, your key, your cost control" lands precisely then. Time content to that date.

## 4. Phase 3 — Structural (6–12 months)

- **Composite metrics views / joins (XL)** ♻ — required to fully replace LookML-modeled estates; the migration tool feeds requirements (real-world join graphs to validate against).
- **Postgres-wire Metrics SQL API (L)** ♻cube §3.1 — replaces Looker's Open SQL Interface / Connected Sheets role; with a Sheets add-on (♻cube §3.3) it covers the "governed metrics in spreadsheets" job Looker sells to GCP shops.
- **GCP Marketplace listing (S–M, BD)**: mirror of the AWS move (♻qs §2.1) — burn GCP committed spend; co-exist with BigQuery estates via warehouse-direct + BYO ClickHouse.
- **Enterprise pack** ♻: SCIM, audit logs, custom roles — Looker's System Activity/IAM maturity is the bar in enterprise evals.

## 5. GTM

1. **"Rill vs Looker" page** + migration guide + latency/cost benchmark. The migration-shortlist absence is the cheapest fixable loss.
2. **Court the LookML developer community** (the people who *like* code-first BI): content comparing LookML→Rill YAML side-by-side; presence where Omni/Lightdash recruit refugees.
3. **Time the Oct 1, 2026 AI-overage moment**: "Looker's AI meter starts today" campaign — Rill's BYO LLM + flat token quotas.
4. **Partner-position in BigQuery shops**: warehouse-direct OLAP mode + BigQuery connector; don't fight the platform, fight the BI layer.
5. **Point at strategy churn** (Modeler shelved, Looker reports killed, Data Studio renames, customer-hosted abandoned) for multi-year-standardization risk — same stability argument as vs Quick.

## 6. Non-goals

Slides generation, Looker Studio's 800-connector free-tier fight, marketplace/blocks ecosystem parity, mobile apps, LookML-compatibility runtime (convert, don't emulate).

**One-line strategy:** get on the Looker-migration shortlist with a converter and honest content (Phase 1), match the three Gemini features buyers actually shortlist on — embeddable chat API, verified answers, monitoring agents (Phase 2) — and let joins + SQL API + GCP Marketplace finish the enterprise story (Phase 3), all while selling the thing Looker can't fix: warehouse-bound latency.
