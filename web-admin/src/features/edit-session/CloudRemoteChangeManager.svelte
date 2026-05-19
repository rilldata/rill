<script lang="ts">
  import { getDeploymentGithubStatus } from "@rilldata/web-admin/features/edit-session/selectors.ts";
  import RemoteSyncDialogs from "@rilldata/web-common/features/project/RemoteSyncDialogs.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onMount } from "svelte";
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

  let remoteChangeOpen = false;

  // Other components (e.g. PublishPopover) request the dialog via the bus
  // instead of mounting their own RemoteSyncDialogs, so this stays the single
  // owner of the pull/conflict dialog state in the edit session.
  onMount(() =>
    eventBus.on("remote-changes-detected", () => {
      remoteChangeOpen = true;
    }),
  );
</script>

<RemoteSyncDialogs
  bind:remoteChangeOpen
  {gitStatusSource}
  {primaryBranch}
  autoOpen
  autoPush
  debounceMs={500}
/>
