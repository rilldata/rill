# i18n Migration Audit

Baseline: 2026-07-27

This document records the remaining hardcoded-string findings discovered by
`node scripts/i18n-guard.js` and divides them into reviewable migration PRs.
The guard output is the live source of truth for line-level findings; counts in
this document are a snapshot and should decrease as PRs land.

## Summary

- Corrected baseline: 270 candidates across 86 files
- Current after PRs 2–9 and 12: 37 candidates across 17 files
- Current split: 37 candidates in `web-admin` and 0 in `web-common`
- No candidates in `web-local`
- Current categories: 12 visible-text findings, 20 attribute findings, and 5
  copy-property findings
- Currently, 6 candidates have an exact English-text match in the existing
  catalog; each still needs a semantic/context check before reusing that key
- Message catalogs currently pass integrity checks

The original guard reported no hardcoded strings because it covered only four
explicit path globs. The guard has been expanded to preserve those globs and
automatically scan Svelte, TypeScript, and JavaScript files that import the
generated message namespace. It now checks Svelte text and human-facing
attributes as well as common copy-bearing object properties such as `label`,
`description`, and `message`.

The corrected baseline also preserves Svelte control directives as text-scan
boundaries. The initial expanded scan merged literals from adjacent `{#if}` and
`{:else}` branches into single findings; splitting those branches added 10
independently actionable findings.

## Migration rules

For each candidate:

1. Reuse an existing message only when meaning and context match, not merely
   because the English text is identical.
2. Otherwise add matching English and Spanish messages using the documented key
   naming convention.
3. Use named interpolation and Paraglide variants instead of concatenation or
   manual plural logic.
4. Add `i18n-ignore` only for genuinely non-translatable values such as a CLI
   command or standalone product name. Mixed product-name sentences and page
   titles should generally still be messages.
5. Clear every guard finding in the PR's assigned paths.

Each migration PR should run:

```sh
npm run build:i18n
npm run test -w web-common
node scripts/i18n-guard.js
npm run quality
```

The final PR should additionally require:

```sh
node scripts/i18n-guard.js --strict
```

## PR plan

| PR  | Scope                                                      | Baseline candidates | Status    |
| --- | ---------------------------------------------------------- | ------------------: | --------- |
| 1   | Guard coverage improvements                                |                   — | Completed |
| 2   | Time presets and comparisons                               |                  37 | Completed |
| 3   | Canvas components and inspector                            |                  35 | Completed |
| 4   | Workspaces and Visual Metrics                              |                  18 | Completed |
| 5   | Dashboard UI, pivot, filters, and charts                   |                  30 | Completed |
| 6   | Resources, connectors, and models                          |                  16 | Completed |
| 7   | Shared alerts, chat, reports, exports, and components      |                  18 | Completed |
| 8   | Admin project status, GitHub, and user management          |                  29 | Completed |
| 9   | Admin edit sessions and branches                           |                  32 | Completed |
| 10  | Admin organizations, bookmarks, and view-as-user           |                  19 | Planned   |
| 11  | Admin alerts, reports, public URLs, and personal files     |                  18 | Planned   |
| 12  | Admin routes, page titles, onboarding, and access requests |                  18 | Completed |
| 13  | Final audit, intentional suppressions, and strict mode     |                   — | Planned   |

PRs 1–9 and 12 are completed. PRs 10 and 11 can be developed independently.
PR 13 should land only after warning mode is clean.

## Finding inventory

### PR 2: Time presets and comparisons (37, completed)

- 37 — `web-common/src/lib/time/config.ts`

### PR 3: Canvas components and inspector (35, completed)

- 4 — `web-common/src/features/canvas/CanvasBuilder.svelte`
- 3 — `web-common/src/features/canvas/components/kpi/KPI.svelte`
- 2 — `web-common/src/features/canvas/components/markdown/index.ts`
- 2 — `web-common/src/features/canvas/components/pivot/index.ts`
- 2 — `web-common/src/features/canvas/inspector/ComponentTabs.svelte`
- 2 — `web-common/src/features/canvas/inspector/LabelsInput.svelte`
- 2 — `web-common/src/features/canvas/inspector/chart/MetricsSQLInput.svelte`
- 13 — `web-common/src/features/canvas/inspector/chart/field-config/SortConfig.svelte`
- 3 — `web-common/src/features/canvas/inspector/filters/DimensionFiltersInput.svelte`
- 2 — `web-common/src/features/canvas/inspector/filters/TimeFiltersInput.svelte`

