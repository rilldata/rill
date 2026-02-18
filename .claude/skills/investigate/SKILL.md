---
description: Investigate bugs by spawning parallel agents to explore multiple hypotheses
allowed-tools: Bash(git:*), Bash(gh:*), Glob, Grep, Read, Task, AskUserQuestion
argument-hint: "<problem description, Slack link, or Linear issue>"
---

Investigate a bug by gathering context, then spawning parallel Task agents to test multiple hypotheses simultaneously.

Input: $ARGUMENTS

## Instructions

### 1. Gather Context

Parse the input for links and fetch context:
- **Slack link** (contains `slack.com/archives`): Fetch the message and thread replies via Slack MCP tools
- **Linear issue** (ID like `ENG-1234` or URL): Fetch via `mcp__linear__get_issue`, including comments and attachments
- **Free text**: Use directly as the problem statement

User-provided hunches should be prioritized when forming hypotheses.

### 2. Investigate in Parallel

Generate 3-5 hypotheses, then use the **Task tool** to spawn one agent per hypothesis **in a single message** (parallel, not sequential). For each agent:
- `subagent_type: "Explore"`, `model: "sonnet"`
- A focused prompt with the hypothesis, what to look for, and what would confirm or refute it

### 3. Synthesize

After all agents complete, report: which hypotheses were confirmed, refuted, or inconclusive, and recommend next steps.
