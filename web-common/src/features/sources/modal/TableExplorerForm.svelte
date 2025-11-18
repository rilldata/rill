<script lang="ts">
  import ConnectorExplorer from "../../connectors/explorer/ConnectorExplorer.svelte";
  import { connectorExplorerStore } from "../../connectors/explorer/connector-explorer-store";

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
      if (table) {
        onSelect?.({
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

  <ConnectorExplorer {store} />
</section>
