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

## Documentation Style Guide

### Tone
- **Professional but approachable**: Clear and precise, but not overly formal
- **User-focused**: Explain WHY before HOW
- **Active voice**: "Use this method to..." not "This method can be used to..."
- **Concise**: Respect the reader's time
- **Progressive clarity**: Start with a simple example or use case, then gradually introduce more complex concepts or configurations.  
  - Lead with the easiest-to-understand path.
  - Build toward advanced or nuanced details.  
  - This helps readers gain confidence before diving into complexity.

### Structure
- Dont remove inline comments.
- Start with a brief description / Overview (1–2 sentences)
- Include a practical code example early
- Explain parameters/options in a table
- End with common use cases or gotchas
- ⚙️ **Complexity order**:
  - Present examples and scenarios in **increasing complexity**.
  - Example order: simple → intermediate → advanced.
  - Keep sections consistent so readers can easily follow escalation of difficulty.
- ⚠️ **Header consistency**:
  - When changing or renaming section headers (`#`, `##`, `###`), update all in-page links or references that use that header.
  - Update anchor-style markdown links (`[See this section](#old-header-name)` → `[See this section](#new-header-name)`).
  - Verify Table of Contents and navigation remain correct.
  - Avoid using `####`.
- 🧩 **Inline edits only**:
  - Update the actual markdown files in `docs/`.
  - Do **not** generate separate summary or new markdown pages.
  - If a new doc seems required, propose it in the PR description instead of creating it.
- 🚫 **No deprecated examples**:
  - Do not use outdated patterns like `type: source`.
  - Replace them with modern and correct usage examples, check the runtime/parser/* to understand the actual usage of each component.
  - Follow best practices and dont put raw text in the examples for connectors and instead reference the `.env` 
    - IE `google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"`

<!-- Added from PR review - 2025-01-xx: Document structure organization -->
### Document Organization

**Appendix sections:**
- Move detailed technical reference material to an "Appendix" section at the end
- Consider moving to appendix:
  - Alternative authentication methods (e.g., service account JSON after HMAC keys)
  - Legacy or less common configuration patterns
  - Advanced troubleshooting details
- Keep main body focused on the primary/recommended approach
- Structure: Introduction → Main content → Common use cases → Appendix
- **Appendix header formatting**: Use consistent title format across all appendix sections
  - Format: `### How to [action using specific tool/interface]`
  - Example: `### How to create a service account using the Google Cloud Console`
  - Maintain parallel structure for all appendix headings in a document

### Terminology Standards (Updated from Review Comments)

**Critical terminology changes to always apply:**

- ✅ **Use "model"** instead of "source" (sources are deprecated)
  - WRONG: `type: source`, `sources/my_data.yaml`, "create a source file"
  - RIGHT: `type: model`, `models/my_data.yaml`, "create a model file"
  - Apply this to: file paths, YAML properties, explanatory text, and all documentation

- ✅ **Use "connector"** for authentication/connection configuration
  - Connectors go in `connectors/` directory
  - Models reference connectors via the `connector:` property

- ✅ **File structure** for data sources:
  - `connectors/[name].yaml` - Contains authentication credentials
  - `models/[name].yaml` or `models/[name].sql` - Contains data model configuration

<!-- Added from PR review - 2025-01-xx: YAML type and driver specifications -->
- ✅ **YAML `type` and `driver` fields**:
  - Connector files: `type: connector` and `driver: <service-name>`
    - Example: `type: connector` with `driver: gcs`
  - Model files: `type: model` and `driver: <service-name>` or `connector: <connector-name>`
    - Example: `type: model` with `driver: gcs` OR `connector: duckdb` for cloud storage
  - Always verify the correct `type` and `driver` combination for each file type
  
**When processing review comments:**
- If a comment mentions "source is deprecated", update ALL occurrences throughout the file
- Check file paths, code examples, and explanatory text
- Update both YAML examples AND the prose that describes them

<!-- Added from PR review - 2025-01-xx: Cloud storage connector and property corrections -->
### Cloud Storage Specific Standards

**For cloud storage data sources (GCS, S3, Azure, etc.):**

- ✅ **Connector type**: Always use `connector: duckdb` in model files (not the storage service name)
  - Cloud storage models use DuckDB's native capabilities to read from cloud storage
  - WRONG: `connector: gcs` or `connector: s3` in model files
  - RIGHT: `connector: duckdb` in model files
  - Note: The connector file itself uses `driver: gcs` or `driver: s3`

- ✅ **Authentication property names** (case-sensitive):
  - GCS with HMAC: `key_id` and `secret` (not `access_key_id` or `secret_access_key`)
  - S3: `access_key_id` and `secret_access_key`
  - Always verify property names against `runtime/parser/*` connector definitions

- ✅ **SQL vs path property**:
  - Use `sql:` property with DuckDB table functions (this is the correct, non-deprecated way)
  - WRONG: `path: gs://bucket/file.parquet` (deprecated)
  - RIGHT: `sql: SELECT * FROM read_parquet('gs://bucket/file.parquet')`
  - The `sql:` approach is required, not optional
  - Never mark `sql:` usage as "optional" in comments or documentation

- ✅ **Environment variable naming**:
  - Format: `connector.<connector-name>.<property>`
  - Example: `connector.gcs.key_id=<value>`
  - Example: `connector.gcs.secret=<value>`
  - Not: `gcs_key_id` or other variations

<!-- Added from PR review - 2025-01-xx: Authentication requirements and public access -->
- ✅ **Authentication optionality**:
  - Some cloud storage services support public bucket access
  - When authentication is optional (e.g., for public buckets), clearly indicate this
  - Example: "Authentication (or skip for public buckets)"
  - Don't imply authentication is always required when public access is possible

### Code Examples
- Always include working, runnable examples
- Show both success and error cases
- Use realistic variable names (not `foo`, `bar`)
- Include necessary imports
- Verify examples build successfully with `npm run build docs/`
- Prefer short, focused examples over large blocks of code
- Check docs/reference that the example code matches the correct YAML keys

<!-- Added from PR #8166 review - 2025-01-xx: Data source YAML completeness -->
#### Data Source YAML Examples
When documenting data sources (GCS, S3, Azure, etc.), ensure YAML examples include:
- ✅ **Triggers** - Data sources typically need refresh triggers (e.g., `refresh: cron: "0 */6 * * *"`)
- ✅ **Connector reference** - Must reference a connector file via `connector:` property
- ✅ **Complete working configuration** - Not just authentication, but the full model setup
- ❌ Don't show partial configs that won't work in practice
- Verify against `runtime/parser/*` for correct property names and structure

