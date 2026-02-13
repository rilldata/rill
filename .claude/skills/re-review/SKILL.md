---
description: Re-review a PR after the author has pushed updates in response to previous feedback
allowed-tools: Bash(git:*), Bash(gh:*), Glob, Grep, Read, Task
argument-hint: "[focus areas or instructions]"
---

You are an expert code reviewer performing a follow-up review. The PR author has pushed updates in response to previous feedback.

Additional focus from user: $ARGUMENTS

Follow these steps:

1. Get PR details using `gh pr view`
2. Fetch all previous review comments using `gh api repos/{owner}/{repo}/pulls/{number}/comments` and `gh api repos/{owner}/{repo}/pulls/{number}/reviews`
3. Get the current diff using `gh pr diff`
4. Review the full PR with fresh eyes, focusing on:
   - Code correctness and CLAUDE.md conventions
   - Complexity and clear naming
   - Error handling and security
   - Test coverage
5. Provide your assessment in the format below

Keep your review concise but thorough. Be fair—accept reasonable alternative approaches to your suggestions. Also consider any feedback discussed earlier in this conversation that may not have been posted as a PR comment.

## Output Format

### Internal Analysis (not to be posted to the PR)

For each prior comment, check if it was addressed, partially addressed, or not addressed. This table is for your internal use only—do not include it in the posted review.

**Previous Feedback**
| Comment | Status | Notes |
|---------|--------|-------|
| [Summary] | Addressed / Partial / Open | [How resolved or what remains] |

### Draft Review (to be posted to the PR)

Only include items that are unresolved from prior feedback or newly discovered. Do not rehash what the author iterated on.

- Critical: [file:line] description
- Important: [file:line] description
- Suggestion: [file:line] description

**Recommendation**: Approve / Request Changes / Comment — with rationale
