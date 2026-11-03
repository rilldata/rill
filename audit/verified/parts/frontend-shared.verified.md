# Verified audit: frontend-shared

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **12 numbered findings**. Every `fs-shared-01` through `-12` appears exactly once; `fs-shared-03` is correctly classified INVALID rather than FIXED.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| fs-shared-01 | high | P1 | Rename/delete helpers swallow failures | `entity-management/actions/actions.ts:206-232,258-276` still catch errors and do not rethrow/return failure | Propagate typed errors; only show success/update UI after success. |
| fs-shared-02 | medium | P1 | Reconcile versions compare lexicographically | `entity-management/actions/actions.ts:95-111` still compares string versions with `>` | Parse as BigInt/number and add 9→10 and timeout tests. |
| fs-shared-04 | medium | P1 | Failed connector flow can leave secrets in `.env` | `add-data/manager/steps/connector.ts:112-166`; `connectors/ai/saveAiConnector.ts:87-100`; `steps/import.ts:218-251` still write before operation success with incomplete rollback | Validate first or use a transaction/compensating cleanup on every exit. |
| fs-shared-05 | low | P2 | Upload URL contains raw filename | `features/sources/modal/file-upload.ts:18-27` still interpolates filename into the path | Encode the filename segment and test spaces, `#`, `?`, slash, Unicode. |
| fs-shared-06 | medium | P1 | Duplicate silently overwrites `(copy)` | `entity-management/actions/actions.ts:234-256` still writes the fixed target without collision/create-only semantics | Allocate a unique suffix or present replace/keep-both/cancel. |
| fs-shared-07 | low | P2 | Hand-built YAML fails to escape strings | `connectors/code-utils.ts:144,167,306-311`; `sources/sourceUtils.ts:84-89` still interpolate quoted values | Use a YAML serializer for scalar/header values. |
| fs-shared-08 | low | P2 | File dialog focus timer races change event | `features/sources/modal/file-upload.ts:36-63` still resolves empty after a fixed 1s focus delay | Resolve from change/cancel and guard one settlement. |
| fs-shared-09 | low | P2 | Force-delete dialog closes before delete resolves | `file-explorer/ForceDeleteConfirmationDialog.svelte:36-40`; `FileExplorer.svelte:120-126` | Await mutation, show busy/error, close only on success. |
| fs-shared-10 | low | P2 | Name regex admits dot/trailing-space names | `entity-management/name-utils.ts:1` still accepts `.`, `..`, and trailing dot/space | Add semantic validation beyond the regex and test OS-reserved edge cases. |
| fs-shared-11 | low | P2 | Connector name is used as raw regex/replacement | `connectors/code-utils.ts:342-365,386-397` still builds dynamic regex/replacements without escaping | Escape regex input and use replacement callbacks. |
| fs-shared-12 | low | P3 | Theme sanitizer/injection code is dead | `features/themes/css-sanitizer.ts`; `themes/theme-manager.ts:195-256,341-353` remain unreferenced | Delete dead surface and its tests, or wire one supported path. |

## Fixed / remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| fs-shared-03 | INVALID; remove as vulnerability | The sanitizer is not on a reachable theme application path; the source audit itself refuted the claimed rule-breakout. Dead code remains separately tracked by `fs-shared-12`. | No fix commit; invalid premise |

## Partial / needs test

None.

Counts: **11 pending, 0 fixed, 0 partial, 1 invalid/remove = 12**.

## Machine-readable inventory

```text
PENDING|frontend-shared.md|fs-shared-01|high|P1|-|Rename/delete swallow failures|actions.ts:206-232,258-276
PENDING|frontend-shared.md|fs-shared-02|medium|P1|-|String version comparison|actions.ts:95-111
INVALID|frontend-shared.md|fs-shared-03|REMOVED-refuted|-|-|Unreachable sanitizer injection|css-sanitizer.ts:87-136
PENDING|frontend-shared.md|fs-shared-04|medium|P1|-|Secrets persist after failed connector flow|connector.ts:112-166
PENDING|frontend-shared.md|fs-shared-05|low|P2|-|Raw upload filename URL|file-upload.ts:18-27
PENDING|frontend-shared.md|fs-shared-06|medium|P1|-|Duplicate overwrites copy|actions.ts:234-256
PENDING|frontend-shared.md|fs-shared-07|low|P2|-|YAML strings unescaped|code-utils.ts:306-311
PENDING|frontend-shared.md|fs-shared-08|low|P2|-|File dialog timing race|file-upload.ts:36-63
PENDING|frontend-shared.md|fs-shared-09|low|P2|-|Delete dialog closes early|ForceDeleteConfirmationDialog.svelte:36-40
PENDING|frontend-shared.md|fs-shared-10|low|P2|-|Degenerate names accepted|name-utils.ts:1
PENDING|frontend-shared.md|fs-shared-11|low|P2|-|Unescaped regex/replacement|code-utils.ts:342-397
PENDING|frontend-shared.md|fs-shared-12|low|P3|-|Dead theme CSS code|css-sanitizer.ts
```
