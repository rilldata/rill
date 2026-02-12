---
title: "Token-Based Pricing: Technical Draft"
description: System changes required to implement token/compute-based pricing for Rill Cloud
sidebar_label: "Token-Based Pricing (Draft)"
sidebar_position: 01
---

:::note
This is a **technical draft** outlining system changes required to implement token-based pricing. It is not yet live.
:::

## Background

Rill is evolving toward a **Live Connect–first architecture**, where customers query their own OLAP engines (ClickHouse, Snowflake, BigQuery, etc.). In this model, customers already pay their OLAP vendor for compute, storage, and query scan costs. Pricing based on data size (GB ingested or stored) leads to double-charging.

The new pricing axis is **application usage** — how much Rill is actually used — represented by a single abstraction:

> **Tokens** — derived from bytes scanned and returned per query.

Tokens also unify pricing for **AI features**, which incur both query execution and LLM inference costs.

---

## Plan Structure

### Developer Plan (Free, Rill-Managed)

Zero-friction entry point for individuals or small teams evaluating Rill.

| Detail | Value |
| --- | --- |
| Compute & storage | Rill-managed |
| Token allowance | 1,000 tokens/month |
| Dashboards | Core dashboards and Explore |
| Support | Community or email |

**Hard limits** on total data size, query scope, and complexity.

**Token top-ups** (no plan upgrade required):
- +1,000 tokens → $10
- +5,000 tokens → $50

**Not included:** Live Connect, Slack/24×7 support, enterprise security.

---

### Teams Plan (Live Connect)

Cost-covering PLG tier for small teams running production workloads on their own OLAP.

| Detail | Value |
| --- | --- |
| Price | $249/month (no annual commitment) |
| Token allowance | 10,000 tokens/month |
| Features | Dashboards, Explore, embeds, exports, API access |
| Alerts | Usage alerts at 50 %, 80 %, 100 % |

**Overages** — tokens are a hard monthly quota. Token packs:
- +5,000 tokens → $50
- +10,000 tokens → $100

When the quota is exceeded, queries and AI are paused until the admin tops up.

---

### Enterprise Plan

Flat annual contract (e.g. $75 k+/year).

| Detail | Value |
| --- | --- |
| Token allowance | Very high or unlimited |
| Users / dashboards / projects | Unlimited |
| Support | Dedicated Slack channel + SLA |
| Security | SSO, audit logs, SOC 2 |
| AI | BYO AI models (no AI token consumption) |

Optional on-prem deployment.

---

## Token System

### Token Definition

A **token** represents normalized **query impact**, derived from:

- **Bytes scanned** by the OLAP engine
- **Bytes returned** to the client

Initial heuristic (subject to tuning):

> ~250 MB scanned ≈ 1 token

This abstraction smooths differences across OLAP engines while remaining explainable.

### Token Consumers

| Action | Token Logic |
| --- | --- |
| Dashboard view | Bytes-based |
| Explore interaction | Bytes-based |
| Embedded dashboards | Bytes × multiplier (e.g. 3×) |
| Export | Bytes + fixed multiplier |
| API calls | Fixed tokens per call |
| AI chat | Bytes + AI multiplier |

### AI Token Model

AI chat incurs two costs: OLAP query execution and LLM inference.

- **Rill-managed AI:** Higher multiplier (e.g. 5×)
- **BYO AI (Enterprise):** Lower multiplier (e.g. 1×)

```
AI Chat Tokens = Query Tokens + AI Multiplier
```

---

## System Changes Required

### 1. Query Metering Service

**What:** A new backend service that captures bytes-scanned and bytes-returned for every query dispatched through Rill, converts them to tokens, and writes metering events.

**Key work items:**

- Instrument every OLAP query path (ClickHouse, MotherDuck, StarRocks Pinot, Druid, DuckDB) to extract `bytes_scanned` and `bytes_returned` from query metadata / result stats.
- Define a `MeteringEvent` proto with fields: `org_id`, `project_id`, `user_id`, `action_type`, `bytes_scanned`, `bytes_returned`, `tokens_consumed`, `timestamp`, `query_id`, `cache_hit`.
- Build a metering writer that batches events and flushes to a durable store (e.g. ClickHouse table or time-series store).
- **Cache-hit handling:** When `cache_hit = true`, log the event for observability but set `tokens_consumed = 0`. Cached results are free.
- **Local AI metering:** The Rill Developer local runtime must emit metering events to the cloud service when the user is authenticated and invokes AI features. Standard local queries (no AI) are not metered.
- Apply the token conversion formula (`bytes_scanned / 250 MB`, rounded up, then multiplied by the action multiplier).

