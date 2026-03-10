<script lang="ts">
  import { connectors } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import {
    connectorClassMapping,
    connectorIconMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  export let onSelect: (name: string) => void;

  const sortedConnectors = connectors.sort((a, b) => {
    if (a.name === "https" || a.name === "local_file") return 1;
    if (b.name === "https" || b.name === "local_file") return -1;
    return a.displayName.localeCompare(b.displayName);
  });
</script>

<div class="source-selector">
  <div class="source-selector-header">
    <div class="source-selector-header-text">Where is your data?</div>
  </div>
  <div class="source-selector-grid">
    {#each sortedConnectors as connector (connector.name)}
      {@const icon = connectorIconMapping[connector.name]}
      {@const className = connectorClassMapping[connector.name] ?? ""}
      <button
        class="source-selector-cell"
        on:click={() => onSelect(connector.name)}
      >
        <svelte:component this={icon} size="24px" class={className} />
        <span class="text-sm">{connector.displayName}</span>
      </button>
    {/each}
  </div>
</div>

<style lang="postcss">
  .source-selector {
  }

  .source-selector-header {
    @apply flex flex-row items-center gap-x-2 p-2.5 px-6;
    @apply border-b;
  }

  .source-selector-header-text {
    @apply text-lg font-semibold;
  }

  .source-selector-grid {
    @apply grid grid-cols-3 p-6 gap-2;
  }

  .source-selector-cell {
    @apply flex flex-row items-center gap-x-2 p-4;
    @apply bg-surface-overlay border rounded-lg shadow-sm;
  }
</style>
