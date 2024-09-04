<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { WatchFilesClient } from "@rilldata/web-common/features/entity-management/WatchFilesClient";
  import { WatchResourcesClient } from "@rilldata/web-common/features/entity-management/WatchResourcesClient";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";

  const fileWatcher = new WatchFilesClient().client;
  const resourceWatcher = new WatchResourcesClient().client;
  const fileAttempts = fileWatcher.retryAttempts;
  const resourceAttempts = resourceWatcher.retryAttempts;
  const closed = fileWatcher.closed;

  export let host: string;
  export let instanceId: string;

  $: fileWatcher.watch(`${host}/v1/instances/${instanceId}/files/watch`);

  $: resourceWatcher.watch(
    `${host}/v1/instances/${instanceId}/resources/-/watch`,
  );

  $: failed = $fileAttempts >= 2 || $resourceAttempts >= 2;

  onMount(() => {
    void fileArtifacts.init(queryClient, instanceId);

    return () => {
      fileWatcher.close();
      resourceWatcher.close();
    };
  });

  function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      fileWatcher.reconnect().catch(console.error);
      resourceWatcher.reconnect().catch(console.error);
    } else {
      fileWatcher.throttle(true);
      resourceWatcher.throttle(true);
    }
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => {
    fileWatcher.throttle(true);
    resourceWatcher.throttle(true);
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
  {#if $closed}
    <div class="bg-yellow-100 py-1 w-full">
      <div class="flex flex-row items-center mx-auto w-fit gap-x-2">
        <InfoCircle />
        <span>
          Connection closed due to inactivity. Interact with the page to
          reconnect.
        </span>
      </div>
    </div>
  {/if}
  <slot />
{/if}
