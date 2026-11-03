# LATER — scheduled opportunistically

Ids and full detail in [`../../AUDIT_TODO.md`](../../AUDIT_TODO.md) unless noted. These are real issues, but bounded blast radius: schedule when touching the surrounding code, or batch by area.

## 1. Remaining mediums, batched by area

- [ ] **Admin lifecycle/compensation batch:** A-1 (CreateProject partial state), A-2 (`used_flusher` drops IDs), A-6 (ClickHouse DDL rollback), A-11 (orphaned managed repos), A-9 (custom-domain annotation refresh — verified still broken 2026-09-20), A-10 (synchronous GitHub webhook → durable job), A-4/A-5 (k8s provisioner ping + panics).
- [ ] **Controller/reconciler robustness batch:** B-1 (catalog release on init failure — verified still present), B-2 (mutations after shutdown), B-3 (staged-model cleanup on canceled ctx), B-4 (migration template panic), B-6 (non-durable refresh triggers), B-9 (SSE sends not cancellation-aware).
- [ ] **Driver context/cleanup batch:** C-3/C-4 (Athena/Redshift pagination ignores ctx), C-5 (unbounded embedded-CH download), C-8 (stale handles after spec update), C-10 (uncancelable open-time validation).
- [ ] **Shared-utility batch:** D-1 (blob temp-dir cleanup), D-2 (blob `Next` ignores ctx), D-3/D-6 (git helpers without ctx), D-4 (conncache nil `HangingFunc`).
- [ ] **Frontend polish batch:** E-3 (SSE reconnect errors after close), E-4/E-5 (request-queue priority + cancellation), E-6 (OLAP list invalidation on connector change), F-1 (embed RPC disposer), F-3 (alert/report actions lack permission+confirm gates — verified still present), F-4 (duplicate submit), F-7 (read-before-guard), G-1 (canvas `lastVisitedState` keyed by name — verified still present), G-6 (file-read errors as missing content), G-8 (local-file source detection never matches), G-9 (committed secrets after failed import), G-10 (selector subscription leaks), G-11/G-12/G-13 (file-invalidation races, dir deletes, `.db` substring match).

## 2. Verification debt

- [ ] Work through [`../verified/NEEDS-VERIFICATION.md`](../verified/NEEDS-VERIFICATION.md) (32 partial/needs-test items): write the missing regression tests, promote or close each.
- [ ] The `audit/verified/` queue files are frozen at the 2026-07-31 snapshot and superseded by `AUDIT_TODO.md`; either refresh the parts manifests + rerun `build_reports.mjs`, or add a supersession banner and archive the directory.

## 3. Product / PLG ideas

- [ ] Triage [`../verified/IDEAS-BACKLOG.md`](../verified/IDEAS-BACKLOG.md) (84 proposals from `product-ideas.md`, `plg_improvement_ideas.md`, `plg-first-user-experience.md`) against the competitive roadmaps; promote the overlaps (they appear in both = strongest signal).
- [ ] `OPEN_ISSUES_REPORT.md` Phases 2–5 (customer-requested Explore/filter/time features → ClickHouse & connectors → Developer UX quick wins → refactors/big bets).

## 4. Competitive-roadmap engineering items (from `audit/*/roadmap.md`)

Frequency-ranked shared items ([`../COMPETITORS.md`](../COMPETITORS.md)); two already shipped (maps #8931, agent skills #9851 — mark done in the roadmaps when publicized):

- [ ] Customer-facing AI evals + published accuracy benchmark (appears in all 9 audits; `type: eval` Phase 1 already exists uncommitted on branch `ai_evals`)
- [ ] Slack agent (8 audits)
- [ ] Semantic joins / composite metrics views (7)
- [ ] SCIM + audit logs (7)
- [ ] Remaining chart breadth beyond maps: waterfall/gauge/boxplot/histogram (6)
- [ ] Viewer-side ad-hoc calcs — ephemeral measures (#9855/#9864) are the foundation; extend per omni roadmap §3.1 (5)
- [ ] React embed SDK + embeddable chat API (5)
- [ ] Outbound MCP connectors (4; design exists on branch `mcp_connector`)
- [ ] Python/forecasting agent tool (4) · delivery destinations (4) · cloud SQL console (4)
- [ ] Kafka/streaming ingestion (3, decisive vs Tinybird) · report layouts/PDF polish (3) · drill-to-records (3)
