<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { WatchFilesClient } from "@rilldata/web-common/features/entity-management/WatchFilesClient";
  import { WatchResourcesClient } from "@rilldata/web-common/features/entity-management/WatchResourcesClient";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { TabCommunicator } from "@rilldata/web-common/lib/tab-communicator.ts";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import Banner from "@rilldata/web-common/components/banner/Banner.svelte";

  const fileWatcher = new WatchFilesClient().client;
  const resourceWatcher = new WatchResourcesClient().client;
  const { retryAttempts: fileAttempts, closed: fileWatcherClosed } =
    fileWatcher;
  const { retryAttempts: resourceAttempts, closed: resourceWatcherClosed } =
    resourceWatcher;
  const tabCommunicator = new TabCommunicator<void>("rill-dev");

  export let host: string;
  export let instanceId: string;

  $: fileWatcher.watch(
    `${host}/v1/instances/${instanceId}/files/watch?stream=files`,
    true,
  );

  $: resourceWatcher.watch(
    `${host}/v1/instances/${instanceId}/resources/-/watch?stream=resources`,
    true,
  );

  $: failed = $fileAttempts >= 2 || $resourceAttempts >= 2;

  onMount(() => {
    void fileArtifacts.init(queryClient, instanceId);
    tabCommunicator.on("focused", handleAnotherRillDevFocused);

    return () => {
      fileWatcher.close();
      resourceWatcher.close();
      tabCommunicator.close();
      tabCommunicator.off("focused", handleAnotherRillDevFocused);
    };
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      fileWatcher.heartbeat();
      resourceWatcher.heartbeat();
      tabCommunicator.send("focused");
    } else {
      fileWatcher.throttle(true);
      resourceWatcher.throttle(true);
    }
  }

  function handleAnotherRillDevFocused() {
    fileWatcher.close();
    resourceWatcher.close();
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => {
    fileWatcher.throttle();
    resourceWatcher.throttle();
  }}
  on:click={() => {
    fileWatcher.heartbeat();
    resourceWatcher.heartbeat();
  }}
  on:keydown={() => {
    fileWatcher.heartbeat();
    resourceWatcher.heartbeat();
  }}
  on:focus={() => {
    fileWatcher.heartbeat();
    resourceWatcher.heartbeat();
  }}
/>

{#if failed}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body="Try restarting the server"
  />
{:else}
  {#if $fileWatcherClosed || $resourceWatcherClosed}
    <Banner
      banner={{
        message:
          "Connection closed due to inactivity. Interact with the page to reconnect.",
        type: "warning",
        iconType: "alert",
      }}
    />
  {/if}
  <slot />
{/if}
