<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import { V1AnalyzedConnector } from "../../../runtime-client";
  import DatabaseEntry from "./DatabaseEntry.svelte";
  import { useDatabases } from "./selectors";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;

  $: connectorName = connector?.name as string;

  $: dataBasesQuery = useDatabases(instanceId, connectorName);

  $: ({ data, error, isLoading } = $dataBasesQuery);
</script>

<div class="wrapper">
  {#if isLoading}
    <span class="message">Loading tables...</span>
  {:else if error}
    <span class="message">Error: {error.response.data.message}</span>
  {:else if data}
    {#if data.length === 0}
      <span class="message">No tables found</span>
    {:else}
      <ol transition:slide={{ duration }}>
        {#each data as database (database)}
          <DatabaseEntry {instanceId} {connector} {database} />
        {/each}
      </ol>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex flex-col overflow-y-auto;
  }

  .message {
    @apply pl-2 pr-3.5 pt-2 pb-2 text-gray-500;
  }
</style>
