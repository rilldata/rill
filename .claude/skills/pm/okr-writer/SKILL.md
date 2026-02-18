---
description: Write and refine Objectives and Key Results (OKRs) using the Measure What Matters methodology, with coaching on common pitfalls
allowed-tools: Read, AskUserQuestion
argument-hint: "<team name, strategic goal, or draft OKRs to refine>"
---

Write strong OKRs using John Doerr's *Measure What Matters* methodology — or critique and improve a draft set of OKRs the user provides.

Input: $ARGUMENTS

## Instructions

### 1. Understand Context

If not provided, ask via `AskUserQuestion`:
- Is this for a company, team, or individual?
- What time period? (quarterly is standard)
- What is the overarching company/product goal this OKR should ladder up to?
- Are these new OKRs or a draft to review?

### 2. OKR Fundamentals (apply throughout)

**Objectives** must be:
- Inspirational and qualitative — describe a destination, not a task
- Ambitious but achievable ("moonshot" vs. "roofshot" — label which)
- Memorable in a single sentence
- NOT a metric (no numbers in the objective)

**Key Results** must be:
- Measurable and time-bound
- Outcome-oriented (measure impact, not output or activity)
- 3–5 per objective
- Scored 0–1.0 at the end of the period (0.7 = success; 1.0 = sandbagging)

### 3. Common Pitfalls to Check Against

Flag any of the following:
- **Task masquerading as KR**: "Launch feature X" is output, not outcome. Rewrite as "Feature X drives 20% increase in Y by [date]."
- **Vanity metric**: Metrics that look good but don't signal real value (e.g., page views without conversion)
- **Too safe**: If hitting 1.0 seems easy, push the number higher
- **No owner**: Each KR should have a clear DRI (directly responsible individual)
- **Too many**: More than 3 objectives or 5 KRs per objective dilutes focus

### 4. Write or Rewrite OKRs

For each Objective:
```
Objective: [Inspiring, qualitative goal]
  KR1: [Metric] from [baseline] to [target] by [date]
  KR2: [Metric] from [baseline] to [target] by [date]
  KR3: [Metric] from [baseline] to [target] by [date]
```

### 5. Health Check

After drafting, evaluate each OKR set:
- Does each KR, if achieved, make the objective undeniably true?
- Are any two KRs measuring the same thing? Consolidate if so.
- Is there a leading indicator KR (early signal) alongside lagging KRs?
- Are the KRs collectively sufficient to achieve the objective, or is something missing?

## Output Format

Present final OKRs in clean markdown. For each KR, add a one-line note explaining why this metric was chosen. If reviewing a draft, show the original vs. the improved version side by side.
