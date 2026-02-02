<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import DataExplorerDialog from "./DataExplorerDialog.svelte";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { generateMetricsFromTable } from "../../metrics-views/ai-generation/generateMetricsView";

  export let connector: V1ConnectorDriver;
  export let connectorInstanceName: string | null = null;
  export let onModelCreated: (path: string) => void | Promise<void>;
  export let onBack: () => void;
  export let isGenerating = false;

  let selectedConnector = "";
  let selectedDatabase = "";
  let selectedSchema = "";
  let selectedTable = "";
  let generatingMetrics = false;

  // Keep the exported isGenerating in sync with the internal generatingMetrics state
  $: isGenerating = generatingMetrics;
  let instanceId: string;

  $: ({ instanceId } = $runtime);
  $: isOlapConnector = Boolean(connector?.implementsOlap);

  async function handleGenerateMetrics() {
    if (!selectedConnector || !selectedTable) return;
    try {
      generatingMetrics = true;
      await generateMetricsFromTable(
        instanceId,
        selectedConnector,
        selectedDatabase,
        selectedSchema,
        selectedTable,
        false, // Don't create explore dashboard, just metrics
        isOlapConnector,
      );
      // Close the modal after successful generation
      await onModelCreated("");
    } finally {
      generatingMetrics = false;
    }
  }
</script>

<div class="flex flex-col h-[600px]">
  <div class="flex flex-col flex-1 overflow-hidden">
    <DataExplorerDialog
      connectorDriver={connector}
      initialConnectorName={connectorInstanceName}
      onSelect={(detail) => {
        selectedConnector = detail.connector;
        selectedDatabase = detail.database;
        selectedSchema = detail.schema;
        selectedTable = detail.table;
      }}
    />
  </div>

  <div
    class="w-full bg-surface border-t border-gray-200 px-4 py-3 flex justify-between gap-2"
  >
    <Button onClick={onBack} type="secondary">Back</Button>

    <Button
      disabled={!selectedTable || generatingMetrics}
      loading={generatingMetrics}
      loadingCopy="Generating..."
      onClick={handleGenerateMetrics}
      type="primary"
    >
      Generate metrics with AI
    </Button>
  </div>
</div>
