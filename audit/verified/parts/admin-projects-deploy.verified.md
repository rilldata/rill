# Verified: admin-projects-deploy.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **12** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| admin-projects-deploy-04 | MEDIUM | P1 | Nil primary deployment panics handlers | `admin/server/reports.go:320`; `admin/server/virtual_file.go:77,177,236`; dereferences remain unguarded | Return FailedPrecondition when project has no primary deployment. |
| admin-projects-deploy-01 | MEDIUM | P0 | Failed editable checkpoint still destroys work | `admin/deployments.go:413-476`; failed `GitPush` only logs, then deprovision/branch/database deletion proceeds | Abort retryably on checkpoint failure or preserve a recoverable snapshot before teardown. |
| admin-projects-deploy-02 | MEDIUM | P0 | Any public project exposes private org assets | `admin/server/assets.go:186-194`; non-public asset is allowed when org has any public project, irrespective of asset type | Limit anonymous fallback to explicitly public branding assets; never serve private deploy archives. |
| admin-projects-deploy-06 | MEDIUM | P1 | LookupCanvas state test is inverted and panic-prone | `admin/canvases.go:14-36`; valid state returns NotFound, nil state is dereferenced, type assertion is unchecked | Correct nil/state guard and use checked resource assertion. |
| admin-projects-deploy-03 | MEDIUM | P2 | Custom-domain update never reaches deployments | `admin/server/organizations.go:1021-1050`; updated object is compared to requested value, making propagation branch dead | Capture old domain before update and compare old/new. |
| admin-projects-deploy-07 | MEDIUM | P2 | GitHub webhook blocks synchronously | `admin/github.go:344-411`; `admin/server/github.go:926-930`; all deployment triggers run inline and any error returns webhook failure | Enqueue an idempotent background job and acknowledge delivery promptly. |
| admin-projects-deploy-08 | MEDIUM | P1 | CreateProject leaves committed orphan on later failure | `admin/projects.go:24-124`; project commit precedes deployment creation/pointer update with no compensation | Make the flow convergent/idempotent and compensate project/repo/deployment on failure. |
| admin-projects-deploy-05 | LOW | P1 | GetRepoMeta nil-dereferences for ordinary caller | `admin/server/repos.go:58-65,92-93`; `depl` is optional but fields are unconditional | Guard response fields or require deployment-owned claims. |
| admin-projects-deploy-09 | LOW | P1 | Project reader can fetch another user’s private bookmark | `admin/server/bookmarks.go:39-69`; Get checks project read only, unlike update/remove ownership gates | Require owner unless bookmark is shared/default. |
| admin-projects-deploy-11 | LOW | P2 | Failed managed-repo row insert leaks GitHub repo | `admin/github.go:250-274`; external repo creation precedes DB insert with no cleanup | Best-effort delete on insert failure and add orphan reconciliation. |
| admin-projects-deploy-10 | LOW | P3 | GitHub callback writes two redirects | `admin/server/github.go:885-902`; non-collaborator redirect still has no return | Return immediately after retry redirect. |
| admin-projects-deploy-12 | LOW | P2 | Asset read failure is reported as HTTP 200 | `admin/server/assets.go:227-230`; status is committed before opening object reader | Open first, then set headers/status and stream. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| — | None | Current main still contains all twelve paths. PR #9753 only adds cleanup for old stopped deployments and is unrelated. | — |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| — | — | — | None | Each finding is directly established by current control flow. |

Counts: **12 pending, 0 fixed/remove, 0 partial/needs-test, 0 invalid = 12 total**.

## Machine-readable ledger

```text
PENDING|admin-projects-deploy.md|admin-projects-deploy-04|MEDIUM|P1|-|Nil primary deployment panics handlers|admin/server/reports.go:320; admin/server/virtual_file.go:77,177,236
PENDING|admin-projects-deploy.md|admin-projects-deploy-01|MEDIUM|P0|-|Failed editable checkpoint still destroys work|admin/deployments.go:413-476
PENDING|admin-projects-deploy.md|admin-projects-deploy-02|MEDIUM|P0|-|Public org project exposes private assets|admin/server/assets.go:186-194
PENDING|admin-projects-deploy.md|admin-projects-deploy-06|MEDIUM|P1|-|LookupCanvas inverted state check|admin/canvases.go:14-36
PENDING|admin-projects-deploy.md|admin-projects-deploy-03|MEDIUM|P2|-|Custom-domain propagation dead branch|admin/server/organizations.go:1021-1050
PENDING|admin-projects-deploy.md|admin-projects-deploy-07|MEDIUM|P2|-|GitHub webhook processed synchronously|admin/github.go:344-411
PENDING|admin-projects-deploy.md|admin-projects-deploy-08|MEDIUM|P1|-|CreateProject partial-failure orphan|admin/projects.go:24-124
PENDING|admin-projects-deploy.md|admin-projects-deploy-05|LOW|P1|-|GetRepoMeta nil deployment panic|admin/server/repos.go:58-65,92-93
PENDING|admin-projects-deploy.md|admin-projects-deploy-09|LOW|P1|-|Private bookmark ownership bypass|admin/server/bookmarks.go:39-69
PENDING|admin-projects-deploy.md|admin-projects-deploy-11|LOW|P2|-|Managed GitHub repository orphan|admin/github.go:250-274
PENDING|admin-projects-deploy.md|admin-projects-deploy-10|LOW|P3|-|GitHub callback missing return|admin/server/github.go:885-902
PENDING|admin-projects-deploy.md|admin-projects-deploy-12|LOW|P2|-|Asset handler commits success before read|admin/server/assets.go:227-230
```
