<script lang="ts">
  import {
    connectorClassMapping,
    connectorIconMapping,
  } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import type { AddDataConfig } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import { Search } from "@rilldata/web-common/components/search";
  import { ChevronRightIcon } from "lucide-svelte";
  import { getSupportedConnectorInfos } from "@rilldata/web-common/features/add-data/manager/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let config: AddDataConfig;
  export let onSelect: (name: string) => void;
  export let onBack: () => void;

  const runtimeClient = useRuntimeClient();

  const supportedConnectors = getSupportedConnectorInfos(runtimeClient, config);

  let searchText = "";
  $: searchTextLowerCase = searchText.toLowerCase();
  $: filteredConnectors = $supportedConnectors.filter(
    (connector) =>
      connector.name.toLowerCase().includes(searchTextLowerCase) ||
      connector.displayName.toLowerCase().includes(searchTextLowerCase) ||
      connector.category.toLowerCase().includes(searchTextLowerCase) ||
      connector.keywords.some((keyword) =>
        keyword.toLowerCase().includes(searchTextLowerCase),
      ),
  );
</script>

<div class="source-selector">
  <div class="source-selector-header">
    <div class="source-selector-header-text">Where is your data?</div>
    <div class="grow"></div>
    <div class="w-64">
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
          onclick={() => onSelect(connector.name)}
          aria-label={`Connect to ${connector.name}`}
        >
          <svelte:component this={icon} size="24px" class={className} />
          <span class="source-label">{connector.displayName}</span>
          <ChevronRightIcon size="16px" />
        </button>
      {:else}
        <div class="source-selector-no-matches">No matches found</div>
      {/each}
    </div>
  </div>
  {#if config.welcomeScreen}
    <div class="source-selector-footer">
      <Button type="secondary" onClick={onBack}>Back</Button>
    </div>
  {/if}
</div>

<style lang="postcss">
  .source-selector {
    @apply flex flex-col;
  }

  .source-selector-content {
    @apply min-h-0 flex-1 py-4 px-6 overflow-auto;
  }

  .source-selector-header {
    @apply flex flex-row items-center gap-x-2 py-4 px-6;
    @apply border-b;
  }

  .source-selector-header-text {
    @apply text-lg font-semibold;
  }

  .source-selector-grid {
    @apply grid grid-cols-3 grid-rows-7 gap-2;
  }

  .source-label {
    @apply text-left text-sm grow;
  }

  .source-selector-cell {
    @apply flex flex-row items-center gap-x-2 p-4;
    @apply bg-surface-overlay border rounded-lg shadow-sm;
  }

  .source-selector-footer {
    @apply flex justify-between pt-4 pb-6 px-6 gap-2 border-t;
  }

  .source-selector-no-matches {
    @apply h-[58px] text-sm text-fg-disabled;
  }
</style>
