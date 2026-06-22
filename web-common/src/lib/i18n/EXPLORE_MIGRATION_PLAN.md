# Explore Dashboard i18n Migration Plan (Paraglide)

Status tracker + path forward for localizing the **explore dashboard experience**
(`web-admin/src/routes/[organization]/[project]/explore` and the shared
`web-common` dashboard tree it renders) with Paraglide JS (inlang).

This file is the source of truth across chat sessions. Update the **Progress**
checkboxes as chunks land.

## Background (already in place — no setup needed)

- `@inlang/paraglide-js@^2.10` installed at repo root.
- Vite plugin configured in `web-admin/vite.config.ts` and `web-local` with
  strategy `["localStorage", "preferredLanguage", "baseLocale"]`.
- Messages: `web-common/src/lib/i18n/messages/{en,de}.json`
  (`en` is base/source, `de` is machine-translated). Keep keysets identical.
- Generated functions: `import { m } from "@rilldata/web-common/lib/i18n/gen/messages"`,
  used as `{m.some_key()}`. The `gen/` dir is gitignored and auto-compiled by
  Vite in dev; regenerate manually with `npm run build:i18n`.
- Project config: `web-common/src/lib/i18n/project.inlang/settings.json`
  (baseLocale `en`, locales `["en", "de"]`).
- Guard: `scripts/i18n-guard.js` scans the file globs in `MIGRATED_GLOBS` for
  hardcoded copy in markup text + the attrs `placeholder/title/aria-label/alt/label`.
  Runs warning-only today via `scripts/web-test-code-quality.sh`; the **final
  chunk** flips it to `--strict` (fatal).
- Two chunks already migrated: **A** (org overview), **B** (project overview).

## Scope

Full explore experience = route files + the shared dashboard tree they mount.

| Area | Files | Notes |
|---|---|---|
| web-admin explore route | 2 | 3 strings (404 page); `+layout.svelte` has none |
| web-common `features/explores/` | 6 | edit/preview menus, CTAs |
| web-common `features/dashboards/` | 140 | the bulk — zero paraglide usage today |

`web-common/src/features/dashboards` immediate-subdir `.svelte` counts:

```
25 time-controls      8 dimension-table     2 tab-bar
25 pivot              5 time-dimension-details  2 rows-viewer
19 time-series        4 toolbars            2 listing
19 filters            4 state-managers      2 workspace
11 leaderboard        3 dimension-search    2 (root)
                      3 big-number          1 each: url-state, stores,
                                              granular-access-policies, errors
```

## Conventions

- **Key naming:** `feature_component_purpose`, e.g. `dashboards_filters_clear_all`.
- **Add keys to BOTH** `en.json` and `de.json` (identical keysets).
- **Usage:** `import { m } from "@rilldata/web-common/lib/i18n/gen/messages"` then `{m.key()}`.
- **Interpolation:** named placeholders `{variable}` — never string concatenation.
  (e.g. existing `projects_redeploy_wake_failed: "Failed to wake project: {error}"`).
- **Pluralization:** use Paraglide variants, not hand-rolled `n === 1` logic.
- **Suppress** intentional non-copy with an `i18n-ignore` comment on the line or line above.
- **Never hand-edit** `web-common/src/lib/i18n/gen/` — it is generated.

## Per-file workflow (same every chunk)

1. Find user-facing copy: visible text nodes + human-facing attrs
   (`placeholder`, `title`, `aria-label`, `alt`, `label`). Skip identifiers,
   event-bus names, CSS, URLs, paths.
2. Add keys to `en.json` **and** `de.json`.
3. Replace strings with `m.*()`; dynamic values via named placeholders.
4. Append the chunk's file globs to `MIGRATED_GLOBS` in `scripts/i18n-guard.js`.
5. Verify (below).

## Verification per chunk

- `npm run build:i18n` (regenerate `gen/`).
- `node scripts/i18n-guard.js` — clean for the newly-added globs.
- `npm run test -w web-common` for affected components.
- Visual smoke of the explore dashboard in `rill devtool start cloud`.
- `npm run quality`.

## Chunks (ordered)

**All chunks land in ONE PR** (branch `feat/localization-dashboards`) — partial
i18n migrations are awkward, so the whole explore experience ships together. The
chunks below are units of *work split across chat sessions*, not separate PRs.
Each session: pick the next unchecked chunk, migrate it, append its globs to the
guard, tick the box here. The final chunk (L) flips the guard to `--strict`.

- [x] **C — route files** (DONE). Migrated
      `web-admin/.../explore/[dashboard]/+page.svelte` (404 header/body + page
      `<title>`) and `web-admin/src/features/dashboards/DashboardErrored.svelte`
      (errored-state copy; "Discord" left as `i18n-ignore` brand name).
      Keys added: `dashboards_page_title`, `dashboards_not_found_header`,
      `dashboards_not_found_body`, `dashboards_errored_title`,
      `dashboards_errored_body_manage`, `dashboards_errored_body_read`,
      `dashboards_errored_view_status`, `dashboards_errored_view_project`,
      `dashboards_errored_help_prefix`. Globs added to guard.
- [x] **D — explores feature** (DONE). Migrated ExploreEditDropdown (reused
      `common_edit`), PreviewButton, ExploreMenuItems, explore-link/ExploreLink.
      ExplorePreviewCTAs + ExploreEditor had no user-facing copy. Keys added:
      `explores_edit_dropdown_explore`, `explores_edit_dropdown_metrics_view`,
      `explores_preview_button`, `explores_preview_tooltip_{reconciling,disabled,default}`,
      `explores_menu_view_dag`, `explores_link_goto_{named,default,short}`,
      `explores_link_error_title`. Guard glob: `web-common/src/features/explores/**/*.svelte`.
