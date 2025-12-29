<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    type V1ConnectorDriver,
    createRuntimeServiceAnalyzeConnectors,
  } from "@rilldata/web-common/runtime-client";
  import ConnectorEntry from "../../connectors/explorer/ConnectorEntry.svelte";
  import { ConnectorExplorerStore } from "../../connectors/explorer/connector-explorer-store";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { featureFlags } from "../../feature-flags";

  export let onClose: () => void;
  export let connector: V1ConnectorDriver;

  const { ai } = featureFlags;

  let selectedTable: {
    connector: string;
    database: string;
    databaseSchema: string;
    table: string;
  } | null = null;

  $: ({ instanceId } = $runtime);

  // Get the connector name from the stored config (this is the name of the connector we just created)
  $: connectorName = connector.name ?? "";

  // Query for the analyzed connector to get full connector info
  $: analyzedConnectors = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      select: (data) => {
        // Filter to only the connector we just created
        const filtered = data.connectors?.filter(
          (c) => c.name === connectorName,
        );
        return { connectors: filtered };
      },
    },
  });

  $: analyzedConnector = $analyzedConnectors.data?.connectors?.[0];

  // Create a store that allows table selection
  const explorerStore = new ConnectorExplorerStore(
    {
      allowNavigateToTable: false,
      allowContextMenu: false,
      allowShowSchema: false,
      allowSelectTable: true,
      localStorage: false,
    },
    (connName, database, databaseSchema, table) => {
      if (table) {
        selectedTable = {
          connector: connName,
          database: database ?? "",
          databaseSchema: databaseSchema ?? "",
          table,
        };
      }
    },
  );

  $: createMetricsAction = selectedTable
    ? useCreateMetricsViewFromTableUIAction(
        instanceId,
        selectedTable.connector,
        selectedTable.database,
        selectedTable.databaseSchema,
        selectedTable.table,
        false, // createExplore - only generate metrics, not dashboard
        BehaviourEventMedium.Button,
        MetricsEventSpace.Modal,
      )
    : null;

  async function handleCreateMetrics() {
    if (createMetricsAction) {
      onClose();
      await createMetricsAction();
    }
  }
</script>

<div class="flex flex-col h-full">
  <div class="flex-1 overflow-y-auto p-6">
    <div class="mb-4">
      <h3 class="text-lg font-semibold text-gray-900 mb-2">
        Select a table to explore
      </h3>
      <p class="text-sm text-gray-600">
        Choose a table from your {connector.displayName ?? connector.name} connector
        to generate metrics and create a dashboard.
      </p>
    </div>

    <div
      class="border rounded-lg overflow-y-auto bg-white"
      style="max-height: 400px;"
    >
      {#if analyzedConnector}
        <ol class="px-0 pb-4">
          <ConnectorEntry connector={analyzedConnector} store={explorerStore} />
        </ol>
      {:else if $analyzedConnectors.isLoading}
        <div class="p-4 text-gray-500">Loading connector...</div>
      {:else}
        <div class="p-4 text-gray-500">No tables found.</div>
      {/if}
    </div>

    {#if selectedTable}
      <div class="p-3 rounded-lg">
        <p class="text-sm text-black">
          Selected: <span class="font-medium">{selectedTable.table}</span>
        </p>
      </div>
    {/if}
  </div>

  <div
    class="w-full bg-surface border-t border-gray-200 p-6 flex justify-between gap-2"
  >
    <Button onClick={onClose} type="secondary">Cancel</Button>

    <Button
      onClick={handleCreateMetrics}
      type="primary"
      disabled={!selectedTable}
    >
      {#if $ai}
        Create Metrics with AI
      {:else}
        Create Metrics
      {/if}
    </Button>
  </div>
</div>
