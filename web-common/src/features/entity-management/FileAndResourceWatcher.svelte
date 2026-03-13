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
  export let errorBody = "Try restarting the Rill via the CLI";
  /** Keep the SSE connection open indefinitely (no auto-close on idle). */
  export let keepAlive = false;

  // Set client synchronously so children can access it during initial render.
  // init() (in onMount) handles the async resource prefetch.
  fileArtifacts.setClient(runtimeClient);

  $: fileAndResourceWatcher.setRuntimeClient(runtimeClient);
  $: fileAndResourceWatcher.setInstanceId(instanceId);

  $: watcherEndpoint = `${host}/v1/instances/${instanceId}/sse?events=file,resource`;

  $: fileAndResourceWatcher.watch(watcherEndpoint);

  $: status = $statusStore;

  onMount(() => {
    if (keepAlive) {
      fileAndResourceWatcher.disableAutoClose();
    }
    void fileArtifacts.init(runtimeClient, queryClient);

    return () => {
      if (keepAlive) {
        fileAndResourceWatcher.enableAutoClose();
      }
      fileAndResourceWatcher.close(true);
    };
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      heartbeat();
    } else if (!keepAlive) {
      scheduleAutoClose(true);
    }
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => {
    if (!keepAlive) scheduleAutoClose();
  }}
  on:click={heartbeat}
  on:keydown={heartbeat}
  on:focus={heartbeat}
/>

{#if status === ConnectionStatus.CLOSED}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body={errorBody}
  />
{:else}
  <slot />
{/if}