**Connector-specific considerations:**

| Connector | Bytes-scanned source |
| --- | --- |
| ClickHouse | `system.query_log.read_bytes` |
| DuckDB (local) | Profiling output or `EXPLAIN ANALYZE` |

---

### 2. Token Ledger & Quotas

**What:** A per-org ledger that tracks token balance, enforces quotas, and handles top-ups.

**Key work items:**

- New database table: `token_ledger` (`org_id`, `project_id` (nullable), `period_start`, `period_end`, `tokens_allocated`, `tokens_consumed`, `tokens_purchased`).
- When `project_id` is set, the row represents a per-project budget within an Enterprise org. Per-project allocations must sum to ≤ the org-level allocation.
- Atomic increment on every metering event flush.
- Skip increment when the query response indicates a **cache hit** (no tokens consumed for cached results).
- Quota check middleware: before executing a query, compare `tokens_consumed` against `tokens_allocated + tokens_purchased`. If exceeded, reject the query with a clear error. For orgs with per-project budgets, check the project-level ledger first.
- Monthly reset job that rolls the ledger forward on billing-cycle boundaries.

**Proto additions:**

```protobuf
message TokenUsage {
  string org_id = 1;
  int64 tokens_allocated = 2;
  int64 tokens_consumed = 3;
  int64 tokens_purchased = 4;
  google.protobuf.Timestamp period_start = 5;
  google.protobuf.Timestamp period_end = 6;
  string project_id = 7; // optional — set for per-project budgets
}

message GetTokenUsageRequest {
  string org_id = 1;
}

message GetTokenUsageResponse {
  TokenUsage usage = 1;
}
```

---

### 3. Billing Integration (Stripe)

**What:** Extend the existing Stripe integration to support token-based plans and top-up purchases.

**Key work items:**

- Create new Stripe Products/Prices for each plan tier (Developer free, Teams $249/mo).
- Create Stripe Products for token packs ($10/1 k, $50/5 k, $100/10 k).
- On successful top-up checkout, credit `tokens_purchased` in the ledger.
- Update the existing billing CLI commands (`rill billing subscription`) to reflect token-based plans.
- Wire plan tier to the correct `tokens_allocated` value on subscription create/update.

---

### 4. Action Multipliers & Configuration

**What:** A configuration layer that maps action types to their token multipliers, allowing tuning without code changes.

**Key work items:**

- Define an `ActionMultiplier` config (YAML or database-backed):
  ```yaml
  multipliers:
    dashboard_view: 1.0
    explore_interaction: 1.0
    embedded_dashboard: 3.0
    export: 2.0
    api_call_fixed_tokens: 0.5
    ai_chat_rill_managed: 5.0
    ai_chat_byo: 1.0
  ```
- The metering service reads these multipliers at token-computation time.
- Admin API to view current multipliers (Enterprise may negotiate custom rates).

---

### 5. Pre-Query Cost Estimation

**What:** Before executing a query, estimate its token cost and warn the user if it exceeds a threshold.

**Key work items:**

- For each connector, implement a lightweight estimation path:
  - **ClickHouse:** `EXPLAIN ESTIMATE` or table-level byte stats.
  - **Snowflake:** Dry-run / `EXPLAIN` with estimated scan size.
  - **BigQuery:** `dryRun: true` job option returns `totalBytesProcessed`.
  - **DuckDB:** `EXPLAIN ANALYZE` on a sampled subset.
- Surface the estimate in the UI before execution.
- New API endpoint:
  ```protobuf
  message EstimateQueryCostRequest {
    string project_id = 1;
    string sql = 2;
  }

  message EstimateQueryCostResponse {
    int64 estimated_bytes_scanned = 1;
    double estimated_tokens = 2;
    bool exceeds_warning_threshold = 3;
  }
  ```

---

### 6. Admin Dashboard & Usage APIs

**What:** Provide Teams and Enterprise admins with visibility into token consumption.

