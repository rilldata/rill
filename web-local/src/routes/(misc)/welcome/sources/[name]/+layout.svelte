<script lang="ts">
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import type { LayoutData } from "./$types";

  export let data: LayoutData;
  $: ({ connectorName, connectorDriver } = data);

  $: displayIcon =
    connectorIconMapping[connectorName] ??
    connectorIconMapping[connectorDriver?.name ?? ""];
  $: displayName = connectorDriver?.displayName ?? connectorName;
</script>

<div class="add-data-container">
  <div class="add-data-header">
    {#if displayIcon}
      <svelte:component this={displayIcon} size="18px" />
    {/if}
    <span class="text-lg leading-none font-semibold">{displayName}</span>
  </div>
  <div class="add-data-content">
    <slot />
  </div>
</div>

<style lang="postcss">
  .add-data-container {
    @apply flex flex-col w-[900px];
    @apply bg-surface-background border rounded-lg shadow-sm;
  }

  .add-data-header {
    @apply flex flex-row items-center px-6 py-4;
    @apply border-b;
  }

  .add-data-content {
    /* TODO: have AddDataForm footer be sticky and restrict the height */
  }
</style>
