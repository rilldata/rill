<script lang="ts">
  import {
    V1AnalyzedConnector,
    createRuntimeServiceGetInstance,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import TableExplorer from "../tables/TableExplorer.svelte";
  import { connectorIconMapping } from "./connector-icon-mapping";

  export let connector: V1AnalyzedConnector;

  let showTables = true;

  $: ({ instanceId } = $runtime);
  $: instance = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: isOlapConnector = olapConnector === connector.name;
</script>

<li>
  <button on:click={() => (showTables = !showTables)}>
    <div class="flex-none">
      {#if connector.driver?.name}
        <svelte:component
          this={connectorIconMapping[connector.driver.name]}
          size="14px"
        />
      {/if}
    </div>
    <h4>{connector?.name}</h4>
    {#if isOlapConnector}
      <span class="text-xs text-gray-400"> (OLAP)</span>
    {/if}
  </button>
  {#if showTables}
    <TableExplorer {instanceId} {connector} />
  {/if}
</li>

<style lang="postcss">
  li {
    @apply flex flex-col;
  }

  button {
    @apply flex gap-x-1 items-center;
    @apply w-full pl-2 py-2;
  }

  button:hover {
    @apply bg-gray-200;
  }

  h4 {
    @apply text-xs font-medium;
  }
</style>
