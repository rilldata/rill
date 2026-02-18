---
description: Analyze a batch of customer feedback (support tickets, NPS responses, reviews, feature requests) to extract patterns, prioritize opportunities, and surface actionable insights
allowed-tools: Read, Glob, AskUserQuestion
argument-hint: "<feedback file, paste of feedback items, or description of feedback source>"
---

Analyze 10–500+ customer feedback items from any source — support tickets, NPS verbatims, G2/Capterra reviews, feature requests, user interviews — and extract actionable patterns ranked by frequency and impact.

Input: $ARGUMENTS

## Instructions

### 1. Ingest the Feedback

Accept feedback in any of these forms:
- A file path (use `Read` tool)
- Pasted text in the conversation
- A description of feedback themes (if raw data isn't available)

If the feedback is unstructured, ask via `AskUserQuestion`:
- What is the source? (NPS, support, reviews, interviews, sales calls)
- What time period does this cover?
- What product area or feature set does this relate to?

### 2. Categorize Each Item

For each piece of feedback, identify:
- **Theme**: The underlying need or pain point (not the surface request)
- **Sentiment**: Positive / Negative / Neutral
- **Urgency signal**: Does the user describe churn risk, workaround behavior, or blocking frustration?
- **User segment** (if inferrable): Power user, new user, enterprise, SMB, etc.

### 3. Extract Patterns

Group feedback items by theme and count frequency. Look for:
- **High-frequency pain points**: Themes that appear in 10%+ of items
- **High-severity items**: Individual feedback pieces that signal churn, safety, or legal risk — even if rare
- **Unmet jobs**: Things users are trying to do that the product doesn't support well
- **Unexpected use cases**: Users using the product in ways you didn't anticipate — often signals new market opportunities
- **Positive signals**: What users love — protect these from inadvertent regression in roadmap decisions

### 4. Prioritize Opportunities

Rank the top themes using a simple 2×2:
- **X-axis**: Frequency (how many users mentioned this)
- **Y-axis**: Severity (how much does this hurt the user or the business)

Quadrant labels:
- High frequency + High severity → **Urgent: Fix or build now**
- Low frequency + High severity → **Risk: Monitor and triage**
- High frequency + Low severity → **Polish: Quick wins**
- Low frequency + Low severity → **Backlog: Low priority**

### 5. Surface Insights

Write a synthesis covering:
- **Top 3 opportunities** with supporting evidence (quote 2–3 representative pieces of feedback for each)
- **One thing to stop doing** (if the data suggests a feature or behavior is causing net harm)
- **One underserved segment** that shows up in the data with distinct needs
- **Recommended next step**: What should the PM do with this analysis? (user interviews, prototype test, quick fix, escalate to leadership, etc.)

## Output Format

Lead with a short executive summary (5 sentences max). Follow with the opportunity ranking table. Then provide detailed write-ups for the top 3 opportunities, each with representative quotes. End with recommendations. Preserve customer voice — use direct quotes rather than paraphrasing whenever possible.
