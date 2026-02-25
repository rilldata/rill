<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { fileAndResourceWatcher } from "./file-and-resource-watcher";
  import { ConnectionStatus } from "@rilldata/web-common/runtime-client/sse-connection-manager";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const {
    status: statusStore,
    heartbeat,
    scheduleAutoClose,
  } = fileAndResourceWatcher;

  const runtimeClient = useRuntimeClient();

  export let host: string;
  export let instanceId: string;

  $: fileAndResourceWatcher.setInstanceId(instanceId);

  $: fileAndResourceWatcher.setClient(runtimeClient);
  $: watcherEndpoint = `${host}/v1/instances/${instanceId}/sse?events=file,resource`;

  $: fileAndResourceWatcher.watch(watcherEndpoint);

  $: status = $statusStore;

  onMount(() => {
    void fileArtifacts.init(runtimeClient, queryClient, instanceId);

    return () => fileAndResourceWatcher.close(true);
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      heartbeat();
    } else {
      scheduleAutoClose(true);
    }
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => scheduleAutoClose()}
  on:click={heartbeat}
  on:keydown={heartbeat}
  on:focus={heartbeat}
/>

{#if status === ConnectionStatus.CLOSED}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body="Try restarting the Rill via the CLI"
  />
{:else}
  <slot />
{/if}
