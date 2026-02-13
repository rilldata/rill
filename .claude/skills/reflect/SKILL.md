---
description: Reflect on the conversation and persist learnings to project documentation
allowed-tools: Bash(git:*), Glob, Grep, Read, Edit, Write, AskUserQuestion
argument-hint: "[focus area or specific learning]"
---

You are reflecting on this conversation to extract and persist valuable learnings that will improve future work.

Focus area (if specified): $ARGUMENTS

Follow these steps:

1. **Review the conversation** for:
   - Code patterns or conventions established
   - Architectural decisions made
   - Review feedback patterns (what to look for, what to avoid)
   - Style preferences expressed
   - Common mistakes identified
   - Best practices discovered

2. **Categorize learnings** by where they belong:
   - `CLAUDE.md` — General conventions, project guidance, backend patterns
   - `.claude/rules/frontend.md` — Frontend conventions, Svelte/TypeScript/TanStack Query patterns
   - `.claude/skills/*/SKILL.md` — Improvements to skills
   - `CONTRIBUTING.md` — Contribution guidelines
   - Other relevant documentation

3. **Read the target files** to understand existing content and avoid duplication

4. **Draft proposed changes** and present them to the user:
   - In a text message, show each file to be modified with the exact additions/changes
   - Explain why each learning is valuable
   - Then use `AskUserQuestion` to ask for approval or modifications

5. **Apply approved changes** to the appropriate files

## Guidelines

- Be selective — only persist learnings that have lasting value
- Be concise — distill insights to their essence
- Avoid duplication — check if similar guidance already exists
- Preserve structure — match the style of existing documentation
- Group related learnings — don't scatter related concepts

## Examples of Learnings Worth Persisting

- "In this codebase, prefer X over Y because..."
- "When reviewing PRs, always check for..."
- "This pattern caused issues; avoid it by..."
- "The team prefers this naming convention..."
- "Error handling should follow this pattern..."

## What NOT to Persist

- One-off fixes that don't generalize
- Temporary workarounds
- Context-specific decisions that won't apply elsewhere
- Learnings already covered in existing docs
