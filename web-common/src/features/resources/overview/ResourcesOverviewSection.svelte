<script lang="ts">
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { pluralizeKind } from "@rilldata/web-common/features/resources/overview-utils";
  import type { ResourceCount } from "@rilldata/web-common/features/resources/overview-utils";

  export let resourceCounts: ResourceCount[];
  export let onViewAll: () => void;
  export let onChipClick: (kind: string) => void;
</script>

{#if resourceCounts.length > 0}
  <section class="section">
    <div class="section-header">
      <h3 class="section-title">Resources</h3>
      <button class="view-all" on:click={onViewAll}>View all</button>
    </div>
    <div class="resource-chips">
      {#each resourceCounts as { kind, label, count } (kind)}
        <button class="resource-chip" on:click={() => onChipClick(kind)}>
          {#if resourceIconMapping[kind]}
            <svelte:component this={resourceIconMapping[kind]} size="12px" />
          {/if}
          <span class="font-medium">{count}</span>
          <span class="text-fg-secondary">{pluralizeKind(label, count)}</span>
        </button>
      {/each}
    </div>
  </section>
{/if}

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5 text-left w-full;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .view-all {
    @apply text-xs text-primary-500 bg-transparent border-none cursor-pointer p-0;
  }
  .view-all:hover {
    @apply text-primary-600;
  }
  .resource-chips {
    @apply flex flex-wrap gap-2;
  }
  .resource-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle cursor-pointer;
  }
  .resource-chip:hover {
    @apply border-primary-500 text-primary-600;
  }
</style>
