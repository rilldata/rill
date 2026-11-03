# Admin Projects & Deployments — Code Review
**Scope:** `admin/server/{projects,deployment,organizations,github,repos,reports,alerts,bookmarks,provision,virtual_file,assets,telemetry}.go`, `admin/{projects,deployments,provision,github,alerts,reports,canvases,urls}.go` (plus `admin/server/runtime_jwt.go` and the auth redirect path they call into)  |  **Health:** fair

## Overview
This area implements project/deployment CRUD, provisioning orchestration, GitHub integration (OAuth, managed repos, webhooks), editable-deployment checkpointing, custom domains, and magic-token issuance for reports/alerts. The permission model is generally applied consistently (every gRPC handler resolves `ProjectPermissions`/`OrganizationPermissions` before acting), the JWT issuance path (`runtime_jwt.go`) is careful, and asset serving deliberately hardens against SVG/script execution. The code openly acknowledges it is not fault tolerant (`admin/projects.go:21`), and that shows up as the biggest structural risk: multi-step external operations (provision → DB update → git push) are not transactional or idempotent, so failures leave orphaned projects and repos. The most important concrete risks are: (1) a cluster of unguarded `*proj.PrimaryDeploymentID` / `depl` dereferences that panic on hibernated or undeployed projects; (2) editable-deployment deletion that wipes the PVC and deletes the branch even when the checkpoint push fails; (3) a broad public-asset fallback that can expose private project archives to anyone who learns the asset UUID; and (4) a dead conditional that silently drops custom-domain propagation to deployments.

## Findings

### admin-projects-deploy-04 [MEDIUM] Unguarded `*proj.PrimaryDeploymentID` dereferences panic on hibernated/undeployed projects
- **Category:** bug  **Confidence:** high  **Location:** `admin/server/reports.go:320` (`UnsubscribeReport`); `admin/server/virtual_file.go:77` (`CreatePersonalFile`), `:177` (`EditPersonalFile`), `:236` (`DeletePersonalFile`)
- **Problem:** These four handlers dereference `*proj.PrimaryDeploymentID` without the `if proj.PrimaryDeploymentID == nil` guard that every sibling handler has. Verified: `CreateReport`/`EditReport` guard at `reports.go:194/255`; all alert handlers guard (`alerts.go:147/211/283/429/503`); the deployment handlers guard (`deployment.go:520/604`). `PrimaryDeploymentID` is nil for hibernated projects (set nil on delete in `hibernate_expired_deployments.go:84` and `admin/projects.go`) and for projects created with `skip_deploy`.
- **Failure scenario:** A user with `ReadProject` opens the personal-canvas UI on a hibernated (or never-deployed) project, or a recipient clicks an unsubscribe link for a report whose project was later hibernated. `*proj.PrimaryDeploymentID` dereferences a nil pointer; the request panics and is converted to a 500 by the gRPC recovery interceptor instead of a clean `FailedPrecondition`.
- **Recommendation:** Add the same `if proj.PrimaryDeploymentID == nil { return FailedPrecondition }` guard used by the other handlers before each deref.

**Verification:** CONFIRMED — `admin/server/reports.go:320` (`UnsubscribeReport`) and `admin/server/virtual_file.go:77/177/236` (`CreatePersonalFile`/`EditPersonalFile`/`DeletePersonalFile`) all dereference `*proj.PrimaryDeploymentID` with no nil guard, while sibling handlers (`reports.go:194/255`) guard with a clean `FailedPrecondition`.

**Manual Reproduction Steps (for QA):**
1. Deploy a project to Rill Cloud and create a scheduled report with your email as a recipient; wait for one report email to arrive.
2. Hibernate the project from a terminal: `rill project hibernate <project-name>`.
3. Open the "Unsubscribe" link from the report email (or call `POST /v1/organizations/{org}/projects/{project}/reports/{name}/unsubscribe`).
4. Observe an internal-error (HTTP 500 / gRPC Internal) response instead of a clean "project does not have a production deployment" message.
5. Alternatively: while the project is hibernated, open one of its dashboards in the Cloud UI and try to create a personal canvas ("+ New canvas" personal file flow); the create call fails with a 500 for the same reason.

