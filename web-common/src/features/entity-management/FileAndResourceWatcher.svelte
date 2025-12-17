<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { fileAndResourceWatcher } from "./file-and-resource-watcher";
  import { ConnectionStatus } from "@rilldata/web-common/runtime-client/sse-connection-manager";

  export let host: string;
  export let instanceId: string;

  const { status: statusStore } = fileAndResourceWatcher;

  $: watcherEndpoint = `${host}/v1/instances/${instanceId}/sse?events=file,resource`;

  $: fileAndResourceWatcher.watch(watcherEndpoint);

  $: status = $statusStore;

  $: closed = status === ConnectionStatus.CLOSED;

  onMount(() => {
    void fileArtifacts.init(queryClient, instanceId);

    return () => fileAndResourceWatcher.close(true);
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      fileAndResourceWatcher.heartbeat();
    } else {
      fileAndResourceWatcher.scheduleAutoClose(true);
    }
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => fileAndResourceWatcher.scheduleAutoClose()}
  on:click={() => fileAndResourceWatcher.heartbeat()}
  on:keydown={() => fileAndResourceWatcher.heartbeat()}
  on:focus={() => fileAndResourceWatcher.heartbeat()}
/>

{#if closed}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body="Try restarting the Rill via the CLI"
  />
{:else}
  <slot />
{/if}
