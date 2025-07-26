<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import type { V1AnalyzedConnector } from "../../runtime-client";
  import DatabaseEntry from "./DatabaseEntry.svelte";
  import {
    useDatabasesFromSchemas,
    useConnectorCapabilities,
  } from "./selectors";
  import { useDatabases as useDatabasesLegacy } from "./olap/selectors";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;

  // Determine which API to use based on connector capabilities
  $: capabilities = useConnectorCapabilities(instanceId, connectorName);
  $: shouldUseNewAPI =
    $capabilities?.implementsSqlStore || !$capabilities?.implementsOlap;

  // Use appropriate selector based on connector type
  $: databasesQuery = shouldUseNewAPI
    ? useDatabasesFromSchemas(instanceId, connectorName)
    : useDatabasesLegacy(instanceId, connectorName);

  $: ({ data, error, isLoading } = $databasesQuery);
</script>

<div class="wrapper">
  {#if isLoading}
    <span class="message">Loading tables...</span>
  {:else if error}
    <span class="message"
      >Error: {error.message || error.response?.data?.message}</span
    >
  {:else if data}
    {#if data.length === 0}
      <span class="message">No tables found</span>
    {:else}
      <ol transition:slide={{ duration }}>
        {#each data as database (database)}
          <DatabaseEntry
            {instanceId}
            {connector}
            {database}
            {store}
            useNewAPI={shouldUseNewAPI}
          />
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
    @apply pl-6 pr-3.5 py-2;
    @apply text-gray-500;
  }
</style>
