# Merge-to-production: post-merge UX

## Context

The `Merge to production` button was added in PR #9339 to the cloud edit-session header (`web-admin/src/features/edit-session/MergePopover.svelte`). The runtime-side merge handler was implemented in PR #9333 (`runtime/drivers/admin/repo.go` → `MergeToBranch`). Today the button calls `GitMergeToBranch`, fires a "Changes merged" toast, and leaves the user sitting in the editor — with no signal that anything is happening on production.

Once Platform implements the merge, the actual chain is asynchronous: admin merges dev → primary, GitHub fires a webhook to the admin server (`admin/server/github.go:375`), admin calls `TriggerParser` on the prod deployment (`admin/github.go:342-391`), the prod runtime pulls and reconciles. During that window prod keeps serving the last-valid state.

Two facts shape the design:

1. **Reconciliation is per-resource, not commit-atomic.** Each resource reconciles in its own goroutine (`runtime/controller.go:1209-1298`); a model's `ResultTable` is overwritten the moment its build finishes (`runtime/reconcilers/model.go:740-742`). Downstream resources only pick up the new parent state on their next reconcile. There is no project-wide "snapshot" or "promote" step. **Viewers can see mixed state** mid-window: dashboard A may already show new content while dashboard B still shows old.
2. **Reconciles can be long.** A heavy ingest can run 30 minutes to 1+ hour. The publisher's UI cannot block, time out, or claim "live now" before everything has finished.

This plan defines the frontend behavior across that window and the platform-side signals needed to drive it.

## Recommended UX

Three confirmed design choices (from earlier in the conversation):

1. **Redirect immediately on merge success.** The user leaves the editor as soon as the merge RPC returns and lands on `/{org}/{project}`.
2. **Silent reconciliation for viewers.** No project-wide banner, no per-dashboard staleness chip. Only the publisher sees an in-flight indicator (we know who they are because they were redirected from the editor).
3. **On reconcile failure, redirect to prod with a failure banner.** Prod is unaffected (still serving last-valid for the resources whose reconcile failed). Banner explains the deploy had errors and links to the project status page and back into the editor.

### Banner contract

The publisher gets a non-blocking, non-dismissible-until-resolved status bar at the top of the project layout. **It does not time out.** It persists across navigation within the project (sessionStorage-keyed) so the user can browse dashboards while the deploy runs. The user can collapse it; they cannot make it lie.

| Phase | Trigger | Banner copy |
| --- | --- | --- |
| Publishing | Redirect arrives, prod deployment's `current_git_commit` ≠ merged SHA | "Publishing your changes…" |
| Deploying | `current_git_commit` matches merged SHA, but resources still PENDING/RUNNING | "Deploying — N of M resources updated" (progress count from `WatchResources`) |
| Complete | All resources IDLE, no `reconcileError`s | "Deploy complete" — auto-dismiss after ~3s |
| Partial failure | All resources IDLE, ≥1 has `reconcileError` | Red: "Deploy finished with N errors. Production is serving the previous version of those resources." Links: View errors → `/-/status`; Back to editor |
| Parser failure | `ProjectParser.reconcileError` non-empty after commit picked up | Red: "Deploy failed — couldn't parse project files." Same links |

We deliberately do **not** say "live now" mid-flight, because per-resource atomicity means it's never globally true until everything is IDLE. The `N of M` progress is honest and gives the user a feel for whether to wait or come back later.

The banner is gated on a `?published=<sha>` query param + matching sessionStorage entry — set by the editor at redirect time. Other viewers visiting the project never see it.

## Frontend changes (`web-admin`)

Files to touch:

