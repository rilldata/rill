<script lang="ts">
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import { Tag } from "../../components/tag";
  import {
    V1AnalyzedConnector,
    createRuntimeServiceGetInstance,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { connectorIconMapping } from "./connector-icon-mapping";
  import DatabaseExplorer from "./olap/DatabaseExplorer.svelte";

  export let connector: V1AnalyzedConnector;

  let showDatabases = true;

  $: ({ instanceId } = $runtime);
  $: instance = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: isOlapConnector = olapConnector === connector.name;
</script>

{#if connector.name}
  <li aria-label={connector.name} class="connector-entry">
    <button
      class="connector-entry-header"
      on:click={() => (showDatabases = !showDatabases)}
    >
      <CaretDownIcon
        className="transform transition-transform text-gray-400 {showDatabases
          ? 'rotate-0'
          : '-rotate-90'}"
        size="14px"
      />
      <div class="flex-none">
        {#if connector.driver?.name}
          <svelte:component
            this={connectorIconMapping[connector.driver.name]}
            size="16px"
          />
        {/if}
      </div>
      <h4>{connector.name}</h4>
      <div class="flex-grow" />
      {#if isOlapConnector}
        <Tag height={16}>OLAP</Tag>
      {/if}
    </button>
    {#if showDatabases}
      <DatabaseExplorer {instanceId} {connector} />
    {/if}
  </li>
{/if}

<style lang="postcss">
  .connector-entry {
    @apply flex flex-col;
  }

  .connector-entry-header {
    @apply flex gap-x-1 items-center;
    @apply w-full p-1;
    @apply sticky top-0 z-10 bg-white;
  }

  button:hover {
    @apply bg-slate-100;
  }

  h4 {
    @apply text-xs font-medium;
  }
</style>
