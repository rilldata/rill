<script lang="ts">
  import { connectors } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  export let startConnectorSelection: (name: string) => void;
</script>

<div class="source-selector">
  <div class="source-selector-header">
    <div class="source-selector-header-text">Where is your data?</div>
  </div>
  <div class="source-selector-grid">
    {#each connectors as connector (connector.name)}
      {@const icon = connectorIconMapping[connector.name]}
      <button
        class="source-selector-cell"
        on:click={() => startConnectorSelection(connector.name)}
      >
        <svelte:component this={icon} size="24px" />
        <span class="text-sm">{connector.displayName}</span>
      </button>
    {/each}
  </div>
</div>

<style lang="postcss">
  .source-selector {
    @apply flex flex-col gap-y-4 w-full;
    @apply bg-surface-background border rounded-lg shadow-sm;
  }

  .source-selector-header {
    @apply flex flex-row items-center justify-center gap-x-2 p-6;
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
