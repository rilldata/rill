<script lang="ts">
  import {
    type ConnectorInfo,
    connectorInfoMap,
    connectors,
  } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
  import {
    connectorClassMapping,
    connectorIconMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import type { AddDataConfig } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import { Search } from "@rilldata/web-common/components/search";

  export let config: AddDataConfig;
  export let onSelect: (name: string) => void;
  export let onBack: () => void;

  const sortedConnectors = connectors
    .filter(
      (c) =>
        (config.importOnly ? true : c.name !== "duckdb") && c.category !== "ai",
    )
    .sort((a, b) => {
      if (a.name === "https" || a.name === "local_file") return 1;
      if (b.name === "https" || b.name === "local_file") return -1;
      return a.displayName.localeCompare(b.displayName);
    });

  let searchText = "";
  $: filteredConnectors = sortedConnectors.filter(
    (connector) =>
      connector.name.includes(searchText) ||
      connector.displayName.includes(searchText),
  );
</script>

<div class="source-selector">
  <div class="source-selector-header">
    <div class="source-selector-header-text">Where is your data?</div>
    <div class="grow"></div>
    <div class="w-80">
      <Search bind:value={searchText} />
    </div>
  </div>
  <div class="source-selector-content">
    <div class="source-selector-grid">
      {#each filteredConnectors as connector (connector.name)}
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
  <div class="source-selector-footer">
    {#if config.welcomeScreen}
      <Button type="secondary" onClick={onBack}>Back</Button>
    {/if}
  </div>
</div>

<style lang="postcss">
  .source-selector {
    @apply flex h-full flex-col;
  }

  .source-selector-content {
    @apply min-h-0 flex-1 overflow-auto;
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
