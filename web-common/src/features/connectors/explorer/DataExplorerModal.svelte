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

  const { ai } = featureFlags;
  $: ({ instanceId } = $runtime);
  $: ({ open, connector: initialConnector } = $dataExplorerStore);

  // Track the currently selected connector (can be different from initial)
  let selectedConnector: V1AnalyzedConnector | null = null;

  // Query all connectors and filter by same driver
  $: connectorsQuery = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      enabled: open && !!initialConnector?.driver?.name,
      select: (data) => {
        if (!data?.connectors || !initialConnector?.driver?.name) return [];
        // Filter connectors with the same driver
        return data.connectors
          .filter((c) => c?.driver?.name === initialConnector.driver?.name)
          .sort((a, b) => (a?.name ?? "").localeCompare(b?.name ?? ""));
      },
    },
  });

  $: sameDriverConnectors = $connectorsQuery.data ?? [];

  // Auto-select connector: prefer initialConnector if in list, otherwise first available
  $: if (open && sameDriverConnectors.length > 0) {
    const isCurrentValid = selectedConnector && sameDriverConnectors.some((c) => c.name === selectedConnector?.name);
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
  let searchQuery = "";

  // Create selection-mode store with onToggleItem callback
  const selectionStore = new ConnectorExplorerStore(
    {
      allowSelectTable: true,
      allowContextMenu: false,
      allowNavigateToTable: false,
      allowShowSchema: false,
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
    searchQuery = "";
  }

  function handleSelectConnector(connector: V1AnalyzedConnector) {
    selectedConnector = connector;
    selectedTable = null; // Clear table selection when switching connectors
    searchQuery = ""; // Clear search when switching connectors
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
      // Error is handled by generateMetricsFromTable which shows a toast
    } finally {
      isGenerating = false;
    }
  }

  $: displayIcon = initialConnector
    ? connectorIconMapping[getConnectorIconKey(initialConnector)]
    : null;
  $: driverDisplayName =
    initialConnector?.driver?.displayName ?? initialConnector?.driver?.name ?? "OLAP";

  function handleAddNewConnector() {
    if (!initialConnector?.driver) return;
    const driver = initialConnector.driver;
    // Close this modal and open the connector form for the same driver type
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
        <span class="text-lg font-semibold">Select a table</span>
      </div>
      <p class="text-sm text-fg-secondary mt-1 font-normal">
        Choose a table to generate a metrics view
      </p>
    </Dialog.Title>

    <div class="flex {FORM_HEIGHT_DEFAULT}">
      <!-- Left panel: Connector list -->
      <div class="w-64 border-r border-border overflow-y-auto flex-shrink-0">
        <div class="p-2">
          <div class="text-xs font-medium text-fg-secondary uppercase tracking-wide px-2 py-1">
            Existing Connectors
          </div>
          <div class="text-xs font-medium text-fg-secondary tracking-wide px-2 py-1">
            Choose data from an existing connection or create a new connector.
          </div>
          {#each sameDriverConnectors as conn (conn.name)}
            <button
              class="w-full text-left px-2 py-1.5 rounded text-sm truncate {selectedConnector?.name === conn.name
                ? 'bg-primary-100 text-primary-600 dark:bg-primary-900 dark:text-primary-400 font-medium'
                : 'text-fg-primary hover:bg-surface-hover'}"
              on:click={() => handleSelectConnector(conn)}
            >
              {conn.name}
            </button>
          {/each}
          {#if sameDriverConnectors.length === 0}
            <div class="px-2 py-1 text-sm text-fg-secondary">
              No connectors found
            </div>
          {/if}

          <!-- Add new connector button -->
          <button
            class="w-full text-left px-2 py-1.5 rounded text-sm text-primary-500 hover:bg-surface-hover flex items-center gap-1.5 mt-2"
            on:click={handleAddNewConnector}
          >
            <Plus size="14" />
            New {driverDisplayName} connector
          </button>
        </div>
      </div>

      <!-- Right panel: Table browser -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Search input -->
        <div class="px-2 py-2 border-b border-border">
          <div class="relative flex items-center">
            <Search
              size="16"
              class="absolute left-2.5 text-fg-muted"
            />
            <input
              type="text"
              placeholder="Search tables..."
              bind:value={searchQuery}
              class="w-full pl-8 pr-3 py-1.5 bg-transparent border-none text-fg-primary placeholder:text-fg-muted focus:outline-none"
            />
          </div>
        </div>

        <!-- Table list -->
        <div class="flex-1 overflow-y-auto">
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
      <div class="flex items-center gap-3">
        <Button type="secondary" onClick={handleClose}>Back</Button>
      </div>

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