### admin-projects-deploy-01 [MEDIUM] Editable deployment deletion discards uncommitted work when the checkpoint push fails
- **Category:** data-loss  **Confidence:** high  **Location:** `admin/deployments.go:413-421` (best-effort `GitPush`), `:462-473` (branch delete), `:476` (deployment delete)
- **Problem:** `DeleteDeploymentInner` checkpoints an editable deployment via `rt.GitPush`, but the error is only logged and execution proceeds unconditionally to deprovision the runtime (wiping the PVC), delete the `rill/xxxx` branch, and delete the deployment row. This is asymmetric with `StopDeploymentInner` (`:340-396`), which is also best-effort on the push but *preserves* the PVC via hibernation, so a failed checkpoint there is recoverable.
- **Failure scenario:** The primary realistic path is a **direct** teardown of a running editable deployment — `DeleteDeployment` RPC (`admin/server/deployment.go:498`) or `TeardownProject` (`admin/projects.go:146`) — where `DeleteDeploymentInner` runs the only checkpoint attempt; if `GitPush` fails (runtime briefly unreachable or a transient GitHub 5xx), the un-pushed browser edits on the PVC are permanently lost with no user-visible signal. Note the TTL path is narrower than a naive read suggests: `hibernate_expired_deployments.go` first *stops* (checkpoints + preserves PVC) and only *deletes* after `StoppedDeploymentRetention`, so TTL-driven data loss additionally requires the earlier stop checkpoint to have failed.
- **Recommendation:** At minimum, do not delete the branch or deprovision when the checkpoint push did not succeed for an editable, managed-branch deployment; return the error and let the retryable job re-run, or snapshot the working tree before deprovisioning.

**Verification:** CONFIRMED — `admin/deployments.go:413-421` logs the `GitPush` error and falls through unconditionally to deprovisioning (`:434-459`), branch deletion (`:462-473`), and deployment deletion (`:476`); by contrast `StopDeploymentInner` (`:340-396`) hibernates (preserving the PVC) after a failed push.

**Manual Reproduction Steps (for QA):**
1. This requires inducing a transient GitHub/runtime push failure at teardown time, which is not achievable through the product UI alone; treat as fault-injection testing.
2. API/staging-level repro: create an editable (dev) deployment for a GitHub-connected project, make edits in the browser editor, and do not push/save them.
3. Induce a checkpoint push failure (e.g. in a staging environment, temporarily revoke the GitHub App installation token or block egress from the runtime to github.com).
4. Delete the deployment (`DeleteDeployment` RPC) or delete the project (`rill project delete <name>`).
5. Restore GitHub access and inspect the `rill/xxxx` branch: the unsaved edits are gone, the branch is deleted, and no error was surfaced to the user (only a server-side log line "failed to checkpoint repo changes").

### admin-projects-deploy-02 [MEDIUM] `assetHandler` serves private assets (including project archives) whenever the org has any public project
- **Category:** security  **Confidence:** medium  **Location:** `admin/server/assets.go:186-194`
- **Problem:** For a non-public asset, if the caller lacks `ReadOrg` the handler falls back to `CheckOrganizationHasPublicProjects`; if the org has *any* public project, the asset is served. This fallback is asset-type-agnostic, but `deploy` archives (full project source, uploaded with `Public` unset → false, see `cli/pkg/cmdutil/archive.go:39-45`) share the asset namespace with branding images. The route (`server.go:254`) admits anonymous callers, so `claims` may carry no `ReadOrg`.
- **Failure scenario:** Org `acme` has one public dashboard. `acme`'s private project uploaded a `deploy` archive (asset ID is a UUID). Anyone who learns/leaks that UUID (logs, referrer, a former member) can `GET /v1/assets/<id>/download` and retrieve the entire private project source with no org membership. This is UUID-gated (the ID is unguessable), so it is a defense-in-depth exposure rather than a directly enumerable one — but the payload (full source) makes it worth closing.
- **Recommendation:** Scope the public-projects fallback to branding image types only, or gate it on `asset.Public`. Private `deploy` archives should never be reachable via the public-projects path.

**Verification:** CONFIRMED — `admin/server/assets.go:186-194`: for a non-public asset, a caller without `ReadOrg` is allowed through whenever `CheckOrganizationHasPublicProjects` returns true, with no check on the asset type. Deploy archives are created with `Public` unset (false) in `cli/pkg/cmdutil/archive.go:39-45`, and the route (`admin/server/server.go:254`) accepts unauthenticated requests.

