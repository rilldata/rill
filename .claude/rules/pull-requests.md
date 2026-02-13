## Branch Naming

Prefix branch names with your username, followed by a short description:

```
username/fix-share-button-visibility
username/add-activation-metric
```

The Linear ticket and authorship information belong in the PR, not the branch name. Branch names are ephemeral — they get deleted after merge.

## PR Titles

Use backticks around code identifiers (function names, variable names, file names) in PR titles.

## Creating Pull Requests

1. Use the repo's PR template (`.github/PULL_REQUEST_TEMPLATE.md`). Replace the `INSERT DESCRIPTION HERE` placeholder with the description, but keep the checklist intact — the author will check the boxes themselves.
2. Link to the relevant Linear issue or Slack message after the description but before the checklist. Use the format `Closes [APP-123](url)` for Linear issues.
3. Keep the description concise — a few bullet points, not paragraphs. No "Summary" or "Overview" headers.
4. End the PR body (after the checklist) with an attribution footer in exactly this format:
   - A horizontal rule (`---`)
   - Followed by italicized text: `*Developed in collaboration with Claude Code*`
