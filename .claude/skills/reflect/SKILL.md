---
description: Reflect on the conversation and persist learnings to project documentation
allowed-tools: Bash(git:*), Glob, Grep, Read, Edit, Write, AskUserQuestion
argument-hint: "[focus area or specific learning]"
---

Extract learnings from this conversation and persist them to the appropriate project files.

Focus area: $ARGUMENTS

## Instructions

1. Review the conversation for patterns, conventions, or decisions worth persisting
2. Categorize each learning by where it belongs:
   - `CLAUDE.md` — General conventions, project guidance, backend patterns
   - `.claude/rules/*.md` — Scoped rules (frontend, docs, code review, etc.)
   - `.claude/skills/*/SKILL.md` — Improvements to skills
3. Draft proposed changes and present them via `AskUserQuestion` for approval before applying
