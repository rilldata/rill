<script lang="ts">
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import {
    fileAndResourceWatcher,
    MAX_RETRIES,
  } from "./file-and-resource-watcher";

  export let host: string;
  export let instanceId: string;

  const { retryAttempts } = fileAndResourceWatcher;

  $: watcherEndpoint = `${host}/v1/instances/${instanceId}/sse?events=file,resource`;

  $: fileAndResourceWatcher.watch(watcherEndpoint);

  $: failed = $retryAttempts >= MAX_RETRIES;

  onMount(() => {
    void fileArtifacts.init(queryClient, instanceId);

    return () => fileAndResourceWatcher.close();
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

{#if failed}
  <ErrorPage
    fatal
    statusCode={500}
    header="Error connecting to runtime"
    body="Try restarting the Rill via the CLI"
  />
{:else}
  <slot />
{/if}
