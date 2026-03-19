<script lang="ts">
  import { connectors } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import {
    connectorClassMapping,
    connectorIconMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import type { AddDataConfig } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { Button } from "@rilldata/web-common/components/button";

  export let config: AddDataConfig;
  export let onSelect: (name: string) => void;
  export let onBack: () => void;

  const sortedConnectors = connectors
    .filter((c) => (config.importOnly ? true : c.name !== "duckdb"))
    .sort((a, b) => {
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
  <div class="source-selector-footer">
    {#if config.welcomeScreen}
      <Button type="secondary" onClick={onBack}>Back</Button>
    {/if}
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

  .source-selector-footer {
    @apply flex justify-between p-6 gap-2;
  }
</style>
