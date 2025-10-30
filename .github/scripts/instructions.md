## 🚨 CRITICAL: Processing Review Comments on Documentation PRs

### File Deletions

When a review comment asks to delete/remove a file:

1. **You MUST explicitly indicate file deletion** in your response
2. Use the `DELETE_FILE:` marker that the script can parse
3. Format: `DELETE_FILE: path/to/file.md` (one per line)
4. Do NOT just mention deletion in the summary - the script needs an actionable deletion instruction
5. Do NOT output file content with ```file: tags if you're deleting it

**Examples:**

❌ **Wrong** - Review comment: "Remove bigquery.md and snowflake.md"
```
📝 Summary of Changes:
- docs/docs/connect/olap/bigquery.md: File will be deleted (not outputting content)
- docs/docs/connect/olap/snowflake.md: File will be deleted (not outputting content)

[then outputs other files]
```

✅ **Correct** - Same review comment:
```
📝 Summary of Changes:
- docs/docs/connect/olap/bigquery.md: Deleted
- docs/docs/connect/olap/snowflake.md: Deleted

DELETE_FILE: docs/docs/connect/olap/bigquery.md
DELETE_FILE: docs/docs/connect/olap/snowflake.md

[then outputs other files that need updates]