- [x] **E — dashboard shell/chrome** (DONE). Migrated `workspace/` (Dashboard
      mock-user no-access copy, DeployProjectCTA button/tooltip + deploy
      notifications), `toolbars/` (Exclude/Search/SelectAll/StartPivot),
      `tab-bar/` (Tab "Coming Soon", TabBar Explore/Pivot labels — `tabs` made
      reactive; nested-brace `{#each}` suppressed with `i18n-ignore`), `listing/`
      (DashboardsTable + CompositeCell, reusing the chunk-B `dashboards_listing_*`
      keys + `common_error`; `&rarr;` → literal `→` to match web-admin),
      `errors/InlineErrorIndicator`, and root `DashboardBuilding`
      (singular/plural keys). `ThemeProvider` had no user-facing copy. New keys:
      `dashboards_mock_user_no_access_{header,body}`, `dashboards_deploy_{button,tooltip,auth_failed,remote_changes_merged}`,
      `dashboards_toolbar_{exclude_label,exclude_output_exclude,exclude_output_include,exclude_toggle_include,exclude_toggle_exclude,shortcut_click,search,select_all,deselect_all,deselect_all_tooltip,start_pivot}`,
      `dashboards_tab_{explore,pivot,coming_soon}`, `dashboards_error_indicator_{aria,label,no_details,copy,copied}`,
      `dashboards_building_{single,multiple}`. Globs added to guard.
- [x] **F — filters** (DONE). Migrated all of `filters/` (Filters, AdvancedFilter,
      Canvas/FilterButton, Canvas/Explore/FilterChipsReadOnly, Pin/Required
      buttons, TimeRangeReadOnly, `dimension-filters/`, `measure-filters/`) and
      `dimension-search/`. Introduced the first **Paraglide plural variants**
      (`results_count`, `chip_others`, `remove_values`, `dim_search_results`) —
      JSON shape is an array-wrapped complex message with `declarations` /
      `selectors` / `match`. Converted two `.ts` helpers that feed chunk-F
      markup: `dimension-filters/helpers.ts` `getSearchPlaceholder` (now returns
      `m.*()`) and `dimension-filters/constants.ts` `DimensionFilterModeOptions`
      → `getDimensionFilterModeOptions()` (a function so labels resolve in the
      active locale, not at import; `DimensionFilterModeSelector` calls it
      reactively). Reused `dashboards_toolbar_{exclude_label,select_all,deselect_all}`.
      Suppressed two `{#each … as { … }}` destructuring lines + one progress
      `< 100` heuristic false positive (moved the comparison into the script).
      Keys added under `dashboards_filters_*` and `dashboards_dim_search_*`.
      Globs added to guard: `filters/**/*.svelte`, `dimension-search/**/*.svelte`.
      **Known gap:** `measure-filters/measure-filter-options.ts` (operation/type
      option `label`/`shortLabel`/`description`, e.g. "Greater Than", "% change")
      is left in English — it is a module-level constant **shared with the
      unmigrated `alerts` feature** (`alerts/utils.ts`,
      `alerts/criteria-tab/getTypeOptions.ts`), so localizing it belongs to a
      later chunk that also covers alerts. The guard does not police `.ts`, so
      this does not block `--strict`.
- [ ] **G — time controls** (`time-controls/`, 25). Heavy dynamic-label area.
- [ ] **H — pivot** (`pivot/`, 25).
- [ ] **I — time-series + big-number** (19 + 3).
- [ ] **J — leaderboard** (11).
- [ ] **K — dimension-table + rows-viewer** (8 + 2).
- [ ] **L — time-dimension-details** (5) + sweep low-string dirs
      (`state-managers`, `url-state`, `stores`, `granular-access-policies` —
      mostly logic). **This chunk flips the guard to `--strict`** in
      `scripts/web-test-code-quality.sh`.

## Pre-existing guard warnings (clean up before flipping `--strict` in chunk L)

`node scripts/i18n-guard.js` already reports ~24 hardcoded strings in *earlier*
migrated areas (chunks A/B and shared layout), e.g.:
`web-common/src/layout/workspace/WorkspaceHeader.svelte`,
`web-common/src/layout/navigation/Navigation.svelte`,
`web-common/src/features/welcome/TitleContent.svelte`,
`web-common/src/features/onboarding/OnboardingWorkspace.svelte`.
These are warning-only today but will be **fatal** once chunk L adds `--strict`.
Sweep them as part of chunk L (or earlier).

## Risks / watch-outs

- **Embedded dashboards:** the same `web-common/dashboards` tree powers iframe
  embeds. Confirm locale resolution works there (localStorage access /
  cross-origin) — strategy is `["localStorage","preferredLanguage","baseLocale"]`.
- **SPA, no SSR:** web-admin is `adapter-static`, `ssr=false`. Locale is purely
  client-resolved; no server-side negotiation.
- **Dynamic labels:** dashboard code builds many labels from counts, dimension
  names, and time ranges. Design placeholders carefully (esp. `time-controls`,
  `filters`) — do not naively split concatenated strings.
- **de.json drift:** every `en` key must exist in `de`. Consider a CI check that
  the two files have identical keysets.
