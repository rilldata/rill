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
```

### Scope Discipline
- **Stay within the original PR scope**: Only modify files and sections that were part of the original changes
- **Minimal edits for review feedback**: When addressing review comments, make targeted changes to the specific lines or paragraphs mentioned
- **Avoid scope creep**: Do not refactor, reorganize, or rewrite entire files unless explicitly requested
- **Check the PR diff first**: Before making changes, review what files were modified in the original PR. These are your boundaries.

### Addressing Review Comments
1. **Identify the specific issue**: What exact line, section, or example needs to change?
2. **Make surgical edits**: Change only what's necessary to address the comment
3. **Preserve existing content**: Don't rewrite sections that weren't mentioned in the review
4. **Ask for clarification**: If a review comment seems to require broader changes, ask if they want to expand the PR scope

### Examples of Good vs Bad Responses

❌ **Bad** - Review comment: "Remove the advanced authentication example from project-configuration.md"
- Response: Rewrites the entire project-configuration.md file, reorganizes sections, updates all examples

✅ **Good** - Same review comment
- Response: Removes only the specific advanced authentication example, leaves rest of file unchanged

❌ **Bad** - Review comment: "Fix typo in line 45 of ai-chat.md"  
- Response: Fixes typo but also reformats the entire document, changes heading structure, adds new sections

✅ **Good** - Same review comment
- Response: Fixes only the typo on line 45

### Red Flags That Indicate Scope Creep
- Modifying files that weren't in the original PR
- Rewriting sections that weren't mentioned in review comments
- Adding new examples or content when only asked to remove/fix something
- Changing heading structures across entire documents
- Reformatting or reorganizing content that works fine

### When Broader Changes ARE Appropriate
- Reviewer explicitly asks to "rewrite this section"
- Review comment reveals systemic issues across the file (get confirmation first)
- Changes are required by the build validation (broken links, etc.)
- Maintainer explicitly expands the scope in their comment

---

<!-- Added from PR review - 2025-01-xx: New page creation and content consolidation guidelines -->
## 📄 Creating New Documentation Pages

### When NOT to Create New Pages

**Critical: Default to editing existing pages rather than creating new ones**

- ❌ **Don't create new pages for existing topics**: If a topic already has documentation, update that page
  - WRONG: Creating `docs/connect/olap/bigquery.md` when BigQuery is already documented elsewhere
  - RIGHT: Update the existing BigQuery documentation page
  
- ❌ **Don't create category/index pages without approval**: Pages like `olap.md` or `data-source.md` that serve as category overviews
  - These affect site navigation and information architecture
  - Discuss with maintainers first before creating

- ❌ **Don't duplicate content across multiple locations**: If content exists in one place, don't create a second page with similar information
  - Example: Don't create both `connect/olap/snowflake.md` AND `connect/data-source/snowflake.md`
  - Consolidate into one canonical location

### Before Creating a New Page - Checklist

**Always verify these before creating a new documentation page:**

- [ ] Search existing docs to confirm this topic isn't already covered
- [ ] Check if the content should be added to an existing page instead
- [ ] Verify the new page fits into the existing site structure/navigation
- [ ] Confirm there's no duplication with other documentation
- [ ] Propose the new page in the PR description or Linear issue BEFORE creating it
- [ ] Get maintainer approval if it affects navigation or creates a new documentation section

### When New Pages ARE Appropriate

- ✅ Documenting a genuinely new feature that has no existing documentation
- ✅ Creating pages explicitly requested by maintainers
- ✅ Adding pages that are part of an approved documentation restructure
- ✅ Tutorial or guide pages that don't duplicate reference documentation

### Handling Page Removal Requests

**If a reviewer asks to remove entire pages:**

- Remove the page file completely (don't just empty it)
- Check for any navigation references (sidebars, tables of contents) that link to the removed page
- Update or remove those references to prevent broken links
- Consider if the content should be merged into another existing page instead of fully deleted

**Example of proper page removal:**
```bash
# Remove the file
rm docs/docs/connect/olap/bigquery.md

# Check for references
grep -r "bigquery.md" docs/

# Update any navigation configs, sidebars, or index pages
