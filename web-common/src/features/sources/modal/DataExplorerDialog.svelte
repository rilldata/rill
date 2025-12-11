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
  <!-- Left sidebar: existing connectors of the same type -->
  <aside
    class="w-full md:w-64 h-full overflow-y-auto md:border-r border-gray-200 bg-[#FAFAFA]"
  >
    <div class="sticky top-0 z-10 bg-[#FAFAFA] p-4 border-gray-200">
      {#if sidebar.length > 0}
        <h3 class="text-lg font-semibold">Existing connectors</h3>
        <p class="mt-4 text-sm text-muted-foreground">
          Choose data from an existing connector or create a new one.
        </p>
      {:else}
        <h3 class="text-lg font-semibold">Connected successfully!</h3>
        <p class="mt-1 text-sm text-muted-foreground">
          Pick a table to create a model.
        </p>
      {/if}
    </div>
    <div class="first:pt-0 px-4 pb-4">
      <div class="flex flex-col gap-4">
        {#if sidebar.length > 0}
          {#each sidebar as c (c.name)}
            <button
              class="w-full text-left text-sm px-4 py-3 h-[40px] font-medium text-foreground rounded-[10px] border transition-colors flex items-center gap-4 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-300"
              class:border-gray-200={selectedConnectorName !== c.name}
              class:border-indigo-500={selectedConnectorName === c.name}
              class:bg-indigo-50={selectedConnectorName === c.name}
              aria-label={c.name}
              on:click={() => {
                selectedConnectorName = String(c.name);
              }}
            >
              <span
                class="inline-flex items-center justify-center h-4 w-4 rounded-full border-2"
                class:border-gray-300={selectedConnectorName !== c.name}
                class:border-indigo-500={selectedConnectorName === c.name}
                class:bg-indigo-500={selectedConnectorName === c.name}
                aria-hidden="true"
              >
                <span
                  class="h-2 w-2 rounded-full"
                  class:bg-transparent={selectedConnectorName !== c.name}
                  class:bg-white={selectedConnectorName === c.name}
                />
              </span>
              <span class="text-base text-gray-900">{c.name}</span>
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
