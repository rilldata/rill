<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import DataExplorerDialog from "./DataExplorerDialog.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createModelFromExplorerSelection } from "./model-creation-utils";
  import { useIsModelingSupportedForConnectorOLAP as useIsModelingSupportedForConnector } from "../../connectors/selectors";

  export let connector: V1ConnectorDriver;
  export let onModelCreated: (path: string) => void | Promise<void>;
  export let onBack: () => void;
  export let formHeight: string = "";

  let selectedConnector = "";
  let selectedDatabase = "";
  let selectedSchema = "";
  let selectedTable = "";
  let creatingModel = false;
  let instanceId: string;

  $: ({ instanceId } = $runtime);
  $: modelingSupportQuery = useIsModelingSupportedForConnector(
    instanceId,
    selectedConnector || "",
  );
  $: isModelingSupportedForSelected = $modelingSupportQuery.data || false;

  async function handleCreateModel() {
    if (!selectedConnector || !selectedTable) return;
    try {
      creatingModel = true;
      const [newModelPath] = await createModelFromExplorerSelection(
        queryClient,
        {
          connector: selectedConnector,
          database: selectedDatabase,
          schema: selectedSchema,
          table: selectedTable,
          isModelingSupported: isModelingSupportedForSelected,
        },
      );
      await onModelCreated(newModelPath);
    } finally {
      creatingModel = false;
    }
  }
</script>

<div class="flex flex-col flex-grow h-full">
  <!-- Header with connector name -->
  <div class="px-6 pt-6 pb-4 border-b border-gray-200">
    <h2 class="text-lg font-semibold text-gray-900">
      {connector.displayName ?? connector.name}
      <span class="text-gray-500 font-normal">{connector.name}</span>
    </h2>
  </div>

  <div class={cn("flex flex-col flex-grow overflow-hidden p-0", formHeight)}>
    <DataExplorerDialog
      connectorDriver={connector}
      onSelect={(detail) => {
        selectedConnector = detail.connector;
        selectedDatabase = detail.database;
        selectedSchema = detail.schema;
        selectedTable = detail.table;
      }}
    />
  </div>

  <div
    class="w-full bg-surface border-t border-gray-200 p-6 flex justify-between gap-2"
  >
    <Button onClick={onBack} type="secondary">Back</Button>

    <Button
      disabled={!selectedTable || creatingModel}
      loading={creatingModel}
      loadingCopy="Generating..."
      onClick={handleCreateModel}
      type="primary"
    >
      Generate metrics with AI
    </Button>
  </div>
</div>
