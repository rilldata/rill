<script lang="ts">
  import ConnectorExplorer from "../../connectors/explorer/ConnectorExplorer.svelte";
  import { connectorExplorerStore } from "../../connectors/explorer/connector-explorer-store";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { createRuntimeServiceAnalyzeConnectors } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let connectorDriver: V1ConnectorDriver | undefined = undefined;
  export let onSelect:
    | ((detail: {
        connector: string;
        database: string;
        schema: string;
        table: string;
      }) => void)
    | undefined = undefined;

  const store = connectorExplorerStore.duplicateStore(
    (connector, database = "", schema = "", table = "") => {
      // Only emit selection when a table is toggled
      if (table && onSelect) {
        onSelect({
          connector,
          database,
          schema,
          table,
        });
      }
    },
  );

  // Sidebar: list existing connectors of the same driver type
  $: ({ instanceId } = $runtime);
  $: connectorsQuery = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      select: (data) => {
        if (!data?.connectors) return;
        const sameType = data.connectors
          .filter((c) => c?.driver?.name === connectorDriver?.name)
          .sort((a, b) => (a?.name as string).localeCompare(b?.name as string));
        return { connectors: sameType };
      },
    },
  });
  $: sidebar = $connectorsQuery?.data?.connectors ?? [];

  let selectedConnectorName: string | undefined = undefined;
  // Track known connector names to detect newly added connectors
  let knownConnectorNames: Set<string> = new Set();
  let hasInitializedSelection = false;
  $: if (sidebar && sidebar.length > 0) {
    // On first load, default to the first connector (existing behavior)
    if (!hasInitializedSelection) {
      selectedConnectorName = String(sidebar[0]?.name);
      knownConnectorNames = new Set(sidebar.map((c) => String(c?.name)));
      hasInitializedSelection = true;
    } else {
      // Detect any newly added connector by name and auto-select it
      const currentNames = new Set(sidebar.map((c) => String(c?.name)));
      let newlyAddedName: string | undefined = undefined;
      for (const c of sidebar) {
        const name = String(c?.name);
        if (!knownConnectorNames.has(name)) {
          newlyAddedName = name;
        }
      }
      knownConnectorNames = currentNames;
      if (newlyAddedName) {
        selectedConnectorName = newlyAddedName;
      }
    }
  }
</script>

<div class="flex flex-col md:flex-row h-full">
  <!-- Right content: data explorer -->
  <section class="flex-1 flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1">
      <h2 class="text-lg font-semibold">Data explorer</h2>
      <p class="text-slate-500 text-sm">Pick a table to create a model</p>
    </div>

    <div class="border border-gray-200 rounded-md overflow-y-auto py-2">
      <ConnectorExplorer
        {store}
        limitedConnectorDriver={connectorDriver}
        limitToConnector={selectedConnectorName}
      />
    </div>
  </section>
</div>
