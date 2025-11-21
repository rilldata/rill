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
</script>

<div class="flex flex-col md:flex-row h-full">
  <!-- Left sidebar: existing connectors of the same type -->
  <aside
    class="w-full md:w-64 border-b md:border-b-0 md:border-r border-gray-200 bg-[#FAFAFA]"
  >
    <div class="p-4">
      {#if sidebar.length > 0}
        <h3 class="text-sm font-semibold text-gray-700">Existing connectors</h3>
        <p class="mt-1 text-xs text-slate-500">
          Choose data from an existing connector create a new one.
        </p>
      {:else}
        <h3 class="text-sm font-semibold text-gray-700">
          Connected successfully!
        </h3>
        <p class="mt-1 text-xs text-slate-500">
          Pick a table to power your first model.
        </p>
      {/if}
      <div class="mt-3 flex flex-col gap-2">
        {#if sidebar.length > 0}
          {#each sidebar as c (c.name)}
            <button
              class="w-full text-left text-sm px-3 py-2 rounded-md border border-gray-200 hover:bg-slate-50"
              aria-label={c.name}
              disabled
              title="Switching connectors not yet supported"
            >
              {c.name}
            </button>
          {/each}
        {:else}
          <span class="text-sm text-gray-500">No connectors found</span>
        {/if}
      </div>
    </div>
  </aside>

  <!-- Right content: data explorer -->
  <section class="flex-1 flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1">
      <h2 class="text-lg font-semibold">Data explorer</h2>
      <p class="text-slate-500 text-sm">Pick a table to power your model</p>
    </div>

    <div class="border border-gray-200 rounded-md overflow-y-auto py-2">
      <ConnectorExplorer {store} limitedConnectorDriver={connectorDriver} />
    </div>
  </section>
</div>
