<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { createWatchFilesClient } from "@rilldata/web-common/features/entity-management/watch-files-client";
  import { createWatchResourceClient } from "@rilldata/web-common/features/entity-management/watch-resources-client";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
  import RillDeveloperLayout from "@rilldata/web-common/layout/RillDeveloperLayout.svelte";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { onDestroy, onMount } from "svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  const fileWatcher = createWatchFilesClient(queryClient);
  const resourceWatcher = createWatchResourceClient(queryClient);

  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => errorEventHandler?.requestErrorEventHandler(error, query);

  export let data;

  $: host = data.host;
  $: instanceId = data.instanceId;

  $: fileWatcher.watch(`${host}/v1/instances/${instanceId}/files/watch`);

  $: resourceWatcher.watch(
    `${host}/v1/instances/${instanceId}/resources/-/watch`,
  );

  beforeNavigate(retainFeaturesFlags);

  onMount(() => {
    const stopJavascriptErrorListeners =
      errorEventHandler?.addJavascriptErrorListeners();
    void fileArtifacts.init(queryClient, "default");

    return () => {
      stopJavascriptErrorListeners?.();
    };
  });

  onDestroy(() => {
    fileWatcher.abort();
    resourceWatcher.abort();
  });

  function handleVisibilityChange(
    e: Event & {
      currentTarget: EventTarget & Window;
    },
  ) {
    if (e.currentTarget.document.visibilityState === "visible") {
      fileWatcher.reconnect();
      resourceWatcher.reconnect();
    } else {
      fileWatcher.throttle();
      resourceWatcher.throttle();
    }
  }
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

<QueryClientProvider client={queryClient}>
  <RillDeveloperLayout>
    <slot />
  </RillDeveloperLayout>
</QueryClientProvider>
