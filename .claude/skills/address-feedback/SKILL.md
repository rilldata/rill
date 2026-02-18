---
description: Review PR feedback and address each item interactively
allowed-tools: Bash(git:*), Bash(gh:*), Glob, Grep, Read, Task, Edit, Write, AskUserQuestion
argument-hint: "<PR number> [reviewer name]"
---

Fetch PR review feedback, then walk through each item interactively so the author can decide how to address it.

Input: $ARGUMENTS

## Instructions

### 1. Fetch Feedback

1. Get PR details using `gh pr view <pr-number>`
2. Fetch review comments using `gh api repos/{owner}/{repo}/pulls/{number}/comments` (inline) and `gh api repos/{owner}/{repo}/pulls/{number}/reviews` (summaries)
3. If a reviewer name was provided, filter to only their comments
4. Read the relevant code for each comment

### 2. Assess

Summarize: what are the key themes, which items are quick wins, and which might warrant pushback?

### 3. Walk Through Each Item

Go through items in order of significance. For each:

1. Quote the feedback and show the relevant code
2. Offer 2-4 approaches via `AskUserQuestion`, including "push back" when appropriate
3. Execute the chosen approach — make the change, or draft a response
4. Move to the next item

### 4. Wrap Up

Summarize changes made, any drafted responses, and suggest a commit message.
