<script lang="ts">
  import Search from "@rilldata/web-common/components/icons/Search.svelte";
  import ConnectorExplorer from "../../connectors/explorer/ConnectorExplorer.svelte";
  import {
    connectorExplorerStore,
    type ConnectorExplorerStore,
  } from "../../connectors/explorer/connector-explorer-store";

  // Create a local, selection-focused store for the modal (no nav/context menu)
  const store: ConnectorExplorerStore = connectorExplorerStore.duplicateStore();

  let search = "";
</script>

<section class="flex flex-col gap-3">
  <div class="flex flex-col gap-1">
    <h2 class="text-lg font-semibold">Table explorer</h2>
    <p class="text-slate-500 text-sm">
      Pick a table to power your first dashboard
    </p>
  </div>

  <div class="relative">
    <span class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">
      <Search size="16px" />
    </span>
    <input
      bind:value={search}
      placeholder="Search"
      class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-primary-200 focus:border-primary-300"
    />
  </div>

  <div class="border-t border-gray-200" />

  <!-- List of connectors/databases/schemas/tables -->
  <ConnectorExplorer {store} />
</section>
