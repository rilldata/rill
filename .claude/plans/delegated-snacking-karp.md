# Plan: Reframe "Deployments" to "Branches" on Project Status Page

## Context

The product direction positions branches (a content/authoring concept) as the primary frame for the cloud editing experience, rather than deployments (an infrastructure concept). Business users think "I'm working on my changes on a branch," not "I have a running deployment." The UI is already halfway there: the `BranchSelector`, URL structure (`@branch`), and branch utilities all think in branches. But the status page still thinks in deployments: a tab labeled "Deployments," deployment-centric section headers, empty states, and error messages.

This change reframes all user-facing text on the status page from deployment language to branch language, colocates the code in `features/branches/`, and adopts "hibernate/resume" terminology aligned with the 1-hour hibernation policy from the slots & pricing meeting (2026-04-14).

**This PR only reframes deployed branches.** A follow-up will show all git branches (see "Follow-Up" section at the bottom).

## Files to Move

Move three files from `web-admin/src/features/projects/status/deployments/` into `web-admin/src/features/branches/`:

| Old path | New path |
|---|---|
| `features/projects/status/deployments/DeploymentsSection.svelte` | `features/branches/BranchesSection.svelte` |
| `features/projects/status/deployments/DeleteDeploymentConfirmDialog.svelte` | `features/branches/DeleteBranchConfirmDialog.svelte` |
| `features/projects/status/deployments/deployment-actions.ts` | `features/branches/branch-actions.ts` |

Delete the now-empty `features/projects/status/deployments/` directory.

## User-Facing Text Changes

### `BranchesSection.svelte` (was `DeploymentsSection.svelte`)

| Line | Old | New |
|---|---|---|
| 217 | `Deployments` (section header) | `Branches` |
| 222 | `Loading deployments` | `Loading branches` |
| 226 | `Error loading deployments:` | `Error loading branches:` |
| 231 | `No deployments` | `No branches` |
| 181 | `Failed to ${actionName} deployment:` | `Failed to ${actionName} branch:` |
| 208 | `Failed to delete deployment:` | `Failed to delete branch:` |
| 392 | `Deploy a branch from the CLI:` | `Add a branch from the CLI:` |

