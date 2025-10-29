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

### Code Examples
- Always include working, runnable examples
- Show both success and error cases
- Use realistic variable names (not `foo`, `bar`)
- Include necessary imports
- Verify examples build successfully with `npm run build docs/`
- Prefer short, focused examples over large blocks of code
- Check docs/reference that the example code matches the correct YAML keys

### Language
- Use "you" to address the reader
- Avoid jargon unless domain-specific and necessary
- Spell out acronyms on first use
- Use present tense: "returns" not "will return"
- Keep paragraphs short (1–3 sentences)
- Avoid filler words (“basically,” “in order to,” etc.)

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
- If the PR doesn’t relate to an existing Linear issue, include a short reason such as “Internal cleanup — no Linear ticket”.

