<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import { V1AnalyzedConnector } from "../../runtime-client";
  import TableEntry from "./TableEntry.svelte";
  import { useTables } from "./selectors";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;

  $: tablesQuery = useTables(instanceId, connector?.name);
  $: ({ data, error, isLoading } = $tablesQuery);
  $: typedTables = data?.tables as
    | {
        name: string;
        database: string;
        databaseSchema: string;
        hasUnsupportedDataTypes: boolean;
      }[]
    | undefined;
</script>

{#if connector && connector.name}
  <div class="wrapper">
    {#if error}
      <span class="message">Error loading tables</span>
    {:else if isLoading}
      <span class="message">Loading tables...</span>
    {:else if typedTables}
      {#if typedTables.length === 0}
        <span class="message">No tables found</span>
      {:else}
        <ol transition:slide={{ duration }}>
          {#each typedTables as tableInfo (tableInfo)}
            <TableEntry
              connectorInstanceId={instanceId}
              connector={connector.name}
              database={tableInfo.database}
              databaseSchema={tableInfo.databaseSchema}
              table={tableInfo.name}
              hasUnsupportedDataTypes={tableInfo.hasUnsupportedDataTypes}
            />
          {/each}
        </ol>
      {/if}
    {/if}
  </div>
{/if}

<style lang="postcss">
  .wrapper {
    @apply flex flex-col overflow-y-auto;
  }

  .message {
    @apply pl-2 pr-3.5 pt-2 pb-2 text-gray-500;
  }
</style>