Example checklist for data source docs:
- [ ] Connector YAML example includes all required authentication fields
- [ ] Model YAML example includes `type: model`, `connector:`, and appropriate trigger
- [ ] Both examples use `.env` references for sensitive values
- [ ] File paths use `connectors/` and `models/` directories (not `sources/`)

### Language
- Use "you" to address the reader
- Avoid jargon unless domain-specific and necessary
- Spell out acronyms on first use
- Use present tense: "returns" not "will return"
- Keep paragraphs short (1–3 sentences)
- Avoid filler words ("basically," "in order to," etc.")

<!-- Added from PR review - 2025-01-xx: Deployment documentation standards -->
### Deployment Instructions

**When documenting deployment and environment configuration:**

- ✅ **Keep it simple**: Use `rill env configure` without additional arguments
  - The CLI will walk through all required connectors interactively
  - WRONG: `rill env configure connector.gcs.google_application_credentials`
  - RIGHT: `rill env configure`

- ❌ **Don't use these commands in deploy docs**:
  - `rill env set` - Not part of standard deployment workflow
  - `rill env push` - Not part of standard deployment workflow
  - `rill env pull` - Not part of standard deployment workflow

- ✅ **Preserve existing deployment sections**:
  - If a page has a working deployment section, don't remove or oversimplify it
  - When updating, enhance rather than replace unless the content is incorrect

- ❌ **Avoid unnecessary checklists**:
  - Don't add step-by-step checklists for simple deployment processes
  - The `rill env configure` and `rill deploy` workflow is straightforward enough without checkboxes
  - Save checklists for genuinely complex multi-step processes

### Build Validation
- After editing docs, **run `npm run build docs/`**.
- Fix any errors or warnings, including:
  - Broken links or anchors
  - Duplicate heading IDs
  - Invalid frontmatter or MDX syntax
  - Unresolved imports or missing code blocks
- If warnings persist after autofix, note them in the PR description.
- The build must complete successfully with **no broken links or critical errors** before merging.

---

## Creating Documentation PRs

### Branch Naming
- Format: `docs/pr-{original-pr-number}-{brief-description}`
- Example: `docs/pr-1234-api-auth-endpoints`

### PR Title Format
- `"[DOCS] [Brief description]"`
- Example: `"[DOCS] Add documentation for new authentication endpoints"`

### PR Description Template

### Linking Linear Issues

- Every documentation PR should reference a Linear issue.
    - Format in the PR description:
        **Linear:** [ABC-123](https://linear.app/rilldata/issue/ABC-123)

- The automation workflow will also detect the key (ABC-123) in the branch name or title and insert the correct link automatically.
- Keep the Linear key in your branch name (docs/ABC-123-update-auth-docs) so the workflow can link it even if you forget to edit the body.
- If the PR doesn't relate to an existing Linear issue, include a short reason such as "Internal cleanup — no Linear ticket".

---

<!-- Added from review comments - 2025-01-xx: Workflow file separation guidance -->
## 🔧 Workflow and Automation Changes

### Scope Separation
- **Keep workflow changes in separate PRs** - Don't mix automation/workflow changes with documentation content changes
- If you need to modify `.github/workflows/` files:
  - Create a separate PR focused only on workflow changes
  - Link the workflow PR to the documentation PR if they're related
  - This allows independent review of automation logic vs. content
  
### Consolidation with Existing Workflows
- Before creating new workflow files, check if similar functionality exists
- Propose consolidation with existing workflows (e.g., `claude.yml`) when appropriate
- Discuss in PR description if a new workflow is truly needed vs. extending an existing one

---

<!-- Added from PR review - 2025-01-xx: Content preservation guidelines -->
## 📝 Content Revision Guidelines

### Stay Focused on PR Purpose

**When iterating on changes based on review comments:**

- ✅ **Maintain original scope**: Review comments should refine the original PR changes, not trigger a complete rewrite
  - Before making changes, understand what the original PR was trying to accomplish
  - Only modify content that's directly related to the review comments
  - Don't expand into unrelated sections or topics
  
- ❌ **Avoid scope creep**: Don't use review comments as an excuse to revise the entire document
  - WRONG: A comment about fixing terminology in one section leads to restructuring the entire page
  - RIGHT: Fix the terminology issue mentioned, leave other sections alone unless they have the same issue
  
- ✅ **Targeted fixes**: Apply review feedback surgically
  - If a reviewer comments on authentication examples, fix authentication examples
  - Don't also rewrite the overview, reorganize sections, or change unrelated code blocks
  - Stay within the boundaries of what's being reviewed

**Red flags that indicate scope creep:**
- You're modifying sections that weren't mentioned in review comments
- You're adding new content beyond what was requested
- You're restructuring the document when reviews only asked for small fixes
- The diff is much larger than necessary to address the specific comments

**Best practice:**
- Read all review comments first to understand the scope
- Make a mental checklist of exactly what needs to change
- After making changes, verify you haven't touched unrelated content
- When in doubt, make the minimal change that addresses the feedback

### Preserving Quality Content

**When revising documentation:**

- ✅ **Evaluate before replacing**: If existing content is clear and accurate, enhance it rather than rewriting
- ✅ **Overview sections**: Keep overview text that effectively explains the service/feature
  - Don't replace good overview content with generic descriptions
  - If the original overview is better, restore it
  - Review comments like "replace with old version" indicate the original was superior
- ✅ **Critical sections**: Never remove important sections that were previously present
  - Always check what content existed before your changes
  - If a section was removed accidentally, restore it with a comment explaining why it's important
  - Pay special attention to sections reviewers mark as "!important"

**Red flags that indicate you may be removing valuable content:**
- Reviewer asks to "bring back" or "return" sections
- Reviewer says "this is missing !important"
- Simplifying deployment sections that had nuanced, correct information
- Removing worked examples or CLI command sequences that were accurate
- Comments to "replace overview with old version"

**Best practice:**
- Before making major structural changes, understand why the current structure exists
- When in doubt, add to existing content rather than replacing it
- Mark alternative or advanced patterns appropriately with inline comments, don't remove them
- If replacing overview or introduction text, verify your version is actually clearer and more accurate

<!-- Added from PR review - 2025-01-xx: Overview content positioning -->
### Overview Section Placement

**Critical: Overview content must appear at the top of documentation pages**

- ✅ **Overview position**: The overview/introduction explaining what the service is should be the **first content** after the page title
  - WRONG: Starting with configuration steps or authentication details
  - RIGHT: Service description → Authentication → Configuration
  - Review comments like "return this to the top of the page" indicate overview was moved incorrectly

- ✅ **Overview content quality**: 
  - If reviewer says "replace this with the current 'overview' it was better", use the original text
  - Don't replace specific, accurate service descriptions with generic text
  - Keep technical details that help users understand what the service does

**Example structure:**
```markdown
# Service Name

[Overview paragraph explaining what the service is and its key features]

## Authentication

[Authentication methods and setup]

## Configuration

[Configuration details]
```
