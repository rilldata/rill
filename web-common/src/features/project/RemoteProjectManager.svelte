<script lang="ts">
  import { page } from "$app/stores";
  import type { ConnectError } from "@connectrpc/connect";
  import { isMergeConflictError } from "@rilldata/web-common/features/project/deploy/github-utils.ts";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import type { GitStatusResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb.ts";
  import {
    createLocalServiceGitPull,
    createLocalServiceGitStatus,
    getLocalServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";

  const gitStatusQuery = createLocalServiceGitStatus();
  const gitPullMutation = createLocalServiceGitPull();

  let remoteChangeDialog = false;
  let mergeConflictResolutionDialog = false;
  $: inDeployPage = $page.route.id?.startsWith("/(misc)/deploy") ? true : false;

  $: if ($gitStatusQuery.data) {
    processGithubStatus($gitStatusQuery.data);
  }

  $: ({ isPending: githubPullPending, error: githubPullError } =
    $gitPullMutation);
  let errorFromGitCommand: ConnectError | null = null;
  $: error = githubPullError ?? errorFromGitCommand;

  function processGithubStatus(status: GitStatusResponse) {
    remoteChangeDialog = status.remoteCommits > 0;
  }

  async function handleFetchRemoteCommits() {
    if (inDeployPage) return; // Do not show the modal in deploy pages

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

    void queryClient.invalidateQueries({
      queryKey: getLocalServiceGitStatusQueryKey(),
    });

    remoteChangeDialog = false;

    if (!resp.output) {
      mergeConflictResolutionDialog = false;
      eventBus.emit("notification", {
        message: "Remote project changes fetched and merged.",
      });
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

    void queryClient.invalidateQueries({
      queryKey: getLocalServiceGitStatusQueryKey(),
    });

    if (!resp.output) {
      remoteChangeDialog = false;
      mergeConflictResolutionDialog = false;
      eventBus.emit("notification", {
        message:
          "Remote project changes fetched and merged. Your changes have been stashed.",
      });
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
