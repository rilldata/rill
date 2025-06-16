<script lang="ts">
  import type { ConnectError } from "@connectrpc/connect";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import RemoteProjectContainsUpdatesDialog from "@rilldata/web-common/features/project/RemoteProjectContainsUpdatesDialog.svelte";
  import {
    createLocalServiceGitPull,
    createLocalServiceGitStatus,
  } from "@rilldata/web-common/runtime-client/local-service";

  const gitStatusQuery = createLocalServiceGitStatus();
  const gitPullMutation = createLocalServiceGitPull();

  let remoteChangeDialog = false;
  let mergeConflictResolutionDialog = false;

  $: if (!$gitStatusQuery.isPending) {
    remoteChangeDialog = $gitStatusQuery.data
      ? $gitStatusQuery.data?.remoteCommits > 0
      : false;
  }

  $: ({ isPending: githubPullPending, error: githubPullError } =
    $gitPullMutation);
  let customError: ConnectError | null = null;
  $: error = githubPullError ?? customError;

  const MergeConflictsError =
    /Your local changes to the following files would be overwritten by merge/;

  async function handleFetchRemoteCommits(discardLocal: boolean) {
    if (!discardLocal && $gitStatusQuery.data!.localCommits > 0) {
      // Since we can't really merge remote commits with local commits,
      // we just show the user the merge conflicts dialog for confirmation to clear it.
      mergeConflictResolutionDialog = true;
      return;
    }

    customError = null;
    const resp = await $gitPullMutation.mutateAsync({
      discardLocal,
    });
    // `output` is populated with the error from the actual git pull command run.
    if (!resp.output) {
      remoteChangeDialog = false;
      mergeConflictResolutionDialog = false;
      return;
    }
    if (!discardLocal && MergeConflictsError.test(resp.output)) {
      mergeConflictResolutionDialog = true;
    } else {
      customError = {
        message: resp.output,
      } as ConnectError;
    }
  }
</script>

<RemoteProjectContainsUpdatesDialog
  bind:open={remoteChangeDialog}
  loading={githubPullPending}
  {error}
  onFetchAndMerge={() => handleFetchRemoteCommits(false)}
/>

<MergeConflictResolutionDialog
  bind:open={mergeConflictResolutionDialog}
  loading={githubPullPending}
  {error}
  onUseLatestVersion={() => handleFetchRemoteCommits(true)}
/>
