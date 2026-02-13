---
description: Checkout PR Branch
allowed-tools: Bash(git:*), Bash(gh:*), AskUserQuestion
argument-hint: "[PR number]"
---

Switch to a PR's branch and fetch the latest updates.

## Arguments

- `$ARGUMENTS`: PR number (optional - if not provided, infer from conversation history)

## Instructions

1. Determine the PR number:
   - If `$ARGUMENTS` is provided, use that
   - Otherwise, look at the conversation history to find the PR that has been discussed (e.g., from `/review` or `/re-review` commands, or PR URLs/numbers mentioned)
   - If no PR can be inferred, ask the user which PR to checkout

2. Get the branch name for the PR using `gh pr view <pr-number> --json headRefName -q '.headRefName'`
3. Fetch the branch: `git fetch origin <branch-name>`
4. Checkout the branch: `git checkout <branch-name>`
5. Pull the latest changes: `git pull origin <branch-name>`
6. Report the current status with `git log --oneline -5` to show recent commits
