---
description: Generate a complete Product Requirements Document (PRD) from a problem statement or feature idea, using JTBD analysis, opportunity trees, and sprint-ready user stories
allowed-tools: Read, Glob, Grep, AskUserQuestion, WebSearch
argument-hint: "<problem statement, feature idea, or opportunity>"
---

Transform a problem statement or feature idea into an engineering-ready PRD using Jobs-to-be-Done analysis, opportunity mapping, and sprint-ready user stories.

Input: $ARGUMENTS

## Instructions

### 1. Clarify Scope (if input is vague)

If the input lacks sufficient detail, use `AskUserQuestion` to ask:
- Who is the primary user/customer affected?
- What outcome are we trying to achieve (not the solution)?
- What constraints exist (timeline, platform, team size)?

### 2. JTBD Analysis

Frame the problem using the Jobs-to-be-Done format:

> **When** [situation], **I want to** [motivation/job], **so I can** [expected outcome].

Identify:
- **Functional job**: The practical task the user is trying to accomplish
- **Emotional job**: How they want to feel
- **Social job**: How they want to be perceived

### 3. Opportunity Tree

Map the problem space before jumping to solutions:

```
Goal: [Desired outcome]
  └── Opportunity: [Why users can't achieve the goal today]
        ├── Assumption: [What must be true for this opportunity to be real]
        └── Solution Space: [2–3 possible approaches — do NOT commit to one yet]
```

### 4. PRD Structure

Write the full PRD with the following sections:

**Overview**
- Problem statement (1–2 sentences)
- Why now (business context, urgency)
- Success metrics (primary KPI + 1–2 supporting metrics)

**Users & Segments**
- Primary user persona
- Secondary users (if any)
- Out of scope users

**Requirements**
- P0 (must have for launch)
- P1 (strongly preferred)
- P2 (nice to have, post-launch)

For each requirement, state: what the system must do, not how.

**Non-Requirements**
- Explicitly call out what this PRD does NOT cover to prevent scope creep.

**Open Questions**
- List any unresolved decisions that need stakeholder input before engineering starts.

### 5. Sprint-Ready User Stories

Convert P0 requirements into user stories using the format:

> **As a** [user type], **I want to** [action], **so that** [benefit].

For each story, include:
- **Acceptance criteria** (3–5 bullet points using Given/When/Then)
- **Story points estimate** (1, 2, 3, 5, or 8 — Fibonacci scale)
- **Dependencies** (other stories or systems this relies on)

## Output Format

Use clear markdown with headers. Keep the PRD scannable — PMs and engineers should be able to grasp the full scope in under 5 minutes. Avoid padding; every sentence should add information.
