<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { isActiveDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
  import { useParserCommitSha } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { isMergeConflictError } from "@rilldata/web-common/features/project/deploy/github-utils.ts";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitPullMutation,
    createRuntimeServiceGitPushMutation,
    getRuntimeServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { ConnectError } from "@connectrpc/connect";
  import { Rocket } from "lucide-svelte";
  import { buildPostMergeUrl } from "./post-merge-url";
  import { goto } from "$app/navigation";
  import {
    fetchDeploymentGithubStatusChanges,
    getDeploymentGithubStatus,
  } from "@rilldata/web-admin/features/edit-session/selectors.ts";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined;

  const AUTO_COMMIT_MESSAGE = "Updates from cloud editor";

  type PublishSnapshots = {
    hadProdDeployment: boolean;
    needsRedeploy: boolean;
    preCommitSha: string | undefined;
  };

  let isPublishing = false;
  let remoteChangeDialog = false;
  let mergeConflictResolutionDialog = false;
  // Which mutation triggered the conflict dialog; drives the dispatch in
  // `handleForceFetchRemoteCommits`. `null` when the dialog isn't open.
  let conflictSource: "pull" | "merge" | null = null;
  // Captured at click time so the publish flow can resume after a force
  // merge without re-reading state that may have changed.
  let pendingPublishSnapshots: PublishSnapshots | null = null;
  let errorFromGitCommand: ConnectError | null = null;

  const client = useRuntimeClient();
  const gitPushMutation = createRuntimeServiceGitPushMutation(client);
  const gitMergeMutation = createRuntimeServiceGitMergeToBranchMutation(client);
  const gitPullMutation = createRuntimeServiceGitPullMutation(client);
  const gitStatusQuery = getDeploymentGithubStatus(client, primaryBranch);
  // Query GetProject without a branch param so `data.deployment` reflects
  // the project's primary (prod) deployment — the same source of truth the
  // project layout uses. ListDeployments is too loose: it includes orphan
  // prod records whose project pointer has been cleared.
  const projectQuery = createAdminServiceGetProject(organization, project);
  const redeployProjectMutation = createAdminServiceRedeployProject();

  $: ({
    isPending,
    data: {
      hasLocalChanges,
      hasChangesOnCurrent,
      hasRemoteChanges,
      hasLocalCommitsOnCurrent,
      alreadyOnPrimary,
      disabledPerGitStatus,
    },
  } = $gitStatusQuery);

  $: ({ isPending: gitPullPending, error: gitPullError } = $gitPullMutation);
  $: dialogError = (gitPullError as ConnectError | null) ?? errorFromGitCommand;

  $: projectLoaded = $projectQuery.data !== undefined;
  $: prodDeployment = $projectQuery.data?.deployment;
  $: prodDeploymentActive =
    !!prodDeployment && isActiveDeployment(prodDeployment);
  $: disabled = !projectLoaded || disabledPerGitStatus || isPublishing;

  // Prefetch prod's project parser commit SHA so the deploying page can
  // wait for prod to advance past it before redirecting to the dashboard,
  // avoiding the stale-content race. Deployment + JWT come straight from
  // `projectQuery` rather than via a dedicated `useProdRuntimeClient`
  // hook: the popover doesn't make other prod-runtime calls, so the
  // wrapper wouldn't earn its place.
  $: parserShaQuery = useParserCommitSha(
    prodDeployment,
    $projectQuery.data?.jwt,
  );

  function invalidateGitStatusQueries() {
    // gitStatus tracks localChanges and currentBranch; mutations below change
    // both, so refresh after every flow. We cache both an empty-remoteBranch
    // and a primary-branch keyed query (see `getDeploymentGithubStatus`), so
    // invalidate both.
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(client.instanceId, {}),
    });
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(client.instanceId, {
        remoteBranch: primaryBranch,
      }),
    });
  }

  async function handlePublish() {
    if (!primaryBranch || isPublishing) return;

    // If the remote has commits we don't have locally, stop and prompt the
    // user to pull first. After a successful pull the user re-clicks Publish.
    if (hasRemoteChanges) {
      errorFromGitCommand = null;
      remoteChangeDialog = true;
      return;
    }

    isPublishing = true;

    // Snapshot the prod deployment state at click time. Three relevant cases:
    //   1. No prod deployment yet → first publish; route to /-/invite.
    //   2. Prod deployment exists but is dormant (STOPPED/ERRORED) → wake it
    //      via RedeployProject; route to /-/deploying.
    //   3. Prod deployment is already active (PENDING/RUNNING/UPDATING) → the
    //      merge alone reconciles changes; route to /-/deploying.
    // RedeployProject (admin/projects.go:333) handles cases 1 and 2 with a
    // single RPC: if there's no current deployment it creates one, otherwise
    // it provisions a fresh one and tears down the old.
    const snapshots: PublishSnapshots = {
      hadProdDeployment: !!prodDeployment,
      needsRedeploy: !prodDeploymentActive,
      preCommitSha: $parserShaQuery.data,
    };

    // Refetch local changes status, we predict this based on file watcher response.
    // But we dont check if changes flipped to with changes to without changes.
    hasLocalChanges = await fetchDeploymentGithubStatusChanges(
      client,
      queryClient,
      primaryBranch,
    );
    if (!hasLocalChanges && !hasChangesOnCurrent) {
      eventBus.emit("notification", {
        type: "default",
        message: "No changes detected",
      });
      isPublishing = false;
      return;
    }

    let mergeResp;
    try {
      if (hasLocalChanges) {
        await $gitPushMutation.mutateAsync({
          commitMessage: AUTO_COMMIT_MESSAGE,
          force: false,
        });
      }
      mergeResp = await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: false,
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(err) || "Failed to publish",
      });
      isPublishing = false;
      return;
    } finally {
      invalidateGitStatusQueries();
    }

    // GitMergeToBranch surfaces conflicts via `output` rather than an error;
    // unhandled, the user would see a silent failure (the merge didn't land
    // but the publish appears to have succeeded). Branch on it explicitly.
    if (mergeResp?.output) {
      if (isMergeConflictError(mergeResp.output)) {
        conflictSource = "merge";
        pendingPublishSnapshots = snapshots;
        errorFromGitCommand = null;
        mergeConflictResolutionDialog = true;
      } else {
        eventBus.emit("notification", {
          type: "error",
          message: mergeResp.output,
        });
      }
      isPublishing = false;
      return;
    }

    await completePublishFlow(snapshots);
    isPublishing = false;
  }

  async function completePublishFlow(snapshots: PublishSnapshots) {
    if (snapshots.needsRedeploy) {
      try {
        // TODO: detect billing/quota errors here and surface a friendly
        // message, mirroring `getPrettyDeployError` in
        // `web-common/src/features/project/deploy/deploy-errors.ts`.
        await $redeployProjectMutation.mutateAsync({
          org: organization,
          project,
        });
        // Refresh the project query so the layout sees the new primary
        // deployment pointer, and ListDeployments so subscribers like
        // BranchSelector and BranchesSection pick up the new row.
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          }),
          queryClient.invalidateQueries({
            queryKey: getAdminServiceListDeploymentsQueryKey(
              organization,
              project,
            ),
          }),
        ]);
      } catch (err) {
        // The merge succeeded but the prod deployment failed to (re)start.
        // Be explicit so the user knows their changes are on the primary
        // branch and that the deployment step is what needs retrying.
        const detail = extractErrorMessage(err);
        eventBus.emit("notification", {
          type: "error",
          message: `Changes merged to production, but starting the production deployment failed${
            detail ? `: ${detail}` : ""
          }.`,
        });
        return;
      }
    }

    const targetUrl = buildPostMergeUrl({
      organization,
      project,
      page: $page,
      hadProdDeployment: snapshots.hadProdDeployment,
      preCommitSha: snapshots.preCommitSha,
    });
    const targetWindow = window.open(targetUrl, "_blank");
    if (!targetWindow) {
      // Go to target directly if popup is blocked.
      void goto(targetUrl);
      eventBus.emit("notification", {
        type: "error",
        message: "Pop-up was blocked.",
      });
    }
  }

  async function handleFetchRemoteCommits() {
    // GitPull can't auto-merge cleanly when there are unpushed local commits;
    // jump straight to the force/keep choice (mirrors RemoteProjectManager).
    if (hasLocalCommitsOnCurrent) {
      conflictSource = "pull";
      mergeConflictResolutionDialog = true;
      return;
    }

    errorFromGitCommand = null;
    const resp = await $gitPullMutation.mutateAsync({ discardLocal: false });
    invalidateGitStatusQueries();

    if (!resp.output) {
      remoteChangeDialog = false;
      eventBus.emit("notification", {
        message: "Remote project changes fetched and merged.",
      });
      return;
    }

    if (isMergeConflictError(resp.output)) {
      conflictSource = "pull";
      mergeConflictResolutionDialog = true;
      return;
    }

    errorFromGitCommand = { message: resp.output } as ConnectError;
  }

  async function handleForceFetchRemoteCommits() {
    errorFromGitCommand = null;

    if (conflictSource === "pull") {
      const resp = await $gitPullMutation.mutateAsync({ discardLocal: true });
      invalidateGitStatusQueries();

      if (resp.output) {
        errorFromGitCommand = { message: resp.output } as ConnectError;
        return;
      }

      remoteChangeDialog = false;
      mergeConflictResolutionDialog = false;
      conflictSource = null;
      eventBus.emit("notification", {
        message:
          "Remote project changes fetched and merged. Your changes have been stashed.",
      });
      return;
    }

    // conflictSource === "merge": force-merge to primary, then resume the
    // publish flow with the snapshots captured at click time.
    isPublishing = true;
    let resp;
    try {
      resp = await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: true,
      });
    } catch (err) {
      errorFromGitCommand = err as ConnectError;
      isPublishing = false;
      return;
    } finally {
      invalidateGitStatusQueries();
    }

    if (resp?.output) {
      errorFromGitCommand = { message: resp.output } as ConnectError;
      isPublishing = false;
      return;
    }

    mergeConflictResolutionDialog = false;
    conflictSource = null;
    const snapshots = pendingPublishSnapshots;
    pendingPublishSnapshots = null;
    if (snapshots) {
      await completePublishFlow(snapshots);
    }
    isPublishing = false;
  }
