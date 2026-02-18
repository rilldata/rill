---
description: Score and rank a list of features or initiatives using RICE, ICE, or weighted scoring frameworks, with clear rationale for each score
allowed-tools: Read, AskUserQuestion
argument-hint: "<list of features, backlog items, or initiatives to prioritize>"
---

Turn a backlog of features or initiatives into a clearly ranked, defensible prioritization using RICE, ICE, or weighted scoring — with documented rationale for every score.

Input: $ARGUMENTS

## Instructions

### 1. Choose a Scoring Framework

Ask the user (via `AskUserQuestion`) which framework to use, or recommend based on context:

- **RICE** — best when you have data on reach and effort estimates. Formula: `(Reach × Impact × Confidence) / Effort`
- **ICE** — best for fast, gut-check prioritization. Formula: `Impact × Confidence × Ease`
- **Weighted Scoring** — best when specific strategic criteria matter (e.g., strategic alignment, revenue potential, technical debt reduction). Ask the user for their weights.

### 2. Define Scoring Criteria

For RICE:

- **Reach**: How many users will this affect per quarter? (estimated number)
- **Impact**: How much will this move the needle per user? (3 = massive, 2 = significant, 1 = low, 0.5 = minimal, 0.25 = trivial)
- **Confidence**: How confident are you in these estimates? (100% = high, 80% = medium, 50% = low)
- **Effort**: Total person-months to build, test, and ship.

For ICE:

- **Impact**: 1–10 scale. How much will this move the key metric?
- **Confidence**: 1–10 scale. How sure are you it will work?
- **Ease**: 1–10 scale. How easy/fast is this to ship?

For Weighted Scoring, define 4–6 criteria and assign weights that sum to 100%.

### 3. Score Each Item

Create a scoring table for each feature:

- Document the score for each dimension
- Provide a 1-sentence rationale for the score
- Calculate the final score

### 4. Output a Ranked List

Present:

1. A **ranked table** (highest score first) with all scores visible
2. A **recommended top 3** to ship next, with brief justification
3. **Items to defer or cut**, with the reason why

### 5. Flag Risks and Dependencies

For top-ranked items, note:

- Dependencies on other teams or systems
- Technical or business risks that could invalidate the score
- Any items that scored low but have strategic override reasons (e.g., compliance, exec mandate)

## Output Format

Use a markdown table for the scoring matrix. Follow with a short narrative summary of the recommendations. Avoid jargon — the output should be presentable directly to stakeholders.
