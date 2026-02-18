---
description: Run 7 stakeholder perspectives on any document or decision in parallel — Engineering, Design, Executive, Legal, Customer, Devil's Advocate, and Sales — then synthesize consensus, tensions, blockers, and questions
allowed-tools: Read, AskUserQuestion, Task
argument-hint: "<document, decision, or proposal to review>"
---

Stress-test any document or decision by simulating 7 distinct stakeholder perspectives in parallel, then synthesizing the results into a unified review with consensus, tensions, blockers, and key questions to prepare for.

Input: $ARGUMENTS

## Instructions

### 1. Read and Understand the Input

Read the provided document thoroughly. If it's a file path, use the `Read` tool. Identify:

- What type of document is this? (PRD, proposal, strategy doc, roadmap, design brief, etc.)
- What decision or approval is being sought?
- What is the timeline for feedback?

### 2. Run 7 Parallel Stakeholder Reviews

Use the `Task` tool to spawn 7 parallel agents simultaneously (in a single message), each reviewing the document from a distinct perspective:

**Agent 1 — Engineering Lead**
Prompt: Review this document as a senior engineering lead. Identify: technical feasibility concerns, missing technical requirements, hidden complexity, system dependencies, and questions you'd ask before starting work. Be specific and practical.

**Agent 2 — Product Designer / UX**
Prompt: Review this document as a product designer. Identify: user experience gaps, missing user research or validation, design assumptions that need testing, accessibility considerations, and any flows that will be confusing to users.

**Agent 3 — Executive / Business**
Prompt: Review this document as a C-suite executive. Identify: strategic alignment, ROI clarity, resource implications, risks to company goals, and what's missing that you'd need to approve this.

**Agent 4 — Legal / Compliance**
Prompt: Review this document as a legal and compliance reviewer. Identify: privacy concerns (data collection, GDPR/CCPA), terms of service implications, regulatory risks, IP considerations, and liability exposure.

**Agent 5 — Customer Voice**
Prompt: Review this document as an advocate for the end customer. Identify: assumptions about user needs that aren't validated, potential for confusion or frustration, missing use cases, and whether the proposed solution actually solves the stated problem.

**Agent 6 — Devil's Advocate**
Prompt: Review this document as a skeptic whose job is to find every flaw. What's the strongest case against doing this? What could go wrong? What assumptions are most likely to be wrong? What would you need to be convinced this is worth doing?

**Agent 7 — Sales / Revenue**
Prompt: Review this document as a sales leader. Identify: how this affects current deals, the sales narrative and messaging implications, pricing or packaging concerns, competitive positioning, and what you'd need to explain this to a customer.

### 3. Synthesize Results

After all agents complete, organize findings into:

**Consensus** — Points raised by 3+ perspectives as significant concerns or needs.

**Tensions** — Points where two perspectives are in conflict (e.g., Engineering says "keep it simple" while Sales says "we need more configuration options").

**Hard Blockers** — Issues that must be resolved before this can move forward (typically Legal, Engineering feasibility, or missing strategic alignment).

**Open Questions to Prepare For** — The most important questions you'll be asked in your next review meeting, ranked by likelihood. Include a suggested answer for each.

**Suggested Revisions** — Top 3 changes to the document that would address the most significant feedback.

## Output Format

Present the synthesis clearly with sections for Consensus, Tensions, Hard Blockers, and Open Questions. Lead with the most critical finding. Keep the stakeholder-by-stakeholder breakdown in an appendix section at the end for reference.
