<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import {
    createAdminServiceGetProject,
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { isActiveDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
  import {
    getDeploymentGithubStatus,
    invalidateGitStatusQueries,
  } from "@rilldata/web-admin/features/edit-session/selectors.ts";
  import { useParserCommitSha } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitStatus,
    type V1GitMergeToBranchResponse,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { ConnectError } from "@connectrpc/connect";
  import { ExternalLink, GitPullRequest } from "lucide-svelte";
  import ChangedFilesList from "@rilldata/web-common/features/project/ChangedFilesList.svelte";
  import { buildPostMergeUrl } from "./post-merge-url";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined;

  type MergeSnapshots = {
    hadProdDeployment: boolean;
    needsRedeploy: boolean;
    preCommitSha: string | undefined;
  };

  let open = false;
  let isMerging = false;
  let errorMessage = "";
  let mergeConflictDialog = false;
  // Captured at click time so the merge flow can resume after a force merge
  // without re-reading state that may have changed. `preCommitSha` is refreshed
  // before completing the flow because prod's parser may have advanced while
  // the user resolved the conflict.
  let pendingMergeSnapshots: MergeSnapshots | null = null;
  let errorFromGitCommand: ConnectError | null = null;

  const client = useRuntimeClient();
  const gitMergeMutation = createRuntimeServiceGitMergeToBranchMutation(client);
  const gitStatusQuery = getDeploymentGithubStatus(client, primaryBranch);
  // Raw current-branch status drives the branch name and GitHub link shown in
  // the popover; the derived booleans come from `getDeploymentGithubStatus`.
  const currentBranchStatusQuery = createRuntimeServiceGitStatus(client, {});
  // Query GetProject without a branch param so `data.deployment` reflects
  // the project's primary (prod) deployment. Self-managed projects can lack
  // a prod deployment when created via `rill project connect-github
  // --skip-deploy` (see `cli/cmd/project/connect_github.go`), and can also
  // sit dormant after hibernation; we mirror PublishPopover's three-state
  // logic and route accordingly.
  const projectQuery = createAdminServiceGetProject(organization, project);
  const redeployProjectMutation = createAdminServiceRedeployProject();

  $: ({
    isPending,
    data: {
      hasLocalChanges,
      hasRemoteChanges,
      alreadyOnPrimary,
      disabledPerGitStatus,
    },
  } = $gitStatusQuery);

  $: currentBranch = $currentBranchStatusQuery.data?.branch ?? "";
  $: branchUrl =
    $currentBranchStatusQuery.data?.githubUrl && currentBranch
      ? `${getGitUrlFromRemote($currentBranchStatusQuery.data.githubUrl)}/tree/${encodeURIComponent(currentBranch)}`
      : "";
  $: projectLoaded = $projectQuery.data !== undefined;
  $: prodDeployment = $projectQuery.data?.deployment;
  $: prodDeploymentActive =
    !!prodDeployment && isActiveDeployment(prodDeployment);
  $: disabled = !projectLoaded || disabledPerGitStatus || isMerging;

  // Prefetch prod's project parser commit SHA so the deploying page can
  // wait for prod to advance past it before redirecting (see
  // `PublishPopover` for the same pattern, including why we read
  // deployment + JWT directly from `projectQuery`).
  $: parserShaQuery = useParserCommitSha(
    prodDeployment,
    $projectQuery.data?.jwt,
  );

  $: if (!open) {
    errorMessage = "";
  }

  async function handleMerge() {
    if (!primaryBranch || isMerging) return;

    // If the remote has commits we don't have locally, stop and ask the
    // shared dialog (owned by CloudRemoteChangeManager) to open via the bus.
    // After a successful pull the user re-clicks Merge.
    if (hasRemoteChanges) {
      eventBus.emit("remote-changes-detected");
      open = false;
      return;
    }

    isMerging = true;
    errorMessage = "";

    // Snapshot the prod deployment state at click time. Same three cases as
    // PublishPopover (see comment there); RedeployProject covers both the
    // no-deployment and dormant-deployment paths with one RPC.
    const snapshots: MergeSnapshots = {
      hadProdDeployment: !!prodDeployment,
      needsRedeploy: !prodDeploymentActive,
      preCommitSha: $parserShaQuery.data,
    };

    let mergeResp: V1GitMergeToBranchResponse | undefined = undefined;
    try {
      mergeResp = await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: false,
      });
    } catch (err) {
      errorMessage = extractErrorMessage(err) || "Failed to merge";
      isMerging = false;
      return;
    } finally {
      invalidateGitStatusQueries(client, primaryBranch);
    }

    // GitMergeToBranch surfaces conflicts via `output` rather than an error;
    // unhandled, the user would see a silent failure (the merge didn't land
    // but the merge appears to have succeeded). Branch on it explicitly.
    if (mergeResp?.output) {
      if (mergeResp?.conflict) {
        pendingMergeSnapshots = snapshots;
        errorFromGitCommand = null;
        mergeConflictDialog = true;
        open = false; // only close when opening merge conflict dialog
      } else {
        errorMessage = mergeResp.output;
      }
      isMerging = false;
      return;
    }

    await completeMergeFlow(snapshots);
    isMerging = false;
    open = false;
  }

  async function completeMergeFlow(snapshots: MergeSnapshots) {
    if (snapshots.needsRedeploy) {
      try {
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
        errorMessage = `Changes merged to production, but starting the production deployment failed${
          detail ? `: ${detail}` : ""
        }.`;
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

  async function handleForceMerge() {
    errorFromGitCommand = null;
    isMerging = true;
    let resp;
    try {
      resp = await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: true,
      });
    } catch (err) {
      errorFromGitCommand = err as ConnectError;
      isMerging = false;
      return;
    } finally {
      invalidateGitStatusQueries(client, primaryBranch);
    }

    if (resp?.output) {
      errorFromGitCommand = { message: resp.output } as ConnectError;
      isMerging = false;
      return;
    }

    mergeConflictDialog = false;
    const snapshots = pendingMergeSnapshots;
    pendingMergeSnapshots = null;
    if (snapshots) {
      // Prod's parser may have advanced while the user was on the conflict
      // dialog; re-read so the deploying page waits past the correct SHA.
      snapshots.preCommitSha = $parserShaQuery.data;
      await completeMergeFlow(snapshots);
    }
    isMerging = false;
    open = false;
  }
</script>

<Tooltip distance={8} suppress={open}>
  <Popover.Root bind:open>
    <Popover.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="primary" {disabled}>
          <GitPullRequest size="14" />
          Merge to production
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content align="end" class="!w-[320px]">
      <div class="flex flex-col gap-y-3">
        <p class="text-xs text-fg-secondary">
          {#if !prodDeployment}
            Merging
            <span class="font-semibold text-fg-primary">"{currentBranch}"</span>
            sets up your production deployment. We'll open a new tab where you can
            invite teammates while it reconciles.
          {:else if !prodDeploymentActive}
            Production is hibernated. Merging
            <span class="font-semibold text-fg-primary">"{currentBranch}"</span>
            will resume it and apply your changes. We'll open the deployment in a
            new tab so you can watch updates reconcile.
          {:else}
            Merging pushes changes from
            <span class="font-semibold text-fg-primary">"{currentBranch}"</span>
            to production. We'll open a new tab so you can watch updates reconcile.
          {/if}
        </p>
        <ChangedFilesList remoteBranch={primaryBranch} {open} />
        {#if branchUrl}
          <a
            class="github-link"
            href={branchUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            View branch on GitHub
            <ExternalLink size="11" />
          </a>
        {/if}
        <Button
          type="primary"
          small
          disabled={isMerging}
          loading={isMerging}
          loadingCopy="Merging..."
          onClick={handleMerge}
        >
          Merge
        </Button>
        {#if errorMessage}
          <p class="text-xs text-red-600">{errorMessage}</p>
        {/if}
      </div>
    </Popover.Content>
  </Popover.Root>
  <TooltipContent slot="tooltip-content" maxWidth="220px">
    <span class="text-xs">
      {#if alreadyOnPrimary}
        Already on production
      {:else if isPending || !projectLoaded}
        Loading project...
      {:else if !hasLocalChanges}
        No changes to merge
      {:else if hasRemoteChanges}
        Remote has updates not in your session. Click to review.
      {:else}
        Review and confirm before merging
      {/if}
    </span>
  </TooltipContent>
</Tooltip>

<MergeConflictResolutionDialog
  bind:open={mergeConflictDialog}
  loading={isMerging}
  error={errorFromGitCommand}
  onUseLatestVersion={handleForceMerge}
/>

<style lang="postcss">
  .github-link {
    @apply inline-flex items-center gap-x-1 text-xs text-fg-secondary;
    @apply hover:text-fg-primary hover:underline;
  }
</style>
