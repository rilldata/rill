<script lang="ts">
  import { page } from "$app/stores";
  import type { ConnectError } from "@connectrpc/connect";
  import { isMergeConflictError } from "@rilldata/web-common/features/project/deploy/github-utils.ts";
  import MergeConflictResolutionDialog from "@rilldata/web-common/features/project/MergeConflictResolutionDialog.svelte";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createRuntimeServiceGitPullMutation,
    createRuntimeServiceGitStatus,
    getRuntimeServiceGitStatusQueryKey,
    type V1GitStatusResponse,
  } from "@rilldata/web-common/runtime-client";

  const runtimeClient = useRuntimeClient();

  const gitStatusQuery = createRuntimeServiceGitStatus(runtimeClient, {});
  const gitPullMutation = createRuntimeServiceGitPullMutation(runtimeClient);

  let remoteChangeDialog = false;
  let mergeConflictResolutionDialog = false;
  $: inDeployPage = $page.route.id?.startsWith("/(misc)/deploy") ? true : false;

  $: if ($gitStatusQuery.data) {
    processGithubStatus($gitStatusQuery.data);
  }

  $: ({ isPending: githubPullPending, error: githubPullError } =
    $gitPullMutation);
  let errorFromGitCommand: ConnectError | null = null;
  $: error = (githubPullError as ConnectError | null) ?? errorFromGitCommand;

  function processGithubStatus(status: V1GitStatusResponse) {
    remoteChangeDialog = Boolean(status.remoteCommits);
  }

  async function handleFetchRemoteCommits() {
    if (inDeployPage) return; // Do not show the modal in deploy pages

    if ($gitStatusQuery.data?.localCommits) {
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
      queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {}),
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
      queryKey: getRuntimeServiceGitStatusQueryKey(runtimeClient.instanceId, {}),
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
