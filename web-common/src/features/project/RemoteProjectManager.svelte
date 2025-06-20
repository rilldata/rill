<script lang="ts">
  import type { ConnectError } from "@connectrpc/connect";
  import { isMergeConflictError } from "@rilldata/web-common/features/project/github-utils.ts";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import {
    createLocalServiceGitPull,
    createLocalServiceGitStatus,
  } from "@rilldata/web-common/runtime-client/local-service";

  const gitStatusQuery = createLocalServiceGitStatus();
  const gitPullMutation = createLocalServiceGitPull();

  let remoteChangeDialog = false;
  let mergeConflictResolutionDialog = false;

  $: if ($gitStatusQuery.data) {
    remoteChangeDialog = $gitStatusQuery.data.remoteCommits > 0;
  }

  $: ({ isPending: githubPullPending, error: githubPullError } =
    $gitPullMutation);
  let errorFromGitCommand: ConnectError | null = null;
  $: error = githubPullError ?? errorFromGitCommand;

  async function handleFetchRemoteCommits() {
    if ($gitStatusQuery.data!.localCommits > 0) {
      // Since we can't really merge remote commits with local commits,
      // we just show the user the merge conflicts dialog for confirmation to clear it.
      // We could directly show it since the data is in gitStatusQuery, but it feels like weird UX.
      mergeConflictResolutionDialog = true;
      return;
    }

    errorFromGitCommand = null;
    const resp = await $gitPullMutation.mutateAsync({
      discardLocal: false,
    });
    // TODO: download diff once API is ready

    if (!resp.output) {
      remoteChangeDialog = false;
      mergeConflictResolutionDialog = false;
      return;
    }

    if (isMergeConflictError(resp.output)) {
      mergeConflictResolutionDialog = true;
      return;
    }

    errorFromGitCommand = {
      message: resp.output,
    } as ConnectError;
  }

  async function handleForceFetchRemoteCommits() {
    errorFromGitCommand = null;
    const resp = await $gitPullMutation.mutateAsync({
      discardLocal: true,
    });
    // TODO: download diff once API is ready

    if (!resp.output) {
      remoteChangeDialog = false;
      mergeConflictResolutionDialog = false;
      return;
    }

    errorFromGitCommand = {
      message: resp.output,
    } as ConnectError;
  }
</script>

<ProjectContainsRemoteChangesDialog
  bind:open={remoteChangeDialog}
  loading={githubPullPending}
  {error}
  onFetchAndMerge={handleFetchRemoteCommits}
/>

<MergeConflictResolutionDialog
  bind:open={mergeConflictResolutionDialog}
  loading={githubPullPending}
  {error}
  onUseLatestVersion={handleForceFetchRemoteCommits}
/>
