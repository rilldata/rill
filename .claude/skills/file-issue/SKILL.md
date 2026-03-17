---
description: Create a Linear issue from the current conversation context
allowed-tools: mcp__linear__create_issue, mcp__linear__list_issue_labels, mcp__linear__list_teams, mcp__linear__list_issue_statuses, AskUserQuestion, Read, Grep, Glob
argument-hint: "[description of the issue]"
---

Create a Linear issue based on the current conversation context.

Input: $ARGUMENTS (optional — if not provided, infer from conversation history)

## Instructions

1. Review the conversation and `$ARGUMENTS` to understand the issue. Search the codebase for relevant details if needed.
2. Draft the issue: title, description, team, priority (default Normal), and labels. For team, default to "Application" for frontend/UI, "Platform" for backend/runtime, "Infra" for infrastructure.
3. Preview the draft via `AskUserQuestion` with options to create or edit first.
4. Create using `mcp__linear__create_issue` and report the issue identifier and URL.
