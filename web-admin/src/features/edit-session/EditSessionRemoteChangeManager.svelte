<script lang="ts">
  import { getDeploymentGithubStatus } from "@rilldata/web-admin/features/edit-session/selectors.ts";
  import RemoteSyncDialogs from "@rilldata/web-common/features/project/RemoteSyncDialogs.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { derived } from "svelte/store";

  export let primaryBranch: string | undefined;

  const runtimeClient = useRuntimeClient();
  const gitStatusQuery = getDeploymentGithubStatus(
    runtimeClient,
    primaryBranch,
  );

  const gitStatusSource = derived(gitStatusQuery, ($q) => ({
    hasRemoteChanges: $q.data.hasRemoteChanges,
    hasLocalCommitsOnCurrent: $q.data.hasLocalCommitsOnCurrent,
  }));
</script>

<RemoteSyncDialogs
  {gitStatusSource}
  {primaryBranch}
  autoOpen
  autoPush
  debounceMs={500}
/>
