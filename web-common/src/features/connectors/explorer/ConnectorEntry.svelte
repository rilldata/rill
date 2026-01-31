<script lang="ts">
  import { Plus } from "lucide-svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { Tag } from "../../../components/tag";
  import {
    type V1AnalyzedConnector,
    createRuntimeServiceGetInstance,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import { connectorIconMapping } from "../connector-icon-mapping";
  import { getConnectorIconKey } from "../connectors-utils";
  import DatabaseExplorer from "./DatabaseExplorer.svelte";
  import { addSourceModal } from "../../sources/modal/add-source-visibility";

  export let connector: V1AnalyzedConnector;
  export let store: ConnectorExplorerStore;

  $: connectorName = connector?.name as string;
  $: expandedStore = store.getItem(connectorName);
  $: expanded = $expandedStore;
  $: ({ instanceId } = $runtime);
  $: instance = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: isOlapConnector = olapConnector === connector.name;
  $: implementsOlap = connector.driver?.implementsOlap;
  $: implementsSqlStore = connector.driver?.implementsSqlStore;
  $: implementsWarehouse = connector.driver?.implementsWarehouse;

  // Show connectors that can provide table browsing (OLAP, SQL stores, or warehouses)
  $: canBrowseTables =
    implementsOlap || implementsSqlStore || implementsWarehouse;
</script>

<!-- Show all connectors that support table browsing -->
{#if canBrowseTables}
  {#if connector.name}
    <li class="connector-entry">
      <div class="connector-entry-header text-fg-primary">
        <button
          class="connector-toggle"
          aria-label={connector.name}
          on:click={() => {
            store.toggleItem(connectorName);
          }}
        >
          <CaretDownIcon
            className="transform transition-transform text-fg-secondary {expanded
              ? 'rotate-0'
              : '-rotate-90'}"
            size="14px"
          />
          <span class="flex-none">
            {#if connector.driver?.name}
              <svelte:component
                this={connectorIconMapping[getConnectorIconKey(connector)]}
                size="16px"
              />
            {/if}
          </span>

          <h4>{connector.name}</h4>
        </button>

        {#if isOlapConnector}
          <Tag height={16} class="ml-auto">OLAP</Tag>
        {/if}
        {#if implementsOlap && connector.driver}
          <button
            class="add-model-button"
            aria-label="Add model from {connectorName}"
            title="Add model from this connector"
            on:click={() => {
              if (connector.driver) {
                addSourceModal.openExplorerForConnector(connector.driver);
              }
            }}
          >
            <Plus size="14" />
          </button>
        {/if}
      </div>

      {#if expanded}
        <DatabaseExplorer {instanceId} {connector} {store} />
      {/if}
    </li>
  {/if}
{/if}

<style lang="postcss">
  .connector-entry {
    @apply flex flex-col flex-none;
  }

  .connector-entry-header {
    @apply flex gap-x-1 items-center flex-none;
    @apply w-full pr-2 h-6;
    @apply z-10 bg-surface-subtle;
  }

  .connector-entry-header:hover {
    @apply bg-popover-accent;
  }

  .connector-toggle {
    @apply flex gap-x-1 items-center flex-grow;
    @apply h-full pl-2 outline-none;
    @apply cursor-pointer;
  }

  h4 {
    @apply text-xs font-medium;
  }

  .add-model-button {
    @apply flex items-center justify-center flex-none;
    @apply w-5 h-5 rounded;
    @apply text-fg-secondary hover:text-fg-primary;
    @apply hover:bg-gray-200;
    @apply transition-colors;
    @apply cursor-pointer;
  }
</style>
