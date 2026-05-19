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
    createRuntimeServiceGitPushMutation,
    getRuntimeServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy } from "svelte";
  import type { Readable } from "svelte/store";

  export let gitStatusSource: Readable<{
    hasRemoteChanges: boolean;
    hasLocalCommitsOnCurrent: boolean;
  }>;
  export let primaryBranch: string | undefined = undefined;
  export let autoPush = false;
  export let autoOpen = false;
  export let debounceMs = 0;
  export let remoteChangeOpen = false;

  const runtimeClient = useRuntimeClient();
  const gitPullMutation = createRuntimeServiceGitPullMutation(runtimeClient);
  const gitPushMutation = createRuntimeServiceGitPushMutation(runtimeClient);

  let mergeConflictResolutionDialog = false;
  let errorFromGitCommand: ConnectError | null = null;

  // When autoOpen is true and the two underlying GitStatus queries refetch
  // independently after a mutation, the derived source can briefly re-emit
  // the prior `hasRemoteChanges: true` while one refetch is in flight, then
  // settle to `false` once both land. Debouncing lets the latest value win
  // once emissions go quiet so the dialog doesn't flicker.
  const syncRemoteChangeOpen = debounce((value: boolean) => {
    remoteChangeOpen = value;
  }, debounceMs);

  onDestroy(() => syncRemoteChangeOpen.cancel());

  $: hasLocalCommitsOnCurrent = $gitStatusSource.hasLocalCommitsOnCurrent;
  $: if (autoOpen) {
    syncRemoteChangeOpen($gitStatusSource.hasRemoteChanges);
  }

  $: ({ isPending: gitPullPending, error: gitPullError } = $gitPullMutation);
  $: ({ isPending: gitPushPending, error: gitPushError } = $gitPushMutation);
  $: loading = gitPullPending || gitPushPending;
  $: dialogError =
    (gitPullError as ConnectError | null) ??
    (gitPushError as ConnectError | null) ??
    errorFromGitCommand;

  function invalidate() {
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(
        runtimeClient.instanceId,
        {},
      ),
    });
    if (primaryBranch) {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {
          remoteBranch: primaryBranch,
        }),
      });
    }
  }

  async function handleFetchRemoteCommits() {
    // GitPull can't auto-merge cleanly when there are unpushed local commits;
    // jump straight to the force/keep choice.
    if (hasLocalCommitsOnCurrent) {
      mergeConflictResolutionDialog = true;
      return;
    }

    errorFromGitCommand = null;
    const resp = await $gitPullMutation.mutateAsync({ discardLocal: false });
    invalidate();

    if (!resp.output) {
      if (autoPush) {
        await $gitPushMutation.mutateAsync({});
        invalidate();
      }
      remoteChangeOpen = false;
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
    invalidate();

    if (resp.output) {
      errorFromGitCommand = { message: resp.output } as ConnectError;
      return;
    }

    if (autoPush) {
      // Force-push: discardLocal rewrote local history, so the push must
      // overwrite the remote feature branch to match.
      await $gitPushMutation.mutateAsync({ force: true });
      invalidate();
    }

    remoteChangeOpen = false;
    mergeConflictResolutionDialog = false;
    eventBus.emit("notification", {
      message:
        "Remote project changes fetched and merged. Your changes have been stashed.",
    });
  }
</script>

<ProjectContainsRemoteChangesDialog
  bind:open={remoteChangeOpen}
  {loading}
  error={dialogError}
  onFetchAndMerge={handleFetchRemoteCommits}
/>

<MergeConflictResolutionDialog
  bind:open={mergeConflictResolutionDialog}
  {loading}
  error={dialogError}
  onUseLatestVersion={handleForceFetchRemoteCommits}
/>
