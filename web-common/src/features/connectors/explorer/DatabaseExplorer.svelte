<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import type { V1AnalyzedConnector } from "../../../runtime-client";
  import DatabaseEntry from "./DatabaseEntry.svelte";
  import { useListDatabaseSchemas } from "../selectors";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;

  $: databaseSchemasQuery = useListDatabaseSchemas(instanceId, connectorName);

  $: ({ data, error, isLoading } = $databaseSchemasQuery);
</script>

<div class="wrapper">
  {#if isLoading}
    <span class="message pl-6">Loading tables...</span>
  {:else if error}
    <span class="message pl-6"
      >Error: {error.message || error.response?.data?.message}</span
    >
  {:else if data}
    {#if data.length === 0}
      <span class="message pl-6">No tables found</span>
    {:else}
      <ol transition:slide={{ duration }}>
        {#each data as database (database)}
          <DatabaseEntry {instanceId} {connector} {database} {store} />
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
    @apply pr-3.5 py-2; /* left-padding is set inline above */
    @apply text-gray-500;
  }
</style>