**Manual Reproduction Steps (for QA):**
1. In an org that has at least one public project, deploy a second, private project from a local folder (`rill deploy` without GitHub), so a `deploy` archive asset is uploaded.
2. As an org member, obtain the archive's asset ID (e.g. from the project's `ArchiveAssetID` via the `GetProject` admin API, or from the admin database in a staging environment).
3. Open a private/incognito browser window (not logged in to Rill) and request `https://admin.rilldata.com/v1/assets/<asset-id>/download`.
4. The full private project source archive downloads. Expected behavior would be HTTP 403.
5. Control check: repeat in an org with no public projects; the same request correctly returns 403.

### admin-projects-deploy-06 [MEDIUM] `LookupCanvas` has an inverted state check: returns NotFound when the spec exists and panics when it does not
- **Category:** correctness  **Confidence:** high  **Location:** `admin/canvases.go:32-36`
- **Problem:** The guard returns `NotFound` when `canvasResource.Canvas.State != nil` (i.e. exactly when a valid spec *is* available), and otherwise falls through to `res.Resource.Resource.(*runtimev1.Resource_Canvas).Canvas.State.ValidSpec`, which nil-derefs when `State == nil`. This is the negation of the alert/report equivalents, which return the spec directly (`admin/reports.go:47`, `admin/alerts.go`). The single-value type assertion at `:36` also panics if the resource is not a canvas. Net effect: the function can never fulfill its stated contract — it either returns `NotFound` or panics.
- **Failure scenario:** The only current caller is `CreatePersonalFile`'s name-availability check (`virtual_file.go:82-85`) via `generateVirtualFileName`. For a *non-existent* canvas the earlier `rt.GetResource` error path returns `NotFound` before this code runs, so the common case happens to work. But if a canvas with the candidate name *exists* and has a state, `LookupCanvas` reports `NotFound`, so the collision check treats a taken name as free (masked in practice by the random UUID suffix on names); if it exists but has not finished reconciling (`State == nil`), the check panics. Any future caller that actually needs the spec is fully broken.
- **Recommendation:** Invert the guard: `if canvasResource == nil || canvasResource.Canvas == nil || canvasResource.Canvas.State == nil { return nil, NotFound }; return canvasResource.Canvas.State.ValidSpec, nil`, and use the comma-ok type assertion.

**Verification:** CONFIRMED — `admin/canvases.go:33-36`: the guard returns `NotFound` when `State != nil` (i.e. when a valid spec exists) and otherwise dereferences `...Canvas.State.ValidSpec`, which nil-panics when `State == nil`. The single-value type assertion at `:32/:36` also panics on a non-canvas resource. The contract is inverted versus the code.

**Manual Reproduction Steps (for QA):**
1. Largely masked in production because the sole caller (`CreatePersonalFile` name-availability check) appends a random UUID suffix and the non-existent-canvas case is handled by an earlier `GetResource` error.
2. To exercise the broken path: in a project that already has a canvas dashboard fully reconciled, drive a `CreatePersonalFile` request whose generated candidate name collides with that existing canvas name; the collision check silently treats the taken name as free (functionally harmless due to the UUID suffix).
3. Confirm in code rather than UI: `admin/canvases.go:33-36` — verify the `State != nil` condition returns `NotFound` and that the fallthrough line dereferences `State.ValidSpec`.

### admin-projects-deploy-03 [MEDIUM] `SudoUpdateOrganizationCustomDomain` never propagates the change to deployments (dead conditional)
- **Category:** correctness  **Confidence:** high  **Location:** `admin/server/organizations.go:1021` (reassigns `org`), `:1050` (guard)
- **Problem:** `org` is reassigned to the result of `UpdateOrganization` (whose `CustomDomain` now equals `req.CustomDomain`) before the guard `if req.CustomDomain != org.CustomDomain`. The comparison is therefore always false, so `UpdateOrgDeploymentAnnotations` (`admin/projects.go:308`) is never called. `UpdateOrganization` itself computes `nameChanged` correctly *before* its update — this handler got the ordering wrong.
- **Failure scenario:** A superuser sets or changes an org's custom domain. Deployment annotations (`organization_custom_domain`) and the `frontend_url` the runtime uses to build report/alert email links are not refreshed, so those links keep pointing at the old/absent domain until an unrelated deployment update happens to recompute annotations. Superuser-only and eventually self-healing, but user-visibly wrong in the interim.
- **Recommendation:** Capture `oldCustomDomain := org.CustomDomain` before the `UpdateOrganization` call and compare against it.

