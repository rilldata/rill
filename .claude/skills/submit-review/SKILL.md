---
description: Submit a code review to the PR based on the conversation so far
allowed-tools: Bash(git:*), Bash(gh:*), Read, AskUserQuestion
argument-hint: "[additional instructions]"
---

You are preparing to post a formal code review to a GitHub PR based on the discussion in this conversation.

Additional instructions from user: $ARGUMENTS

Follow these steps:

1. Get PR details using `gh pr view --json number,url,headRefName,baseRefName`
2. Review this conversation to gather all feedback, concerns, and suggestions discussed
3. Organize the feedback into:
   - **Inline comments**: Specific feedback tied to file paths and line numbers (use the current diff to get accurate line numbers)
   - **Overall summary**: High-level observations and recommendation
4. Determine review type: `APPROVE`, `REQUEST_CHANGES`, or `COMMENT`
5. Draft the complete review and present it to the user using `AskUserQuestion`:
   - Show each inline comment with its file:line location
   - Show the overall summary
   - Show the review type (approve/request changes/comment)
   - Ask for approval or modifications
6. Once approved, post using:
   - For inline comments: `gh api repos/{owner}/{repo}/pulls/{number}/reviews` with the comments array
   - The review body should end with the attribution footer (see below)

## Review Format Guidelines

- Be constructive and specific
- Reference file paths and line numbers for inline comments
- Keep the overall summary concise (2-4 sentences)
- Group related feedback together

## Attribution Footer

All reviews must end with:
```
---
*Developed in collaboration with Claude Code*
```

## Posting Reviews via GitHub API

To post a review with inline comments, pipe JSON to `gh api --input -`:
```bash
cat <<'EOF' | gh api repos/{owner}/{repo}/pulls/{number}/reviews --input -
{
  "event": "REQUEST_CHANGES",
  "body": "Overall summary here",
  "comments": [
    {"path": "file.ts", "line": 42, "body": "Comment text"}
  ]
}
EOF
```

Note: Do not use `-f 'comments=[...]'` — gh treats `-f` values as strings, not JSON arrays.

For approval without inline comments:
```bash
gh pr review --approve --body "Summary here"
```

## Important

- Never post without explicit user approval via AskUserQuestion
- Show the complete draft before posting
- Allow the user to request modifications before posting
- Use accurate line numbers from the current diff
