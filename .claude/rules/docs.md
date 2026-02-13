---
paths: docs/docs/**/*
---

## Documentation Principles

When writing or reviewing documentation in `docs/`:

1. **Sidebar labels follow resource naming** — Items in the Build section should use the resource type name (e.g., "Alerts" not "Code Alerts").

2. **Add connective prose** — Don't just present code snippets. Add 1-2 sentences of explanatory context before each section explaining what it does and when to use it.

3. **Link to related concepts** — When referencing concepts documented elsewhere (Custom APIs, Metrics Views, etc.), link to their documentation pages.

4. **Explain use cases** — For configuration options, explain when and why you'd use each option, not just what it does.

5. **Promote important concepts** — If a concept only appears buried in an example, consider giving it its own dedicated section or subsection.

6. **Navigation labels prefer single words** — Top navigation labels should be concise (e.g., "Developers" not "Developer Docs", "Guide" not "User Guide"). URL paths should match labels and be short (`/developers/`, `/guide/`). When splitting documentation for different audiences, base the split on user activity (code-based vs UI-based) rather than product names, which may change over time.
