<script lang="ts">
  import { Search } from "lucide-svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    createRuntimeServiceAnalyzeConnectors,
    type V1AnalyzedConnector,
    type V1ConnectorDriver,
  } from "../../../runtime-client";
  import DatabaseExplorer from "../../connectors/explorer/DatabaseExplorer.svelte";
  import { ConnectorExplorerStore } from "../../connectors/explorer/connector-explorer-store";
  import { getEffectiveDriverName } from "../../connectors/connectors-utils";
  import { onDestroy } from "svelte";

  export let connector: V1ConnectorDriver;
  export let connectorInstanceName: string | null = null;

  /** The selected table (if any), consumed by the parent to enable/disable actions. */
  export let selectedTable: {
    connector: string;
    database: string;
    databaseSchema: string;
    table: string;
  } | null = null;

  $: ({ instanceId } = $runtime);

  $: driverName = connector?.name ?? null;

  function isMatchingConnector(c: V1AnalyzedConnector): boolean {
    return getEffectiveDriverName(c) === driverName;
  }

  $: connectorsQuery = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      enabled: !!driverName,
      select: (data) => {
        if (!data?.connectors || !driverName) return [];
        return data.connectors
          .filter(isMatchingConnector)
          .sort((a, b) => (a?.name ?? "").localeCompare(b?.name ?? ""));
      },
    },
  });

  $: sameDriverConnectors = $connectorsQuery.data ?? [];

  let selectedConnector: V1AnalyzedConnector | null = null;

  $: if (sameDriverConnectors.length > 0) {
    const isCurrentValid =
      selectedConnector &&
      sameDriverConnectors.some((c) => c.name === selectedConnector?.name);
    if (!isCurrentValid) {
      // Prefer the connector instance we just created/tested, otherwise the first
      selectedConnector =
        (connectorInstanceName
          ? sameDriverConnectors.find((c) => c.name === connectorInstanceName)
          : null) ?? sameDriverConnectors[0];
    }
  }

  let searchInput = "";
  let searchQuery = "";
  let searchTimeout: ReturnType<typeof setTimeout>;

  function updateSearch(value: string) {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      searchQuery = value;
    }, 200);
  }

  $: updateSearch(searchInput);

  onDestroy(() => {
    clearTimeout(searchTimeout);
  });

  const selectionStore = new ConnectorExplorerStore(
    {
      allowSelectTable: true,
      allowContextMenu: false,
      allowNavigateToTable: false,
      allowShowSchema: true,
      localStorage: false,
    },
    (connectorName, database, schema, table) => {
      if (table) {
        selectedTable = {
          connector: connectorName,
          database: database ?? "",
          databaseSchema: schema ?? "",
          table,
        };
      } else {
        selectedTable = null;
      }
    },
  );

  function handleSelectConnector(conn: V1AnalyzedConnector) {
    selectedConnector = conn;
    selectedTable = null;
    searchInput = "";
    searchQuery = "";
    selectionStore.clearSelection();
    selectionStore.store.update((state) => ({ ...state, expandedItems: {} }));
  }

  /** Reset state when parent unmounts or navigates away. */
  export function reset() {
    clearTimeout(searchTimeout);
    selectedTable = null;
    selectedConnector = null;
    searchInput = "";
    searchQuery = "";
    selectionStore.clearSelection();
    selectionStore.store.update((state) => ({ ...state, expandedItems: {} }));
  }
</script>

<div class="flex h-full">
  <!-- Left panel: Connector list -->
  <div
    class="w-64 border-r border-border overflow-y-auto flex-shrink-0 bg-surface-subtle"
  >
    <div class="p-4">
      <div class="text-sm font-semibold text-fg-primary">
        Existing connectors
      </div>
      <div class="text-xs text-fg-muted mt-1">
        Choose data from an existing connection.
      </div>

      <div class="flex flex-col gap-2 mt-4">
        {#each sameDriverConnectors as conn (conn.name)}
          {@const isSelected = selectedConnector?.name === conn.name}
          <button
            class="w-full text-left px-3 py-2 rounded-md text-sm flex items-center gap-2 {isSelected
              ? 'border border-accent-primary bg-surface-active'
              : 'border border-transparent hover:bg-surface-hover'}"
            on:click={() => handleSelectConnector(conn)}
          >
            <span
              class="shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center {isSelected
                ? 'border-primary-500'
                : 'border-gray-300 dark:border-gray-600'}"
            >
              {#if isSelected}
                <span class="w-2 h-2 rounded-full bg-primary-500"></span>
              {/if}
            </span>
            <span class="truncate">{conn.name}</span>
          </button>
        {/each}
        {#if sameDriverConnectors.length === 0}
          <div class="text-sm text-fg-secondary">No connectors found</div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Right panel: Table browser -->
  <div class="flex-1 flex flex-col overflow-hidden p-4 gap-3">
    <div>
      <div class="text-sm font-semibold text-fg-primary">Data explorer</div>
      <div class="text-xs text-fg-muted mt-1">
        Pick a table to explore your data.
      </div>
    </div>

    <div class="relative flex items-center border border-border rounded-md">
      <Search size="14" class="absolute left-3 text-fg-muted" />
      <input
        type="text"
        placeholder="Search"
        bind:value={searchInput}
        class="w-full pl-8 pr-3 py-1.5 bg-transparent border-none rounded-md text-sm text-fg-primary placeholder:text-fg-muted focus:outline-none"
      />
    </div>

    <div
      class="flex-1 overflow-y-auto border border-border rounded-lg min-h-0"
      style="--explorer-indent-offset: -14px"
    >
      {#if selectedConnector}
        <DatabaseExplorer
          {instanceId}
          connector={selectedConnector}
          store={selectionStore}
          searchPattern={searchQuery}
        />
      {/if}
    </div>
  </div>
</div>
