<script lang="ts">
  import type { ConnectError } from "@connectrpc/connect";
  import { isMergeConflictError } from "@rilldata/web-common/features/project/deploy/github-utils.ts";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    createRuntimeServiceGitPullMutation,
    getRuntimeServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getDeploymentGithubStatus } from "./selectors.ts";

  export let primaryBranch: string | undefined;

  const runtimeClient = useRuntimeClient();
  const gitStatusQuery = getDeploymentGithubStatus(
    runtimeClient,
    primaryBranch,
  );
  const gitPullMutation = createRuntimeServiceGitPullMutation(runtimeClient);

  let remoteChangeDialog = false;
  let mergeConflictResolutionDialog = false;
  let errorFromGitCommand: ConnectError | null = null;

  // Debounce the open/close sync. The two underlying GitStatus queries
  // refetch independently after each mutation, so the derived store can
  // briefly re-emit the prior `hasRemoteChanges: true` while one refetch is
  // in flight, then settle to `false` once both land — without debouncing,
  // the user sees the dialog flicker back open and immediately close after a
  // pull. The debouncer applies the latest value once emissions go quiet.
  const syncRemoteChangeDialog = debounce((value: boolean) => {
    remoteChangeDialog = value;
  }, 500);

  $: if ($gitStatusQuery.data) {
    syncRemoteChangeDialog($gitStatusQuery.data.hasRemoteChanges);
  }

  $: hasLocalCommitsOnCurrent =
    $gitStatusQuery.data?.hasLocalCommitsOnCurrent ?? false;
  $: ({ isPending: gitPullPending, error: gitPullError } = $gitPullMutation);
  $: dialogError = (gitPullError as ConnectError | null) ?? errorFromGitCommand;

  function invalidateGitStatusQueries() {
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(
        runtimeClient.instanceId,
        {},
      ),
    });
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {
        remoteBranch: primaryBranch,
      }),
    });
  }

  async function handleFetchRemoteCommits() {
    // GitPull can't auto-merge cleanly when there are unpushed local commits;
    // jump straight to the force/keep choice (mirrors RemoteProjectManager).
    if (hasLocalCommitsOnCurrent) {
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
      mergeConflictResolutionDialog = true;
      return;
    }

    errorFromGitCommand = { message: resp.output } as ConnectError;
  }

  async function handleForceFetchRemoteCommits() {
    errorFromGitCommand = null;
    const resp = await $gitPullMutation.mutateAsync({ discardLocal: true });
    invalidateGitStatusQueries();

    if (resp.output) {
      errorFromGitCommand = { message: resp.output } as ConnectError;
      return;
    }

    remoteChangeDialog = false;
    mergeConflictResolutionDialog = false;
    eventBus.emit("notification", {
      message:
        "Remote project changes fetched and merged. Your changes have been stashed.",
    });
  }
</script>

<ProjectContainsRemoteChangesDialog
  bind:open={remoteChangeDialog}
  loading={gitPullPending}
  error={dialogError}
  onFetchAndMerge={handleFetchRemoteCommits}
/>

<MergeConflictResolutionDialog
  bind:open={mergeConflictResolutionDialog}
  loading={gitPullPending}
  error={dialogError}
  onUseLatestVersion={handleForceFetchRemoteCommits}
/>
