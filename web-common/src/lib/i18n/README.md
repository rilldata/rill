# Internationalization (i18n)

Rill's frontend is localized with [Paraglide JS](https://inlang.com/m/gerre34r/library-inlang-paraglideJs)
(inlang). Translatable strings live in this directory and compile to
tree-shakeable, type-safe message functions consumed by both `web-local`
(Rill Developer) and `web-admin` (Rill Cloud).

See [`MIGRATION.md`](./MIGRATION.md) for the chunked plan to migrate existing
hardcoded strings.

## Layout

```
lib/i18n/
├── messages/{en,de}.json   # source translations (edit these)
├── project.inlang/         # inlang project config
└── gen/                    # compiled message functions (auto-generated, gitignored)
```

- `messages/en.json` is the base locale and the source of truth for keys.
- `messages/de.json` is generated via machine translation; do not hand-maintain it.
- `gen/` is compiled output. It is gitignored — never edit or import individual
  files by hand other than the documented entry points below.

## Build

```sh
npm run build:i18n        # compile messages/ -> gen/
npm run machine-translate # fill in non-English locales from en.json
```

The Vite plugin in both `web-local` and `web-admin` recompiles `gen/` on dev and
build, so you rarely need to run `build:i18n` manually during development.

## Usage

```svelte
<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<button>{m.common_cancel()}</button>
<p>{m.welcome_greeting({ name })}</p>
```

- Import the `m` namespace; call messages as functions.
- Pass interpolation values as a named object: `m.welcome_greeting({ name })`.
- Override the locale per call when needed: `m.common_cancel({}, { locale: "de" })`.

Locale detection uses the `["preferredLanguage", "baseLocale"]` strategy: the
browser's preferred language, falling back to English.

## Conventions

### Key naming

Use `feature_component_purpose`, lower snake_case, grouped by prefix in
`en.json`:

```jsonc
{
  "common_cancel": "Cancel",
  "common_save": "Save",
  "welcome_greeting": "Welcome, {name}",
  "dashboards_filters_clear_all": "Clear all filters"
}
```

- `common_` — copy reused across features. Reuse an existing `common_` key
  rather than duplicating identical copy.
- Otherwise prefix with the feature directory name.

### Interpolation

Use named placeholders, never string concatenation:

```jsonc
{ "exports_row_count": "Exporting {count} rows" }
```

### Pluralization and variants

Use Paraglide [variants](https://inlang.com/m/gerre34r/library-inlang-paraglideJs/variants)
rather than hand-rolled `count === 1 ? ... : ...` logic.

## Adding or migrating a string

1. Add the key to `messages/en.json` following the naming convention.
2. Run `npm run machine-translate` to populate other locales.
3. Replace the literal in code with `m.key()` (or `m.key({ var })`).
4. `npm run build:i18n` (or rely on the Vite plugin in dev).
5. Run `npm run test -w web-common` and `npm run quality`.

## Guard against new hardcoded strings

`scripts/i18n-guard.js` scans already-migrated areas for hardcoded
user-facing strings and runs in `npm run quality`. It is a heuristic and
currently **warning-only**; the final migration chunk flips it to `--strict`
(fatal). Each migration chunk appends its directories to `MIGRATED_GLOBS` in
that script. Suppress an intentional literal with an `i18n-ignore` comment on
the line or the line above it.
