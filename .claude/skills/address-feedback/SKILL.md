---
description: Review PR feedback and address each item interactively
allowed-tools: Bash(git:*), Bash(gh:*), Glob, Grep, Read, Task, Edit, Write, AskUserQuestion
argument-hint: "<PR number> [reviewer name]"
---

You are helping the PR author work through reviewer feedback systematically. Your job is to fetch the feedback, provide a holistic assessment, then guide the author through addressing each item one by one.

PR number: $ARGUMENTS

## Instructions

### Phase 1: Fetch and Understand Feedback

1. Get PR details using `gh pr view <pr-number>`
2. Fetch all review comments using:
   - `gh api repos/{owner}/{repo}/pulls/{number}/comments` (inline comments)
   - `gh api repos/{owner}/{repo}/pulls/{number}/reviews` (review summaries)
3. If a reviewer name was provided, filter to only their comments
4. For each comment, read the relevant code to understand the context

### Phase 2: Holistic Assessment

Provide a brief summary:
- **Overall tone**: What's the reviewer's general sentiment?
- **Key themes**: What patterns emerge across the feedback?
- **Quick wins**: Which items are straightforward to address?
- **Discussion items**: Which items may need clarification or pushback?

### Phase 3: Interactive Review

Go through each feedback item one by one. For each item:

1. **Present the feedback**: Quote the reviewer's comment and show the relevant code
2. **Provide context**: Explain what the reviewer is asking for and why
3. **Suggest options**: Offer 2-4 concrete approaches (including "push back with explanation" when appropriate)
4. **Ask for decision**: Use `AskUserQuestion` to let the author choose how to proceed
5. **Execute**: Based on the author's choice:
   - If they want to make a change, implement it
   - If they want to push back, draft a response for their review
   - If they want to discuss further, explore the issue together
6. **Confirm**: Show what was done and move to the next item

### Guidelines

- Present items in order of significance (blocking issues first)
- Group related comments when it makes sense
- Be honest about trade-offs—don't assume the reviewer is always right
- When the author disagrees with feedback, help them articulate why clearly
- Keep the pace moving—don't over-explain obvious items
- Track which items have been addressed for a final summary

### Output at End

After all items are addressed, provide:
- Summary of changes made
- Summary of responses to draft (if any)
- Remaining items that need further discussion
- Suggested commit message for the changes