**Verification:** CONFIRMED — `admin/server/organizations.go:1021` reassigns `org` to the `UpdateOrganization` result (whose `CustomDomain == req.CustomDomain`) before the guard `if req.CustomDomain != org.CustomDomain` at `:1050`, so the comparison is always false and `UpdateOrgDeploymentAnnotations` is never called.

**Manual Reproduction Steps (for QA):**
1. As a superuser, set a custom domain on an org: `rill sudo org set-custom-domain <org> <domain>` (or the `SudoUpdateOrganizationCustomDomain` API).
2. Immediately trigger a scheduled report or alert email for a project in that org.
3. Inspect the links in the resulting email: they still point at the old/default `rilldata.com` URL rather than the new custom domain.
4. Confirm in code: `admin/server/organizations.go:1050` — the `if req.CustomDomain != org.CustomDomain` branch is dead because `org` was reassigned at `:1021`, so `UpdateOrgDeploymentAnnotations` never runs.
5. Note it self-heals: any later unrelated deployment update recomputes annotations, after which links become correct.

### admin-projects-deploy-07 [MEDIUM] GitHub webhook is processed synchronously despite a "moved to background" contract
- **Category:** performance  **Confidence:** high  **Location:** `admin/server/github.go:905-933` (handler), `admin/github.go:344-346` (docstring), `:362-411` (`processGithubPush`)
- **Problem:** `ProcessGithubEvent` is documented as moving work "to the background and returns a nil error," but `processGithubPush` runs inline: it fans out over every project matching the pushed remote and every prod deployment, calling `TriggerParser` sequentially. `TriggerParser` (`admin/projects.go:452-473`) opens a fresh runtime client (JWT issue + gRPC dial) and makes a `CreateTrigger` RPC per deployment. The webhook handler blocks on all of it and returns `http.StatusBadRequest` on any error (`:928`), which GitHub treats as a delivery failure and redelivers. `context.Background()` also detaches the work from the request lifecycle without actually making the response asynchronous.
- **Failure scenario:** A widely-shared remote maps to many projects/deployments, or one runtime is slow/unreachable. Processing exceeds GitHub's ~10s webhook timeout; GitHub marks it failed and redelivers, re-triggering parsers for all deployments (including ones that already succeeded — harmless per-parse but wasteful). For the common single-project remote this is a non-issue; the docstring is simply wrong.
- **Recommendation:** Enqueue a job (like the reconcile jobs) and return 200 immediately; bound concurrency and per-deployment timeouts; return 2xx for transient per-deployment failures so GitHub does not storm-retry. At minimum, correct the docstring to match the synchronous behavior.

**Verification:** CONFIRMED — `ProcessGithubEvent`'s docstring (`admin/github.go:344-346`) claims work is moved "to the background," but `processGithubPush` (`admin/github.go:362-411`) runs inline, iterating over all matching projects/deployments and calling `TriggerParser` sequentially; the webhook handler (`admin/server/github.go:926-930`) blocks on it and returns HTTP 400 on any error, prompting GitHub redelivery. `context.Background()` here only detaches cancellation, it does not make the call asynchronous.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible through normal single-project flows (the common case completes well within GitHub's ~10s timeout).
2. To exercise the slow path in a staging setup: connect several projects/prod deployments to one shared GitHub remote, make one runtime slow or unreachable, then push a commit to that remote.
3. In the GitHub repo's Settings > Webhooks > Recent Deliveries, observe the delivery time out / show a non-2xx response and get redelivered.
4. Confirm in code: `admin/github.go:344-346` docstring vs the synchronous `processGithubPush` at `:362-411`, and the blocking call at `admin/server/github.go:926`.

