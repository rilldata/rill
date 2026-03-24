<script lang="ts">
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceIconMapping,
    resourceLabelMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { resourceKindStyleName } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { Info } from "lucide-svelte";

  let open = false;

  const kinds: ResourceKind[] = [
    ResourceKind.Connector,
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
  ];
</script>

{#if open}
  <div class="legend-panel">
    <div class="legend-header">
      <span class="text-xs font-medium text-fg-primary">Legend</span>
      <button
        class="close-btn"
        aria-label="Close legend"
        on:click={() => (open = false)}
      >
        &times;
      </button>
    </div>
    <div class="legend-items">
      {#each kinds as kind}
        {@const icon = resourceIconMapping[kind]}
        {@const label = resourceLabelMapping[kind]}
        {@const styleName = resourceKindStyleName(kind)}
        {#if icon && label}
          <div class="legend-item">
            <span
              class="shrink-0 flex items-center px-1 py-0.5 rounded {styleName}"
            >
              <svelte:component this={icon} size="12px" />
            </span>
            <span class="text-xs text-fg-secondary">{label}</span>
          </div>
        {/if}
      {/each}
    </div>
  </div>
{:else}
  <button
    class="legend-toggle"
    aria-label="Show legend"
    on:click={() => (open = true)}
  >
    <Info size="14px" />
  </button>
{/if}

<style lang="postcss">
  .legend-panel {
    @apply absolute bottom-3 right-3 z-20 flex flex-col gap-y-1.5 rounded-lg border bg-surface-subtle px-3 py-2 shadow-sm;
  }

  .legend-header {
    @apply flex items-center justify-between gap-x-4;
  }

  .close-btn {
    @apply text-fg-muted text-sm leading-none px-0.5;
  }

  .close-btn:hover {
    @apply text-fg-primary;
  }

  .legend-items {
    @apply flex flex-col gap-y-1;
  }

  .legend-item {
    @apply flex items-center gap-x-2;
  }

  .legend-toggle {
    @apply absolute bottom-3 right-3 z-20 flex items-center justify-center h-7 w-7 rounded border bg-surface-subtle text-fg-secondary shadow-sm;
  }

  .legend-toggle:hover {
    @apply bg-surface-muted text-fg-primary;
  }
</style>