- `web-admin/src/features/edit-session/MergePopover.svelte` — before merging, fetch prod's current `ProjectParserState.current_commit_sha` (call it `prevProdSha`). On merge success, persist `{prevProdSha, mergedAt, expectedSha?}` to sessionStorage keyed by `{org}/{project}`, then `goto(`/${org}/${project}?published=1`)`. Drop the success toast — the destination banner replaces it. (`expectedSha` is the post-merge SHA returned by the merge RPC if Platform ask #1 lands; otherwise it's omitted and the banner falls back to "any new SHA on prod = the merge.")
- `web-admin/src/features/projects/PublishingBanner.svelte` (new) — mounts in the project root layout. Reads `?published=<sha>` and the matching sessionStorage entry. Renders the five phases above. User can collapse but not dismiss while in `Publishing` / `Deploying` / `Errored` phases. On `Complete`, auto-dismisses and clears sessionStorage.
- `web-admin/src/features/projects/status/selectors.ts` — add `usePublishedDeployStatus(client, prevProdSha, expectedSha?)` that returns `{ phase, total, completed, errors }`. Composes:
  - `GetResource(kind=ProjectParser, name=parser)` against the prod runtime → `state.current_commit_sha` (already exists, `proto/rill/runtime/v1/resources.proto:74`). Compare to `prevProdSha` to detect "the new commit was picked up." If `expectedSha` is also threaded through (Platform ask #1), use it as the strict equality check; otherwise fall back to `current_commit_sha !== prevProdSha`.
  - `WatchResources` / `ListResources` for progress and errors. Reuse helpers in `web-admin/src/features/dashboards/listing/deploying-dashboards.ts:84-121` (`isResourceReconciling`, `hasErrored`).
  - `useParserReconcileError` from `selectors.ts:47` for parser-level failure path.
- Optional: collapse the banner into a small status pill in the header when minimized, to keep it out of the way during long deploys.

Viewers' surfaces are unchanged — the banner is only visible to the publisher.

## Platform team asks

None of these are hard blockers — the UX can ship without any platform changes by relying on a prev-vs-current SHA diff against `ProjectParserState.current_commit_sha` (`proto/rill/runtime/v1/resources.proto:74`), which already exists. The asks below are tightenings:

1. **Return `commit_sha` from `GitMergeToBranch`.** `GitMergeToBranchResponse` (`proto/rill/runtime/v1/api.proto:1337`) currently exposes only `output` (populated on conflict). Add `string commit_sha = 2` so the frontend can watch for an exact SHA. This matches the precedent set by `RestoreGitCommitResponse.new_commit_sha` (line 1326). With it, the banner is precise; without it, we fall back to prev-vs-current diff (works in the common case, see race condition below).
2. **Tighten the post-webhook-failure fallback.** `runtime/config_reloader.go:202` polls hourly. If the GitHub webhook drops, the publisher's banner sits in `Publishing` for up to an hour. Either: (a) when admin processes the merge, also call `ReloadConfig` directly on the prod runtime as a belt-and-braces trigger; or (b) shorten the fallback poll for the first ~5 min after a known-recent merge.
3. **Per-deploy reconcile error attribution (nice-to-have).** Today we infer "this commit caused the error" from "this resource is errored after we picked up the new commit." A pre-existing error muddies that. A `last_reconcile_started_at` per resource (or scoping resource errors to the commit they were observed under) would let the banner cleanly say "your commit caused these errors" vs. "your commit deployed cleanly but other resources were already failing." If this is heavy, the frontend can live without it.

**Things we don't need from Platform** (initially considered, then verified): a `current_git_commit` field on `Deployment` is unnecessary — `ProjectParserState.current_commit_sha` already exposes exactly this on the runtime side, and the publisher's redirected session already connects to the prod runtime via the project layout.

### Race condition to acknowledge (only if #1 is skipped)

If Alice clicks publish, gets redirected, and Bob clicks publish before Alice's reconcile finishes, the banner watches "prod's `current_commit_sha` changed from `prevProdSha`." Bob's commit might land first, flipping Alice's banner to "Deploying"/"Complete" against Bob's changes rather than Alice's. The deploy itself is correct — whatever ends up on prod's primary branch is what gets reconciled — but Alice's banner could prematurely declare success before her changes are merged in. Adding `commit_sha` to the merge response avoids this.

### Out of scope (but worth noting)

A true commit-atomic cutover ("blue/green for the whole project") would let the banner honestly say "live now" at a single moment and would prevent viewers from seeing mixed state. That's a much larger platform change — separate database/state generation per commit, dual writes, atomic promote. Calling it out so it's recorded, but not asking for it in this plan.

## Verification

End-to-end:

1. `rill devtool start cloud`. Open a project, click `Edit`, make a trivial change in the dev branch (e.g., rename a dashboard title), commit.
2. Click `Merge to production`. Confirm popover, click `Merge`.
3. Expect: redirect to `/{org}/{project}?published=<sha>`. Banner shows "Publishing your changes…". Dashboards still render the pre-merge title.
4. Once webhook fires and prod runtime picks up the commit: banner switches to "Deploying — N of M resources updated"; counter increments as each resource reconciles.
5. When all IDLE: banner shows "Deploy complete" briefly, then dismisses; affected dashboards now render the new title.
6. **Long-reconcile case:** introduce a model that takes ≥3 minutes to materialize (synthetic `pg_sleep`-style or large generated dataset). Expect: banner persists at "Deploying — N of M…" for the full duration without timing out; user can navigate to other dashboards on the project, banner follows them; "Deploy complete" only fires when the slow model finishes.
7. **Failure case:** introduce a deliberate YAML error on the dev branch before merging. Repeat. Expect: red banner surfaces the parser error, "Back to editor" returns to edit-session on the dev branch with the change still present.
8. **Per-resource failure:** introduce a SQL error in one of three modified models. Expect: banner reaches `Complete` for the two clean resources, then settles into red "Deploy finished with 1 error" once everything is IDLE.
9. **Webhook-drop case:** block the webhook in devtool, merge. Expect: banner stays in "Publishing your changes…" indefinitely (no timeout) until either the fallback trigger fires or the user navigates away.

Unit tests for `usePublishedDeployStatus`: mock the deployment + resource queries to cover Publishing / Deploying / Complete / partial-failure / parser-failure phases.