### admin-projects-deploy-08 [MEDIUM] `CreateProject` leaves an orphaned project (and possibly a managed GitHub repo) on partial failure
- **Category:** correctness  **Confidence:** high  **Location:** `admin/projects.go:24-135` (commit at `:66`, then `CreateDeployment` `:96`, then `UpdateProject` `:108`); handler `admin/server/projects.go:610-613`
- **Problem:** The project row and members are committed (`:66`) before provisioning. If `CreateDeployment` (`:96`) or the subsequent `UpdateProject` that sets `PrimaryDeploymentID` (`:108`) fails, `CreateProject` returns the error but the committed project persists with no primary deployment and no rollback. The handler (`server/projects.go:610`) just returns the error with no compensating cleanup. If the project used a freshly-created managed Git repo, that repo is also orphaned (see finding 11); the `+10` quota slack in `quotaManagedRepos` (`admin/github.go:438-444`) explicitly budgets for these leaks.
- **Failure scenario:** A quota check inside `CreateDeployment` or a transient job-scheduling failure aborts creation. The user is left with a half-created project that blocks re-creating the same name and has no deployment; recovery requires manual/admin intervention.
- **Recommendation:** Make creation idempotent/compensating (the file's own TODO at `:21`): on downstream failure, tear down the deployment and delete the project, or move the whole flow to a retryable reconcile job keyed on project ID. Add a GC job for unreferenced managed repos.

**Verification:** CONFIRMED — `admin/projects.go`: the project row + members are committed at `:66`, before `CreateDeployment` (`:96`) and the `UpdateProject` that sets `PrimaryDeploymentID` (`:108-124`). If either fails, `CreateProject` returns the error with no rollback of the committed project, leaving it with no primary deployment. The file's own TODO at `:21` acknowledges the lack of fault tolerance.

**Manual Reproduction Steps (for QA):**
1. Requires forcing provisioning to fail after the project row commits; not directly triggerable from the UI.
2. Staging repro: set the org's deployment/slot quota so that provisioning a new deployment will fail, or make the runtime provisioner temporarily unreachable.
3. Attempt to create + deploy a project (`rill deploy` / the create-project UI flow).
4. Observe the request errors out, then list projects (`rill project list`): the half-created project exists but has no running deployment, and re-creating a project with the same name is blocked by the name collision.
5. Confirm in code: `admin/projects.go:66` (commit) precedes `:96`/`:108` with no compensating cleanup on error.

### admin-projects-deploy-05 [LOW] `GetRepoMeta` dereferences a nil `depl` for non-deployment callers
- **Category:** bug  **Confidence:** medium  **Location:** `admin/server/repos.go:58-65` (assigns `depl` only when owner is a deployment), `:92-93` (`depl.Branch`, `depl.Editable`)
- **Problem:** `depl` is only set when `claims.OwnerType() == auth.OwnerTypeDeployment`. For a Git-backed project the handler falls through to build the response using `depl.Branch`/`depl.Editable`, so a non-deployment caller reaches the deref with `depl == nil`. The permission gate (`ReadProdStatus || ReadDevStatus`, `:31`) admits ordinary users.
- **Failure scenario:** The only real caller is the runtime's admin repo driver (`runtime/drivers/admin/repo.go:895`), which always uses a deployment token, so this never fires through normal flows. It is reachable only by a project member with `ReadProdStatus` deliberately calling the exposed gateway RPC (`GET /v1/projects/{project_id}/repo/meta`) for a GitHub-connected project, yielding a panic → 500. Downgraded from the prior report's medium severity because no product surface triggers it.
- **Recommendation:** Guard the two field accesses (only populate `GitBranch`/`Editable` when `depl != nil`, or require a deployment claim for this RPC).

**Verification:** CONFIRMED — `admin/server/repos.go:58-65` assigns `depl` only when `claims.OwnerType() == auth.OwnerTypeDeployment`; for a GitHub-connected project the handler then builds the response with `depl.Branch`/`depl.Editable` at `:92-93`, so a non-deployment caller reaches a nil dereference. The permission gate at `:31` (`ReadProdStatus || ReadDevStatus`) admits ordinary members.

**Manual Reproduction Steps (for QA):**
1. Connect a project to GitHub (so it has `GitRemote`/`GithubInstallationID` and no archive asset).
2. As a normal project member (with read access), call the gateway RPC directly: `GET /v1/projects/{project_id}/repo/meta` using your user auth token (not a deployment token).
3. Observe an internal-error (HTTP 500) response caused by the nil `depl` dereference.
4. Note this does not fire in normal product flows because the only real caller (the runtime admin repo driver) always uses a deployment token.

### admin-projects-deploy-09 [LOW] `GetBookmark` lets any project reader fetch another user's private bookmark
- **Category:** security  **Confidence:** high  **Location:** `admin/server/bookmarks.go:39-70`
- **Problem:** `GetBookmark` checks only `permissions.ReadProject` and never verifies `bookmark.UserID == claims.OwnerID()` for non-shared/non-default bookmarks. `UpdateBookmark` (`:158`) and `RemoveBookmark` (`:202`) both enforce that ownership check, and `ListBookmarks` (`:23`) filters by owner — so this is an inconsistency, not a deliberate exception.
- **Failure scenario:** User A creates a private bookmark (personal dashboard state/filters). User B, a member of the same project, calls `GetBookmark` with A's bookmark ID (a UUID) and reads A's private view state. Cross-user read, UUID-gated, non-sensitive payload.
- **Recommendation:** Mirror the ownership check from `UpdateBookmark`: reject when the bookmark is neither shared nor default and `bookmark.UserID != claims.OwnerID()`.

**Verification:** CONFIRMED — `admin/server/bookmarks.go:39-70` (`GetBookmark`) checks only `permissions.ReadProject` and never compares `bookmark.UserID` to `claims.OwnerID()`, whereas `UpdateBookmark` and `RemoveBookmark` both enforce `if !bookmark.Shared && !bookmark.Default && bookmark.UserID != claims.OwnerID()` and `ListBookmarks` (`:23`) filters by owner.

**Manual Reproduction Steps (for QA):**
1. As User A, open a dashboard in a shared project and create a private (non-shared, non-default) bookmark; capture its bookmark ID.
2. As User B (a different member of the same project), call `GET` bookmark by ID — the `GetBookmark` RPC (e.g. via the API `GET /v1/bookmarks/{bookmark_id}`) with User A's bookmark ID.
3. Observe User B receives User A's private bookmark (its saved filters/view state) instead of a permission-denied error.
4. Control check: attempt the same ID with `RemoveBookmark`/`UpdateBookmark` as User B — those correctly return PermissionDenied.

### admin-projects-deploy-11 [LOW] Managed GitHub repo is orphaned when its DB row insert fails
- **Category:** correctness  **Confidence:** high  **Location:** `admin/github.go:263-274`
- **Problem:** `CreateManagedGitRepo` creates the repo on GitHub (`:263`), then inserts the `managed_git_repos` row (`:267`). If the insert fails, the GitHub repo is left behind with no tracking row and is never cleaned up. The `quotaManagedRepos` `+10` buffer (`:438-444`) is an explicit workaround, but repos accumulate against the org's real GitHub org.
- **Failure scenario:** A DB hiccup after `CreateManagedRepo` returns leaves a private repo in the managed GitHub org that Rill no longer references; over time these accumulate.
- **Recommendation:** On insert failure, best-effort delete the just-created repo (`DeleteManagedRepo` already exists), and/or add a reconciliation job that deletes managed-org repos with no corresponding DB row.

**Verification:** CONFIRMED — `admin/github.go:263-274` (`CreateManagedGitRepo`) creates the GitHub repo (`:263`) then inserts the `managed_git_repos` row (`:267`); on insert failure it returns the error with no cleanup of the just-created repo, and `quotaManagedRepos` (`:438-444`) budgets `+10` slack for exactly these leaks.

**Manual Reproduction Steps (for QA):**
1. Not user-reproducible through normal flows; requires a DB failure between the GitHub create and the row insert (fault injection).
2. Staging repro: deploy a project using a managed Git repo (the "Rill-managed" GitHub option) and induce a DB error on `InsertManagedGitRepo` (e.g. temporarily break the DB connection right after repo creation).
3. Inspect the managed GitHub org: a new private repo named `<org>-<name>` exists with no corresponding `managed_git_repos` row, and nothing ever deletes it.
4. Confirm in code: `admin/github.go:263-274` — no compensating `DeleteManagedRepo` on the insert-error path.

### admin-projects-deploy-10 [LOW] `githubAuthCallback` is missing a `return` after the non-collaborator redirect
- **Category:** bug  **Confidence:** high  **Location:** `admin/server/github.go:885-889`
- **Problem:** When `!ok` (user is not a collaborator), the handler writes a redirect to the retry page but does not `return`. Execution continues to `sess.Save` (`:892`) and a second `http.Redirect` (`:898-902`). The equivalent guard in `githubConnectCallback` (`:711-716`) correctly returns.
- **Failure scenario:** Corrected from the prior report: because `http.Redirect` at `:888` calls `WriteHeader(307)` and writes a body, the response headers (`Location: <retry page>`) are committed before the fallthrough runs, so the browser *does* land on the retry page — the user is **not** routed past the gate. The observable defects are that the subsequent `sess.Save` cannot persist the cleared session fields (headers already sent) and Go logs a "superfluous response.WriteHeader call". Still a real bug worth the one-line fix.
- **Recommendation:** Add `return` after the retry redirect, matching `githubConnectCallback`.

**Verification:** CONFIRMED — `admin/server/github.go:885-889`: the `if !ok` (non-collaborator) branch writes the retry redirect but has no `return`, so execution falls through to `sess.Save` (`:892`) and a second `http.Redirect` (`:898-902`). The equivalent guard in `githubConnectCallback` (`:711-716`) correctly returns. As noted, the first `http.Redirect` commits headers so the browser still lands on the retry page; the real defects are a non-persisted session save and a "superfluous WriteHeader" log line.

**Manual Reproduction Steps (for QA):**
1. Start the GitHub user-auth flow (`/github/auth/login?remote=<repo>`) as a user who is NOT a collaborator on the target GitHub repo.
2. Complete the GitHub OAuth authorization.
3. Observe the browser lands on the retry page (correct end state) — the gate is not bypassed.
4. Confirm the underlying bug in server logs: look for a "http: superfluous response.WriteHeader call" warning emitted for this request, indicating the fallthrough.
5. Confirm in code: `admin/server/github.go:889` — missing `return` after the redirect, versus `:715` in `githubConnectCallback`.

### admin-projects-deploy-12 [LOW] `assetHandler` commits `200 OK` before the object read, so a fetch failure is reported as success
- **Category:** bug  **Confidence:** high  **Location:** `admin/server/assets.go:227` (`WriteHeader(http.StatusOK)`) vs `:230-236` (object open + error return)
- **Problem:** The handler calls `w.WriteHeader(http.StatusOK)` at `:227`, then opens the GCS object at `:230`. If `NewReader` fails, it returns `httputil.Error(500, ...)` (or 408), but `httputil.WriteError` (`runtime/pkg/httputil/httputil.go:53-70`) then calls `w.WriteHeader(code)` on an already-committed response — a superfluous, ignored write. The client receives `HTTP 200` with the asset's declared `Content-Type` (set at `:213/:219`) and a JSON error body, i.e. a corrupt/empty download reported as success.
- **Failure scenario:** A transient GCS error (or a deleted/misplaced object) on an otherwise-authorized download returns 200 with a mislabeled error body; a client streaming a project archive or image treats it as a successful, valid file.
- **Recommendation:** Move `WriteHeader(http.StatusOK)` to after `NewReader` succeeds (open the object first, then set status and stream). Errors during the subsequent `io.Copy` still cannot change the status, but the pre-stream open failure can and should surface correctly.

**Verification:** CONFIRMED — `admin/server/assets.go:227` calls `w.WriteHeader(http.StatusOK)` before the object is opened at `:230` (`NewReader`). If `NewReader` fails, the returned `httputil.Error(500, ...)` cannot change the already-committed 200 status, so the client sees HTTP 200 with the asset's declared Content-Type and a JSON error body.

**Manual Reproduction Steps (for QA):**
1. Requires the GCS object to be missing/unreadable for an otherwise-authorized asset; fault injection in staging.
2. Create/authorize an asset whose DB row exists but delete the backing object from the storage bucket (or point `asset.Path` at a non-existent object).
3. Request `GET /v1/assets/<asset-id>/download` as an authorized caller.
4. Observe the response status is `200 OK` (with the asset's Content-Type) but the body is a JSON error, i.e. a corrupt/empty download reported as success.
5. Confirm in code: `admin/server/assets.go:227` `WriteHeader(200)` precedes the `NewReader` open at `:230`.

## Product Improvements
- **Provisioning failures are invisible to users.** When `CreateProject`/deployment provisioning fails partway (finding 08), the user gets a raw error and a half-created project with no in-product recovery. Surfacing deployment status/last-error and a "retry/repair" action would remove a class of support tickets.
- **No safeguard around editable-deployment teardown data loss.** Editable deployments are torn down with a best-effort checkpoint (finding 01). Users editing dashboards in the browser have no "unsaved changes" indicator or confirmation before teardown/expiry, and no signal when a final checkpoint fails.
- **Custom-domain changes should reconcile immediately and verifiably** (finding 03) — today even the intended propagation is dead code, and there is no status showing whether deployments have picked up the new domain.

## New Feature Ideas
- **Idempotent project-provisioning saga (L).** Replace the inline `CreateProject`/`RedeployProject` orchestration with a retryable, compensating background job keyed on project ID (the file's own TODO). Eliminates orphaned projects/repos and makes creation resumable.
- **Garbage-collector for orphaned managed repos & assets (M).** Periodic job that deletes managed-org GitHub repos and uploaded assets with no referencing DB row / use case. Directly retires the `+10` quota hack (findings 08, 11) and the asset "background delete" TODO.
- **Job-queue-based webhook ingestion (M).** Accept the GitHub webhook, persist the event, ack 200 immediately, and process reconcile fan-out asynchronously with bounded concurrency and retries. Fixes the 10s-timeout/retry risk (finding 07) and scales to popular shared remotes.
- **Unsaved-changes guard for editable deployments (M).** Track whether the runtime repo has un-pushed changes and (a) block/confirm teardown, (b) show an indicator in the UI, (c) refuse to delete the branch if the final checkpoint failed. Prevents the data loss in finding 01.
- **Scope the public-asset fallback to branding types (S).** Restrict the "org has public projects" download path to image asset types so private `deploy` archives can never be served through it (finding 02); small change, meaningful hardening.
- **Deployment status & provisioner-error surfacing in the UI (M).** Expose `DeploymentStatus`/`StatusMessage` and provisioner-resource errors in the project settings view so users can self-diagnose stuck/errored deployments.

## Notes & Missing Test Coverage
- **Redirect-domain handling is safe** despite `WithCustomDomainFromRedirectURL` (`admin/urls.go:116`) trusting an arbitrary host: the only caller (`admin/server/auth/handlers.go:321`) operates on a redirect validated by `validateRedirectURL` at login start (which requires `IsSafeRedirectURL` or a known custom domain). No change needed, but the coupling is implicit and worth a comment/test.
- **`UpdateDeploymentsForProject` (`admin/deployments.go:139-175`)** writes `prodErr` from goroutines; this is safe (not a data race) because deployment IDs are unique, so at most one goroutine matches `*p.PrimaryDeploymentID == d.ID` and writes, and the read follows `grp.Wait()`. A test pinning the single-writer intent would help.
- **`GetVirtualFile`/`DeleteVirtualFile` (`admin/server/repos.go:162-241`)** take `req.Environment` from the client and gate only on `ReadProdStatus || ReadDevStatus`, with no per-file owner check (unlike `GetPersonalFile`). Paths carry a random UUID suffix so this is not readily exploitable, but a test/tightening around cross-environment and cross-owner access would be prudent.
- **No tests observed** for: nil `PrimaryDeploymentID` paths in report/personal-file handlers (finding 04); `LookupCanvas` state handling (finding 06); custom-domain propagation (finding 03); partial-failure rollback in `CreateProject` (finding 08); and editable-deployment delete when the checkpoint push fails (finding 01). These are exactly the risky branches and the highest-value additions.

## Verification Notes
Every one of the 11 original findings was re-checked against the code; all describe behavior that is actually present, so none were removed (`removed_count = 0`). Corrections applied:
- **Finding 01 (delete data-loss): HIGH → MEDIUM.** The behavior is confirmed, but the prior report's headline TTL scenario is imprecise: `hibernate_expired_deployments.go` *stops* (checkpoints + preserves the PVC) before it *deletes*, so the TTL path already attempts a checkpoint and preserves data on the first pass. The genuine loss window is a **direct** teardown of a running editable deployment (`DeleteDeployment` RPC / `TeardownProject`). Reworded the scenario accordingly.
- **Finding 05 (GetRepoMeta nil `depl`): MEDIUM → LOW.** Traced callers: the only real caller is the runtime with a deployment token, so no product surface triggers the nil deref; it is reachable only by a deliberate direct gateway call. Severity lowered, confidence kept medium.
- **Finding 10 (missing `return`): scenario corrected.** The prior claim that the user is "routed past the retry gate" is wrong — the first `http.Redirect` commits the response headers, so the browser lands on the retry page. The real effects are a non-persisted session save and a superfluous-`WriteHeader` log. Kept LOW.
- **Findings 02, 03, 06, 07, 08, 09, 11: confirmed** with locations/line numbers checked (sibling guards for 04 verified at `alerts.go:147/211/283/429/503`, `reports.go:194/255`; `quotaManagedRepos` `+10` verified at `github.go:438-444`; archive `Public` default-false verified at `archive.go:39-45`). Minor wording tightened on 02 (UUID-gated defense-in-depth) and 03 (superuser-only, self-healing).
- **Added finding 12 (assetHandler 200-before-read):** a real correctness bug in scope that the prior pass missed — `WriteHeader(200)` at `assets.go:227` precedes the object open at `:230`, so a fetch failure is delivered as HTTP 200.
- Findings are re-ordered most-significant first. No critical/high items survive verification; health remains **fair**.
