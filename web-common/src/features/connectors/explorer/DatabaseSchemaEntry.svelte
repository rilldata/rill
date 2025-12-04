<script lang="ts">
  import { Database, Folder } from "lucide-svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import type {
    V1AnalyzedConnector,
    V1TableInfo,
  } from "../../../runtime-client";
  import TableEntry from "./TableEntry.svelte";
  import { useInfiniteListTables } from "../selectors";
  import Button from "../../../components/button/Button.svelte";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import { onMount } from "svelte";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let database: string;
  export let databaseSchema: string;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;

  $: expandedStore = store.getItem(connectorName, database, databaseSchema);
  $: expanded = $expandedStore;

  $: tablesQuery = useInfiniteListTables(
    instanceId,
    connectorName,
    database,
    databaseSchema,
    100,
    expanded,
  );

  $: ({
    data,
    error,
    isLoading,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  } = $tablesQuery);

  let loadMoreContainer: HTMLDivElement;
  let isNarrow = false;

  onMount(() => {
    if (!loadMoreContainer) return;
    const observer = new ResizeObserver((entries) => {
      const width = entries[0].contentRect.width;
      isNarrow = width < 220; // switch to shorter copy on narrow sidebar
    });
    observer.observe(loadMoreContainer);
    return () => observer.disconnect();
  });

  // Normalize V1TableInfo[] from ListTables into the TableEntry input shape
  $: typedData = data?.tables?.map((table: V1TableInfo) => ({
    name: table.name ?? "",
    database,
    databaseSchema,
    view: table.view ?? false,
  }));
</script>

<li aria-label={`${database}.${databaseSchema}`} class="database-schema-entry">
  <button
    class="database-schema-entry-header {database ? 'pl-[40px]' : 'pl-[22px]'}"
    class:open={expanded}
    on:click={() => store.toggleItem(connectorName, database, databaseSchema)}
  >
    <CaretDownIcon
      className="transform transition-transform text-gray-400 {expanded
        ? 'rotate-0'
        : '-rotate-90'}"
      size="14px"
    />
    <!-- Some databases do not have a full "database -> databaseSchema -> table" hierarchy. 
      When there are only two organizational levels,the API returns "databaseSchema -> table". 
      However, in these cases, we should use a Database icon (not a Folder icon) to represent the organizational structure. -->
    {#if !database}
      <Database size="14px" class="shrink-0 text-gray-400" />
    {:else}
      <Folder size="14px" class="shrink-0 text-gray-400" />
    {/if}
    <span class="truncate">
      {databaseSchema}
    </span>
  </button>

  {#if expanded}
    {#if error && (!typedData || typedData.length === 0)}
      <div class="message {database ? 'pl-[78px]' : 'pl-[60px]'}">
        Error: {error.message}
      </div>
    {:else if isLoading && (!typedData || typedData.length === 0)}
      <div class="message {database ? 'pl-[78px]' : 'pl-[60px]'}">
        Loading tables...
      </div>
    {:else if connector?.errorMessage}
      <div class="message {database ? 'pl-[78px]' : 'pl-[60px]'}">
        {connector.errorMessage}
      </div>
    {:else if !connector.driver || !connector.driver.name}
      <div class="message {database ? 'pl-[78px]' : 'pl-[60px]'}">
        Connector not found
      </div>
    {:else if !typedData || typedData.length === 0}
      <div class="message {database ? 'pl-[78px]' : 'pl-[60px]'}">
        No tables found
      </div>
    {:else if typedData.length > 0}
      <ol>
        {#each typedData as tableInfo (tableInfo)}
          <TableEntry
            driver={connector.driver.name}
            connector={connectorName}
            showGenerateMetricsAndDashboard={(connector.driver.implementsOlap ||
              connector.driver.implementsWarehouse ||
              connector.driver.implementsSqlStore) ??
              false}
            showGenerateModel={(connector.driver.implementsWarehouse ||
              connector.driver.implementsSqlStore) ??
              false}
            isOlapConnector={connector.driver.implementsOlap ?? false}
            {database}
            {databaseSchema}
            table={tableInfo.name}
            {store}
          />
        {/each}
      </ol>

      {#if hasNextPage}
        <div
          class="load-more {database ? 'pl-[78px]' : 'pl-[60px]'}"
          bind:this={loadMoreContainer}
        >
          {#if error}
            <span class="error">Failed to load more tables.</span>
            <Button type="plain" small onClick={() => fetchNextPage()}>
              Retry
            </Button>
          {:else}
            <Button
              type="plain"
              small
              disabled={isFetchingNextPage}
              loading={isFetchingNextPage}
              loadingCopy={isNarrow ? "Loading..." : "Loading tables..."}
              class="w-full"
              onClick={() => fetchNextPage()}
            >
              {isNarrow ? "Load more" : "Load more tables"}
            </Button>
          {/if}
        </div>
      {/if}
    {/if}
  {/if}
</li>

<style lang="postcss">
  .database-schema-entry {
    @apply w-full;
    @apply flex flex-col;
  }

  .database-schema-entry-header {
    @apply h-6 pr-2; /* left-padding is set dynamically above */
    @apply flex items-center gap-x-1;
  }

  button:hover {
    @apply bg-slate-100;
  }

  .message {
    @apply pr-3.5 py-2; /* left-padding is set dynamically above */
    @apply text-gray-500;
  }

  .load-more {
    @apply py-2 pr-2;
  }
</style>