### PR 4: Workspaces and Visual Metrics (18, completed)

- 1 — `web-common/src/features/workspaces/ParquetWorkspace.svelte`
- 17 — `web-common/src/features/workspaces/VisualMetrics.svelte`

### PR 5: Dashboard UI, pivot, filters, and charts (30, completed)

- 2 — `web-common/src/features/dashboards/dimension-search/GlobalDimensionSearchResults.svelte`
- 2 — `web-common/src/features/dashboards/dimension-table/DimensionDisplay.svelte`
- 1 — `web-common/src/features/dashboards/dimension-table/DimensionTable.svelte`
- 4 — `web-common/src/features/dashboards/filters/measure-filters/measure-filter-options.ts`
- 1 — `web-common/src/features/dashboards/leaderboard/Leaderboard.svelte`
- 1 — `web-common/src/features/dashboards/pivot/PivotAutoArrangeZone.svelte`
- 1 — `web-common/src/features/dashboards/pivot/PivotChip.svelte`
- 3 — `web-common/src/features/dashboards/pivot/PivotEmpty.svelte`
- 4 — `web-common/src/features/dashboards/pivot/PivotSidebar.svelte`
- 1 — `web-common/src/features/dashboards/pivot/PivotTable.svelte`
- 2 — `web-common/src/features/dashboards/state-managers/actions/dimension-filters.ts`
- 2 — `web-common/src/features/dashboards/time-dimension-details/TDDHeader.svelte`
- 2 — `web-common/src/features/dashboards/time-dimension-details/TimeDimensionDisplay.svelte`
- 1 — `web-common/src/features/dashboards/time-series/ScreenshotContainer.svelte`
- 1 — `web-common/src/features/dashboards/time-series/measure-chart/MeasureChartBody.svelte`
- 2 — `web-common/src/features/dashboards/time-series/measure-chart/MeasureChartHoverTooltip.svelte`

### PR 6: Resources, connectors, and models (16, completed)

- 3 — `web-common/src/features/connectors/explorer/DatabaseExplorer.svelte`
- 5 — `web-common/src/features/connectors/explorer/DatabaseSchemaEntry.svelte`
- 2 — `web-common/src/features/models/workspace/ModelWorkspaceCTAs.svelte`
- 6 — `web-common/src/features/resources/ResourcesFilterableTable.svelte`

### PR 7: Shared interaction surfaces (18, completed)

- 1 — `web-common/src/components/forms/Select.svelte`
- 1 — `web-common/src/components/searchable-filter-menu/SearchableMenuContent.svelte`
- 1 — `web-common/src/features/alerts/criteria-tab/AlertPreview.svelte`
- 1 — `web-common/src/features/alerts/criteria-tab/CriteriaForm.svelte`
- 1 — `web-common/src/features/alerts/criteria-tab/CriteriaGroup.svelte`
- 4 — `web-common/src/features/alerts/delivery-tab/snooze.ts`
- 3 — `web-common/src/features/chat/connect/ConnectClientPopover.svelte`
- 4 — `web-common/src/features/chat/layouts/fullpage/ConversationSidebar.svelte`
- 1 — `web-common/src/features/exports/pdf/CanvasPdfExportHeader.svelte`
- 1 — `web-common/src/features/scheduled-reports/fields/RowsAndColumnsForm.svelte`

### PR 8: Admin projects (29, completed)

