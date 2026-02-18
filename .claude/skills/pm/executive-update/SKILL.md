---
description: Create a crisp executive update or status report using the SCQA (Situation, Complication, Question, Answer) framework
allowed-tools: Read, AskUserQuestion
argument-hint: "<topic, project, or situation to report on>"
---

Transform messy context into a crisp, structured executive update using the SCQA framework (Situation, Complication, Question, Answer) — the gold standard for executive communication.

Input: $ARGUMENTS

## Instructions

### 1. Gather Context

If not provided, ask via `AskUserQuestion`:
- What is the audience? (CEO, board, cross-functional leadership, skip-level)
- What format is expected? (email, Slack update, slide, verbal brief)
- What decision or action, if any, do you need from the audience?
- What has changed since the last update?

### 2. SCQA Structure

Write the update using the four-part SCQA framework:

**Situation** (1–2 sentences)
Set shared context. State what's true right now that the audience already knows or would agree with. This is NOT the problem — it's the stable backdrop.
> Example: "We're 6 weeks into Q2, and our pipeline conversion rate has been a focus area since the Q1 review."

**Complication** (1–3 sentences)
Introduce the tension. What has changed, gone wrong, or become relevant that makes the situation require attention? This creates the "why are you telling me this" moment.
> Example: "Conversion rate improved from 18% to 22%, but deal velocity slowed — average close time increased from 28 to 41 days, which puts us at risk of missing the Q2 revenue target."

**Question** (1 sentence, implicit or explicit)
The natural question the reader is now asking. Often not stated explicitly, but naming it sharpens the answer.
> Example: "What's driving the slowdown and what should we do about it?"

**Answer** (the bulk of the update)
Lead with the recommendation or key insight, then support it with evidence. Structure as:
- **Bottom line up front**: Your conclusion or recommendation in one sentence
- **Supporting data**: 2–3 data points or observations that back the bottom line
- **Next steps**: What will happen next and by when, with owners named

### 3. Calibrate for Audience

Adjust the update based on audience:
- **CEO / Board**: Lead with business impact and decisions needed. Skip operational detail.
- **Cross-functional leaders**: Emphasize dependencies and what you need from them.
- **Skip-level**: Include enough operational context to be credible, but still lead with the headline.

### 4. Editing Pass

Review the draft against these criteria:
- Can the reader understand the key message in the first 30 seconds?
- Is every sentence earning its place — does it add information or can it be cut?
- Are there any weasel words or passive constructions hiding uncertainty? If there's uncertainty, name it directly.
- Is the ask (if any) unambiguous?

## Output Format

Present the final SCQA update in clean prose — no sub-bullets within the Situation or Complication. Reserve bullets for the Answer section's supporting evidence and next steps. Aim for under 250 words for async written formats; include a "TL;DR" one-liner at the top for long updates.