The CLI command string `rill project deployment create <branch>` on line 394 stays as-is (it's a literal command the user types).

**Hibernate/resume terminology for actions:**

The actions dropdown currently uses "Start" and "Stop." These are deployment terms that don't make sense for branches. Replace with hibernate/resume language:

| Line(s) | Old | New |
|---|---|---|
| 321 | `Open editor` | (unchanged) |
| 333 | `View` / `Preview` | (unchanged) |
| 350 | `Start` | `Resume` |
| 368 | `Stop` | `Hibernate` |
| 382 | `Delete` | (unchanged) |

Also update the action names passed to `mutateDeployment` so error messages read correctly:
| Line | Old | New |
|---|---|---|
| 345 | `"start"` | `"resume"` |
| 363 | `"stop"` | `"hibernate"` |

### `DeleteBranchConfirmDialog.svelte` (was `DeleteDeploymentConfirmDialog.svelte`)

| Line | Old | New |
|---|---|---|
| 32 | `Delete this deployment?` | `Delete this branch?` |
| 35-37 | `The deployment on branch {branch} will be deleted. Any unpushed changes will be lost.` | `The branch {branch} will be deleted. Any unpushed changes will be lost.` |

### `BranchDeploymentStopped.svelte` (already in `features/branches/`)

| Line | Old | New |
|---|---|---|
| 78 | `Deployment is stopping...` | `Hibernating...` |
| 80 | `Deployment stopped` | `Branch hibernated` |
| 82 | `This branch deployment is not running.` | `This branch is hibernated.` |
| 91 | `Start deployment` | `Resume branch` |

### Status page layout (`routes/[organization]/[project]/-/status/+layout.svelte`)

| Line | Old | New |
|---|---|---|
| 17 | `label: "Deployments"` | `label: "Branches"` |

Also change `route: "/deployments"` on line 18 to `route: "/branches"`.

## Import Changes

### `BranchesSection.svelte`

These imports become local (siblings in `features/branches/`):
- `@rilldata/web-admin/features/branches/branch-utils` -> `./branch-utils`
- `@rilldata/web-admin/features/branches/deployment-utils` -> `./deployment-utils`
- `./deployment-actions` -> `./branch-actions`
- `./DeleteDeploymentConfirmDialog.svelte` -> `./DeleteBranchConfirmDialog.svelte`

The `display-utils` import stays as an alias path since it's in a different feature directory:
- `@rilldata/web-admin/features/projects/status/display-utils` (unchanged)

Also update the component tag on line 400: `<DeleteDeploymentConfirmDialog` -> `<DeleteBranchConfirmDialog`

### `branch-actions.ts` (was `deployment-actions.ts`)

- `@rilldata/web-admin/features/branches/deployment-utils` -> `./deployment-utils`

Function names (`optimisticallySetStatus`, `optimisticallyRemoveDeployment`) stay as-is; they describe cache-level operations on API objects, not user-facing concepts.

### Route page

Rename the route directory:
- `routes/[organization]/[project]/-/status/deployments/` -> `routes/[organization]/[project]/-/status/branches/`

In the renamed `+page.svelte`:
- `@rilldata/web-admin/features/projects/status/deployments/DeploymentsSection.svelte` -> `@rilldata/web-admin/features/branches/BranchesSection.svelte`
- `<DeploymentsSection` -> `<BranchesSection`

## Out of Scope

- **API types**: `V1Deployment`, `V1DeploymentStatus`, `createAdminServiceListDeployments` are auto-generated and must not change.
- **`deployment-utils.ts`** in `features/branches/`: function names `invalidateDeployments`, `isActiveDeployment`, `isProdDeployment` stay as-is (they describe API-level operations).
- **`display-utils.ts`** in `features/projects/status/`: stays in place. Used by 4 files across 3 feature directories; moving it expands blast radius for no user-visible benefit.
- **Overview card** (`DeploymentSection.svelte`): shows infrastructure status (runtime version, OLAP engine, etc.) which is genuinely about the deployment. Different concern, different tab.
- **"Slots" column**: will be addressed separately when column terminology is finalized.

## Verification

1. Run `npm run check` from `web-admin/` to verify TypeScript/Svelte compilation
2. Run `npm run quality` to verify lint/format
3. Start the cloud dev server (`rill devtool start cloud`) and verify:
   - Status page left nav shows "Branches" tab (not "Deployments")
   - Clicking the tab loads the branches table with correct data
   - Empty state says "No branches"
   - Delete dialog says "Delete this branch?"
   - Resume/hibernate actions work and error messages use correct terminology
   - Stopped branch page says "Branch hibernated" / "This branch is hibernated."
   - CLI hint still shows the correct copyable command

## Follow-Up: Show All Git Branches

This PR only shows branches that have a deployment. The team is converging on a richer model where the branches table shows **all git branches** with deployment status as an attribute. Context from the Slack discussion (2026-04-14, channel C07UWRB37GR):

- Deployments are 0-or-1-to-1 with branches, not 1-to-1. Deployments are created lazily only when someone wants to edit/preview a branch.
- The UI should consistently show all branches; ones without a deployment get a CTA to create one or open the editor.
- Abandoned deployments should auto-tear-down after X hours/days of inactivity, but the git branch should remain listed so users can resume later.

### What the richer model looks like

Each branch would appear in one of these states:

| State | Has deployment? | Status indicator | Actions |
|---|---|---|---|
| Active | Yes (running/pending/updating) | Status dot + "Active" label | Full actions |
| Hibernated | Yes (stopped) | Status dot + "Hibernated" label | "Resume" CTA |
| Not deployed | No | "Not deployed" label | "Deploy" or "Open editor" CTA |
| Merged/deleted | Orphaned deployment | Should auto-clean up (separate work) | — |


### Other follow-ups

- **Branch deployment auto-cleanup on merge/delete**: Rill currently does nothing when a branch is merged or deleted in git. The deployment keeps running until it errors. Needs webhook handling or polling to detect and clean up.
- **Compute observability**: The branches table should eventually surface per-branch compute allocation and total quota usage.