**Key work items:**

- New API endpoints:
  - `GetTokenUsageByUser` — breakdown per user in the org.
  - `GetTokenUsageByDashboard` — breakdown per dashboard / action type.
  - `GetTokenUsageTimeSeries` — historical trend data for charting.
- Frontend admin page showing:
  - Current period usage vs. quota (progress bar).
  - Per-user and per-dashboard token tables.
  - Historical usage chart.
- Automated alert emails/notifications at configurable thresholds (default: 50 %, 80 %, 100 %).

---

### 7. Quota Enforcement & Grace Mechanism

**What:** Hard-stop query execution when tokens are exhausted, with a one-time grace policy.

**Key work items:**

- Quota check interceptor in the query execution pipeline.
- When quota is hit:
  - Return a structured error (`RESOURCE_EXHAUSTED`) with a message directing the admin to purchase more tokens.
  - UI displays a modal with token-pack purchase options.
  - AI features are disabled.
- Grace mechanism:
  - Track `grace_used` boolean per org per billing period.
  - On first overage spike (e.g. ≤ 10 % above quota), allow queries to continue for up to 1 hour, then enforce.
  - Reset `grace_used` each billing cycle.

---

### 8. Frontend UX Changes

**What:** Surface token usage, warnings, and purchase flows throughout the UI.

**Key work items:**

| Area | Change |
| --- | --- |
| Query execution | Show estimated token cost; warning modal if cost exceeds threshold |
| Dashboard header | Token usage badge (e.g. "3,200 / 10,000 tokens") |
| Explore | Per-interaction token cost shown after each query |
| Admin settings | Token usage page with per-user/dashboard breakdowns |
| Billing page | Token pack purchase buttons; current plan details |
| Alerts | In-app banners at 50 %, 80 %, 100 % usage thresholds |
| Quota exceeded | Full-screen modal: "Token quota reached. Purchase more or wait for next cycle." |

---

### 9. Data Model & Migration

**New tables:**

| Table | Purpose |
| --- | --- |
| `metering_events` | Raw per-query metering events (includes `cache_hit` flag) |
| `token_ledger` | Per-org (and optionally per-project) period-level token accounting |
| `token_purchases` | Top-up purchase records (linked to Stripe) |
| `action_multipliers` | Configurable per-action token multipliers |
| `usage_alerts` | Alert threshold configs and delivery log |

**Migration considerations:**

- Existing Trial and Team plans must be mapped to the new Developer and Teams tiers.
- Existing Enterprise contracts remain unchanged until renewal.
- Historical query data (if available) can be backfilled to seed usage baselines.

---

### 10. Rollout Strategy

| Phase | Scope | Goal |
| --- | --- | --- |
| **1 — Metering (shadow mode)** | Deploy metering service; log events but don't enforce quotas | Validate token calculations; tune the 250 MB heuristic |
| **2 — Admin visibility** | Ship usage dashboards and alerts to admins | Build confidence; gather feedback |
| **3 — New plans** | Launch Developer and Teams plans with token quotas for new sign-ups | Begin monetization |
| **4 — Migration** | Migrate existing Trial/Team customers to new plans with grace period | Complete transition |
| **5 — Tuning** | Adjust multipliers, grace policy, and estimation accuracy based on production data | Optimize |

---

## Open Questions

1. **Heuristic tuning** — Is 250 MB/token the right starting point across all connectors? Should it vary per engine?
2. **Embedded multiplier** — Should embedded dashboard queries charge the embedding customer's org or the end-user's org?
3. **Rate limiting vs. token limiting** — Should there be a per-second query rate limit in addition to monthly token quotas?

### Resolved Decisions

3. **Caching credit** — Cached query results do **not** consume tokens. The metering service must check the query response metadata for a cache-hit flag and skip token accounting when present.
4. **Sub-org billing** — Enterprise customers can allocate token budgets **per-project**. The `token_ledger` table must support an optional `project_id` scope, and the admin dashboard must allow setting per-project quotas that sum up to the org-level allocation.
5. **Offline/local usage** — Rill Developer (local DuckDB) queries are metered **only when the user is logged in and using AI features**. Standard local queries are free. The local runtime must detect an authenticated session and AI invocation before emitting metering events to the cloud ledger.