</script>

<Tooltip distance={8}>
  <Button
    type="primary"
    {disabled}
    loading={isPublishing}
    loadingCopy="Publishing..."
    onClick={handlePublish}
  >
    <Rocket size="14" />
    Publish
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="240px">
    <span class="text-xs">
      {#if alreadyOnPrimary}
        Already on production
      {:else if isPending || !projectLoaded}
        Loading project...
      {:else if !hasLocalChanges}
        No changes to publish
      {:else if hasRemoteChanges}
        Remote has updates not in your session. Click to review.
      {:else if !prodDeployment}
        Publish your project to production. We'll open a new tab where you can
        invite teammates while the deployment reconciles.
      {:else if !prodDeploymentActive}
        Production is hibernated. Publishing will resume it and apply your
        changes. We'll open the deployment in a new tab so you can watch updates
        reconcile.
      {:else}
        Publish your changes to production. We'll open a new tab so you can
        watch updates reconcile.
      {/if}
    </span>
  </TooltipContent>
</Tooltip>

<ProjectContainsRemoteChangesDialog
  bind:open={remoteChangeDialog}
  loading={gitPullPending}
  error={dialogError}
  onFetchAndMerge={handleFetchRemoteCommits}
/>

<MergeConflictResolutionDialog
  bind:open={mergeConflictResolutionDialog}
  loading={gitPullPending || isPublishing}
  error={dialogError}
  onUseLatestVersion={handleForceFetchRemoteCommits}
/>
