<script lang="ts">
  import { page } from "$app/stores";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";
  import type { LayoutData } from "./$types";

  export let data: LayoutData;
  $: ({ connectorName, connectorDriver } = data);

  $: displayIcon =
    connectorIconMapping[connectorName] ??
    connectorIconMapping[connectorDriver?.name ?? ""];
  $: displayName = connectorDriver?.displayName ?? connectorName;

  $: onImportPage = $page.route.id?.endsWith("/import");
</script>

<div class="add-data-container {onImportPage ? 'w-[500px]' : 'w-[900px]'}">
  {#if !onImportPage}
    <div class="add-data-header">
      {#if displayIcon}
        <svelte:component this={displayIcon} size="18px" />
      {/if}
      <span class="text-lg leading-none font-semibold">{displayName}</span>
    </div>
  {/if}
  <div class="add-data-content">
    <slot />
  </div>
</div>

<style lang="postcss">
  .add-data-container {
    @apply flex flex-col;
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
