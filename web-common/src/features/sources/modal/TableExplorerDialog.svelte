<script lang="ts">
  import ConnectorExplorer from "../../connectors/explorer/ConnectorExplorer.svelte";
  import { connectorExplorerStore } from "../../connectors/explorer/connector-explorer-store";
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

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
</script>

<section class="flex flex-col gap-3">
  <div class="flex flex-col gap-1">
    <h2 class="text-lg font-semibold">Table explorer</h2>
    <p class="text-slate-500 text-sm">
      Pick a table to power your first dashboard
    </p>
  </div>

  <div class="border-t border-gray-200" />

  <ConnectorExplorer {store} limitedConnectorDriver={connectorDriver} />
</section>
