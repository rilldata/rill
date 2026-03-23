<script lang="ts">
  import {
    type ConnectorExplorerEntry,
    filterConnectorExplorerTree,
  } from "@rilldata/web-common/features/add-data/explorer/tree.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { writable } from "svelte/store";
  import { getAnalyzedConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
  import ExplorerNode from "@rilldata/web-common/features/add-data/explorer/ExplorerNode.svelte";
  import { Search } from "@rilldata/web-common/components/search";

  export let connectorName: string;
  export let onSelect: (entry: ConnectorExplorerEntry) => void;

  const runtimeClient = useRuntimeClient();

  $: connectors = getAnalyzedConnectors(runtimeClient, false);
  $: analyzedConnector = $connectors.data?.connectors?.find(
    (c) => c.name === connectorName,
  );

  const searchTextStore = writable("");
  $: connectorExplorerTree = filterConnectorExplorerTree(
    runtimeClient,
    queryClient,
    analyzedConnector,
    searchTextStore,
  );
  $: ({ data, loading, error } = $connectorExplorerTree);
  $: allError = analyzedConnector?.errorMessage || error;

  let selectedEntry: ConnectorExplorerEntry | undefined;
  function handleSelect(entry: ConnectorExplorerEntry) {
    selectedEntry = entry;
    onSelect(entry);
  }
</script>

<div class="mb-2 mr-2">
  <Search bind:value={$searchTextStore} />
</div>
{#if allError}
  <span class="message pl-6">Error: {allError}</span>
{:else if loading}
  <span class="message pl-6">Loading tables...</span>
{:else if data}
  {#each data as node (node.name)}
    <ExplorerNode
      {node}
      forceExpand={!!$searchTextStore}
      level={0}
      {selectedEntry}
      onSelect={handleSelect}
    />
  {:else}
    <span class="message pl-6">No tables found</span>
  {/each}
{/if}
