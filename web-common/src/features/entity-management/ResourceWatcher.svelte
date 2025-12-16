<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { fileWatcher, resourceWatcher } from "./watchers";
  import { MAX_RETRIES } from "@rilldata/web-common/runtime-client/watch-request-client";

  const { retryAttempts: fileAttempts } = fileWatcher;
  const { retryAttempts: resourceAttempts } = resourceWatcher;

  export let host: string;
  export let instanceId: string;

  $: fileWatcherEndpoint = `${host}/v1/instances/${instanceId}/files/watch?stream=files`;
  $: resourceWatcherEndpoint = `${host}/v1/instances/${instanceId}/resources/-/watch?stream=resources`;

  $: void fileWatcher.watch(fileWatcherEndpoint, true);

  $: void resourceWatcher.watch(resourceWatcherEndpoint, true);

  $: failed = $fileAttempts >= MAX_RETRIES || $resourceAttempts >= MAX_RETRIES;

  onMount(() => {
    void fileArtifacts.init(queryClient, instanceId);

    return () => {
      fileWatcher.close();
      resourceWatcher.close();
    };
  });

  async function handleVisibilityChange() {
    if (document.visibilityState === "visible") {
      await fileWatcher.heartbeat();
      await resourceWatcher.heartbeat();
    } else {
      fileWatcher.throttle(true);
      resourceWatcher.throttle(true);
    }
  }

  async function keepAlive() {
    await fileWatcher.heartbeat();
    await resourceWatcher.heartbeat();
  }
</script>

<svelte:window
  on:visibilitychange={handleVisibilityChange}
  on:blur={() => {
    fileWatcher.throttle();
    resourceWatcher.throttle();
  }}
  on:click={keepAlive}
  on:keydown={keepAlive}
  on:focus={keepAlive}
/>

{#if failed}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body="Try restarting the server"
  />
{:else}
  <slot />
{/if}
