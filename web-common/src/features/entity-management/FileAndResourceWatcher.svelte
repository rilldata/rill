<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onDestroy, setContext } from "svelte";
  import {
    ConnectionStatus,
    FileAndResourceWatcher,
  } from "./file-and-resource-watcher";
  import { WATCHER_CONTEXT_KEY } from "./watcher-context";

  export let errorBody = "Try restarting the Rill via the CLI";
  /** Idle lifecycle strategy. "aggressive" is right for Rill Developer
   *  (browser HTTP-connection limit bites); "none" keeps the connection
   *  open and is right for consumers that need long-lived streams. */
  export let lifecycle: "aggressive" | "none";
  /** Optional hook fired before each reconnect attempt, used by the cloud
   *  editor to refresh its JWT. */
  export let onBeforeReconnect: (() => Promise<void>) | undefined = undefined;

  const runtimeClient = useRuntimeClient();

  // File artifacts client must be set before descendants try to read it.
  fileArtifacts.setClient(runtimeClient);

  // Construct the watcher synchronously so `setContext` runs during init —
  // descendants that call `getContext` during their own init would see
  // undefined otherwise.
  const watcher = new FileAndResourceWatcher({
    runtimeClient,
    queryClient,
    lifecycle,
    // Keep a stable function reference at construction time while always
    // delegating to the latest prop callback.
    onBeforeReconnect: async () => {
      await onBeforeReconnect?.();
    },
  });

  setContext(WATCHER_CONTEXT_KEY, { watcher, status: watcher.status });

  const status = watcher.status;

  const watcherEndpoint = `${runtimeClient.host}/v1/instances/${runtimeClient.instanceId}/sse?events=file,resource`;
  console.log("start", watcherEndpoint);
  watcher.start(watcherEndpoint);
  void fileArtifacts.init(runtimeClient, queryClient);

  onDestroy(() => {
    console.log("on destroy");
    watcher.close(true);
  });
</script>

{#if $status === ConnectionStatus.CLOSED}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body={errorBody}
  />
{:else}
  <slot />
{/if}
