---
description: Find hardcoded user-facing strings in the branch's changes and migrate them to i18n (paraglide messages), following the repo's i18n conventions
allowed-tools: Bash(git:*), Bash(npm:*), Glob, Grep, Read, Edit, AskUserQuestion
argument-hint: "[optional: files/dirs to narrow focus]"
---

Audit the changes on the current branch for hardcoded user-facing strings that should be moved to i18n (Paraglide/inlang messages), following the conventions in `web-common/src/lib/i18n/README.md`.

Optional focus: $ARGUMENTS

## Instructions

### 1. Determine the diff to audit

- If `$ARGUMENTS` lists files or directories, scope the audit to those paths only.
- Otherwise, audit the PR diff: `git diff --merge-base main` to get changed lines across the branch.
- Only consider **added or modified** lines. Do not flag pre-existing strings unless the file was explicitly passed in `$ARGUMENTS`.
- Only frontend files are in scope: `web-common/`, `web-admin/`, `web-local/`, with extensions `.svelte` and `.ts`/`.svelte.ts`.

### 2. Find hardcoded user-facing strings

Mirror the heuristics in `scripts/i18n-guard.js`:

- **Svelte visible text**: content between `>` and `<` that contains real words.
- **Human-facing attributes**: `placeholder`, `title`, `aria-label`, `alt`, `label`.
- **TypeScript**: string literals used as user-facing copy — e.g. `label:`, `header:`, `body:`, toast/notification text, thrown error messages shown to users, option arrays.

**Skip** (not user-facing): identifiers, URLs, paths, `CONSTANT_CASE`, class/style names, object keys, `console.*`/log messages, code comments, test files (`*.spec.ts`, `*.test.ts`, e2e), and lines with an `i18n-ignore` comment on or above them.

### 3. Propose keys, reusing what exists

For each finding, before proposing a new key, grep `web-common/src/lib/i18n/messages/en.json` for the exact copy — **reuse an existing key** if one matches (especially `common_*`).

When a new key is needed, follow the README conventions:

- Naming: `feature_component_purpose`, lower snake_case, grouped by prefix.
- `common_*` for copy reused across features; otherwise prefix with the feature directory name.
- **Generic shared components** (in `web-common/src/components/`) should use `common_*` keys, not feature-specific ones.
- **Interpolation**: named placeholders, never string concatenation — `"Sort by {label}"`, not `"Sort by " + label`.
- **Pluralization**: use Paraglide variants, not `count === 1 ? ... : ...`.

### 4. Watch for these patterns

- **Module-level `const` in `.ts`** (e.g. an options array with `label`): calling `m.key()` at module load freezes the locale. Use a getter — `get label() { return m.key(); }` — so it resolves lazily and stays locale-reactive.
- **Composed labels**: parameterize the whole string as one interpolated message rather than concatenating a translated prefix with a value.

### 5. Report

Present a table ordered by significance: `file:line`, the current string, the suggested key (or existing key to reuse), and whether interpolation/variants are needed. Note any strings that are borderline or intentionally left as-is.

### 6. Offer to migrate

Ask whether to apply the migration. If yes, for each string:

1. Add the key to **both** `messages/en.json` and `messages/es.json`, placed alphabetically within its prefix group. Provide a real Spanish translation for `es.json`.
2. Replace the literal in code with `m.key()` / `m.key({ var })`.
3. Run `npm run build:i18n` to compile and confirm the message functions generate cleanly.
4. If the change fully migrates a new area, append its directory to `MIGRATED_GLOBS` in `scripts/i18n-guard.js`.
