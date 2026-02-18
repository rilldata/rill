---
description: Build a comprehensive product metrics framework using AARRR (pirate metrics) or input/output methodology, and identify and validate your product's North Star metric
allowed-tools: Read, AskUserQuestion
argument-hint: "<product, feature, or business area to build a metrics framework for>"
---

Design a complete metrics framework for a product or feature — identifying the North Star metric, building out the full AARRR funnel or input/output model, and defining leading and lagging indicators at each stage.

Input: $ARGUMENTS

## Instructions

### 1. Understand the Product Context

If not provided, ask via `AskUserQuestion`:

- What does the product do and who uses it?
- What is the primary business model? (subscription, usage-based, marketplace, freemium, etc.)
- What stage is the product at? (early/PMF search, growth, scale, mature)
- Is there an existing set of metrics in use? If so, what are they?

### 2. North Star Metric

Identify the North Star metric: the single number that best captures the core value the product delivers to customers. A good North Star:

- Reflects customer value (not just revenue or vanity activity)
- Is leading, not lagging (predicts future business health)
- Is actionable — the team can affect it directly
- Is understandable — a new employee can explain it

Present 2–3 candidate North Star metrics and recommend one with reasoning.

**Common examples by business type**:

- SaaS productivity tool → Weekly active users completing core workflow
- Marketplace → Successful transactions per month
- Consumer app → DAU/MAU ratio (stickiness)
- Data platform → Queries run per active user per week

Validate the North Star by testing: "If this metric goes up while everything else stays the same, is the business actually healthier?"

### 3. AARRR Funnel Metrics

Map the full customer lifecycle with 2–3 metrics per stage:

**Acquisition** — How do users discover and arrive?

- Sources: Organic search, paid, referral, direct, partner
- Key metrics: New signups, CAC by channel, trial starts

**Activation** — Do users experience core value quickly?

- Define the "aha moment" — the action that correlates most strongly with retention
- Key metrics: Activation rate (% reaching aha moment), time-to-value

**Retention** — Do users keep coming back?

- Define retention for this product (daily, weekly, monthly depending on use case)
- Key metrics: D7/D30/D90 retention, churn rate, resurrection rate

**Referral** — Do users tell others?

- Key metrics: NPS, viral coefficient (K-factor), referral rate, word-of-mouth signups

**Revenue** — Are we monetizing effectively?

- Key metrics: MRR/ARR, ARPU, LTV, LTV:CAC ratio, expansion revenue

### 4. Input/Output Metric Pairing

For each AARRR stage, pair:

- **Output metric**: The outcome you're trying to achieve (lagging, harder to move quickly)
- **Input metric**: The leading indicator or lever the team can pull to affect the output

Example:

> Output: D30 retention = 40% | Input: % of new users who complete onboarding within 24 hours

### 5. Instrumentation Checklist

For each key metric, document:

- What event or data point needs to be tracked?
- Where does this data live today? (product analytics, data warehouse, CRM, etc.)
- Is this metric currently being measured? If not, what's needed to instrument it?
- What's the current baseline value (if known)?

### 6. Anti-metrics

Name 2–3 metrics to explicitly NOT optimize for — things that could improve without the product actually getting better:

- Vanity metrics (e.g., total registered users without activity filter)
- Metrics that can be gamed easily
- Metrics that could be optimized at the expense of user trust

## Output Format

Present the North Star recommendation with rationale, then a metrics table organized by AARRR stage with input/output pairs. Follow with the instrumentation checklist. Use a clean table format throughout. Keep the narrative tight — this should be a working document, not a strategy deck.
