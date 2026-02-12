<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { Wand2, Search, Plus } from "lucide-svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    createRuntimeServiceAnalyzeConnectors,
    type V1AnalyzedConnector,
  } from "../../../runtime-client";
  import DatabaseExplorer from "./DatabaseExplorer.svelte";
  import { ConnectorExplorerStore } from "./connector-explorer-store";
  import { generateMetricsFromTable } from "../../metrics-views/ai-generation/generateMetricsView";
  import { featureFlags } from "../../feature-flags";
  import { dataExplorerStore } from "./data-explorer-store";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { connectorIconMapping } from "../connector-icon-mapping";
  import { getConnectorIconKey } from "../connectors-utils";
  import { FORM_HEIGHT_DEFAULT } from "../../sources/modal/connector-schemas";
  import { addSourceModal } from "../../sources/modal/add-source-visibility";
  import { debounce } from "../../../lib/create-debouncer";

  const { ai } = featureFlags;
  $: ({ instanceId } = $runtime);
  $: ({ open, connector: initialConnector } = $dataExplorerStore);

  let selectedConnector: V1AnalyzedConnector | null = null;

  $: driverName = initialConnector?.driver?.name ?? null;

  function isMatchingConnector(c: V1AnalyzedConnector): boolean {
    if (driverName === "motherduck") {
      const path = c.config?.path;
      return typeof path === "string" && path.startsWith("md:");
    }
    return c?.driver?.name === driverName;
  }

  $: connectorsQuery = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      enabled: open && !!driverName,
      select: (data) => {
        if (!data?.connectors || !driverName) return [];
        return data.connectors
          .filter(isMatchingConnector)
          .sort((a, b) => (a?.name ?? "").localeCompare(b?.name ?? ""));
      },
    },
  });

  $: sameDriverConnectors = $connectorsQuery.data ?? [];

  $: if (open && sameDriverConnectors.length > 0) {
    const isCurrentValid =
      selectedConnector &&
      sameDriverConnectors.some((c) => c.name === selectedConnector?.name);
    if (!isCurrentValid) {
      selectedConnector =
        sameDriverConnectors.find((c) => c.name === initialConnector?.name) ??
        sameDriverConnectors[0];
    }
  }

  let selectedTable: {
    connector: string;
    database: string;
    databaseSchema: string;
    table: string;
  } | null = null;

  let isGenerating = false;
  let searchInput = "";
  let searchQuery = "";

  const updateSearch = debounce((value: string) => {
    searchQuery = value;
  }, 200);

  $: updateSearch(searchInput);

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
      }
    },
  );

  function handleClose() {
    dataExplorerStore.close();
    selectedTable = null;
    selectedConnector = null;
    isGenerating = false;
    searchInput = "";
    searchQuery = "";
    selectionStore.clearSelection();
  }

  function handleSelectConnector(connector: V1AnalyzedConnector) {
    selectedConnector = connector;
    selectedTable = null;
    searchInput = "";
    searchQuery = "";
    selectionStore.clearSelection();
  }

  async function handleGenerateMetrics() {
    if (!selectedTable) return;

    isGenerating = true;
    try {
      await generateMetricsFromTable(
        instanceId,
        selectedTable.connector,
        selectedTable.database,
        selectedTable.databaseSchema,
        selectedTable.table,
        false, // createExplore - only create metrics view
        true, // isOlapConnector
        BehaviourEventMedium.Button,
        MetricsEventSpace.Modal,
      );
      handleClose();
    } catch (error) {
      console.error("Failed to generate metrics:", error);
    } finally {
      isGenerating = false;
    }
  }

  $: displayIcon = initialConnector
    ? connectorIconMapping[getConnectorIconKey(initialConnector)]
    : null;
  $: driverDisplayName =
    initialConnector?.driver?.displayName ??
    initialConnector?.driver?.name ??
    "OLAP";

  function handleAddNewConnector() {
    if (!initialConnector?.driver) return;
    const driver = initialConnector.driver;
    handleClose();
    addSourceModal.openForConnector(driver.name ?? "", driver);
  }
</script>

<Dialog.Root
  {open}
  onOpenChange={(isOpen) => {
    if (!isOpen) handleClose();
  }}
>
  <Dialog.Content class="max-w-5xl p-0 gap-0 overflow-hidden">
    <Dialog.Title class="p-4 border-b border-border">
      <div class="flex items-center gap-2">
        {#if displayIcon}
          <svelte:component this={displayIcon} size="18px" />
        {/if}
        <span class="text-lg font-semibold">{driverDisplayName}</span>
      </div>
    </Dialog.Title>

    <div class="flex {FORM_HEIGHT_DEFAULT}">
      <!-- Left panel: Connector list -->
      <div
        class="w-64 border-r border-border overflow-y-auto flex-shrink-0 bg-surface-subtle"
      >
        <div class="p-4">
          <div class="text-sm font-semibold text-fg-primary">
            Existing connectors
          </div>
          <div class="text-xs text-fg-muted mt-1">
            Choose data from an existing connection or create a new connector.
          </div>

          <div class="flex flex-col gap-2 mt-4">
            {#each sameDriverConnectors as conn (conn.name)}
              {@const isSelected = selectedConnector?.name === conn.name}
              <button
                class="w-full text-left px-3 py-2 rounded-md text-sm flex items-center gap-2 {isSelected
                  ? 'border border-primary-300 bg-primary-50 dark:border-primary-700 dark:bg-primary-950'
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

          <button
            class="w-full text-left px-3 py-1.5 rounded text-sm font-medium text-primary-500 hover:bg-surface-hover flex items-center gap-1.5 mt-3"
            on:click={handleAddNewConnector}
          >
            <Plus size="14" />
            New connector
          </button>
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

    <div class="p-4 border-t border-border flex items-center">
      <Button type="secondary" onClick={handleClose}>Back</Button>

      <Button
        type="primary"
        class="ml-auto"
        disabled={!selectedTable || isGenerating}
        loading={isGenerating}
        loadingCopy="Generating..."
        onClick={handleGenerateMetrics}
      >
        <span class="flex items-center gap-1.5">
          {#if $ai}
            <Wand2 size="14" />
          {/if}
          Generate Metrics{#if $ai}{" "}with AI{/if}
        </span>
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
