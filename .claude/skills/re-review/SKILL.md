---
description: Re-review a PR after the author has pushed updates in response to previous feedback
allowed-tools: Bash(git:*), Bash(gh:*), Glob, Grep, Read, Task, AskUserQuestion
argument-hint: "[PR number or focus areas]"
---

Re-review a PR that has been updated since a previous review. Unlike a fresh review, focus on whether prior feedback was addressed and flag anything new.

Additional instructions: $ARGUMENTS

## Instructions

1. Get PR details using `gh pr view` and fetch all previous review comments using `gh api repos/{owner}/{repo}/pulls/{number}/comments` and `gh api repos/{owner}/{repo}/pulls/{number}/reviews`
2. For each prior comment, determine if it was addressed, partially addressed, or still open
3. Review the current diff, focusing on changes since the last review
4. Present findings in two parts:
   - **Internal analysis** (not for posting): table of prior comments and their resolution status
   - **Draft review** (for posting): only unresolved items from prior feedback and newly discovered issues. Do not rehash what the author already addressed.
