<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { createWatchFilesClient } from "@rilldata/web-common/features/entity-management/watch-files-client";
  import { createWatchResourceClient } from "@rilldata/web-common/features/entity-management/watch-resources-client";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";

  const fileWatcher = createWatchFilesClient();
  const resourceWatcher = createWatchResourceClient();

  export let host: string;
  export let instanceId: string;

  $: fileWatcher.watch(`${host}/v1/instances/${instanceId}/files/watch`);

  $: resourceWatcher.watch(
    `${host}/v1/instances/${instanceId}/resources/-/watch`,
  );

  onMount(() => {
    const stopJavascriptErrorListeners =
      errorEventHandler?.addJavascriptErrorListeners();
    void fileArtifacts.init(queryClient, instanceId);

    return () => {
      fileWatcher.cancel();
      resourceWatcher.cancel();
      stopJavascriptErrorListeners?.();
    };
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      fileWatcher.reconnect();
      resourceWatcher.reconnect();
    } else {
      fileWatcher.throttle();
      resourceWatcher.throttle();
    }
  }
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

<slot />