- 3 — `web-admin/src/features/projects/ProjectCard.svelte`
- 1 — `web-admin/src/features/projects/github/GithubConnectionDialog.svelte`
- 1 — `web-admin/src/features/projects/status/overview/ClusterSize.svelte`
- 1 — `web-admin/src/features/projects/status/resource-table/ActionsCell.svelte`
- 5 — `web-admin/src/features/projects/status/resource-table/RefreshAllSourcesAndModelsConfirmDialog.svelte`
- 2 — `web-admin/src/features/projects/status/resource-table/RefreshResourceConfirmDialog.svelte`
- 7 — `web-admin/src/features/projects/status/tables/ModelActionsCell.svelte`
- 1 — `web-admin/src/features/projects/status/tables/ProjectTables.svelte`
- 3 — `web-admin/src/features/projects/user-management/OrgUserGroupSetRole.svelte`
- 3 — `web-admin/src/features/projects/user-management/ProjectUserGroupSetRole.svelte`
- 2 — `web-admin/src/features/projects/user-management/UsergroupSetRole.svelte`

### PR 9: Admin edit sessions and branches (32, completed)

- 1 — `web-admin/src/features/branches/BranchesSection.svelte`
- 4 — `web-admin/src/features/edit-session/CommitPopover.svelte`
- 12 — `web-admin/src/features/edit-session/EditBranchDialog.svelte`
- 1 — `web-admin/src/features/edit-session/ExitButton.svelte`
- 14 — `web-admin/src/features/edit-session/MergePopover.svelte`

### PR 10: Admin organizations, bookmarks, and view-as-user (19)

- 1 — `web-admin/src/features/bookmarks/Bookmarks.svelte`
- 1 — `web-admin/src/features/bookmarks/BookmarksMenuItem.svelte`
- 3 — `web-admin/src/features/bookmarks/HomeBookmarkButton.svelte`
- 2 — `web-admin/src/features/organizations/settings/FaviconSettings.svelte`
- 2 — `web-admin/src/features/organizations/settings/LogoSettings.svelte`
- 2 — `web-admin/src/features/organizations/settings/ThumbnailSettings.svelte`
- 2 — `web-admin/src/features/organizations/user-management/dialogs/CreateUserGroupDialog.svelte`
- 2 — `web-admin/src/features/organizations/user-management/dialogs/EditUserGroupDialog.svelte`
- 4 — `web-admin/src/features/view-as-user/ViewAsUserPopover.svelte`

### PR 11: Admin operational surfaces (18)

- 1 — `web-admin/src/features/alerts/metadata/AlertFilterCriteria.svelte`
- 1 — `web-admin/src/features/alerts/metadata/AlertFilters.svelte`
- 4 — `web-admin/src/features/alerts/metadata/AlertMetadata.svelte`
- 3 — `web-admin/src/features/personal-files/canvas/CanvasPersonalFile.svelte`
- 1 — `web-admin/src/features/personal-files/canvas/PersonalCanvasCompositeCell.svelte`
- 3 — `web-admin/src/features/public-urls/CreatePublicURLForm.svelte`
- 2 — `web-admin/src/features/scheduled-reports/history/NoRunsYet.svelte`
- 3 — `web-admin/src/features/scheduled-reports/metadata/ReportMetadata.svelte`

### PR 12: Admin routes and onboarding (18, completed)

- 1 — `web-admin/src/routes/+page.svelte`
- 2 — `web-admin/src/routes/-/embed/+layout.svelte`
- 1 — `web-admin/src/routes/-/github/connect/request/+page.svelte`
- 1 — `web-admin/src/routes/-/github/connect/retry-install/+page.svelte`
- 6 — `web-admin/src/routes/-/welcome/theme/+page.svelte`
- 1 — `web-admin/src/routes/[organization]/-/create-project/+page.svelte`
- 1 — `web-admin/src/routes/[organization]/[project]/+page.svelte`
- 1 — `web-admin/src/routes/[organization]/[project]/-/dashboards/+page.svelte`
- 1 — `web-admin/src/routes/[organization]/[project]/-/edit/(viz)/explore/[name]/+page.svelte`
- 3 — `web-admin/src/routes/[organization]/[project]/-/request-access/[id]/approve/+page.svelte`

## Final strict-mode gate

After PRs 2–12 land:

1. Re-run the guard and investigate any newly exposed or newly introduced
   candidates.
2. Verify every `i18n-ignore` is attached to an intentional literal and has a
   clear reason in review.
3. Run catalog integrity, generated-message build, tests, linting, and quality
   checks.
4. Change the quality pipeline to invoke `scripts/i18n-guard.js --strict`.
5. Require the strict guard in CI for future changes.
