<script lang="ts">
  import { page } from "$app/stores";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";
  import { countByKind } from "./overview-utils";

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  $: resources = useResources(instanceId);
  $: allResources = $resources.data?.resources ?? [];
  $: resourceCounts = countByKind(allResources);
</script>

{#if resourceCounts.length > 0}
  <section class="section">
    <div class="section-header">
      <h3 class="section-title">Resources</h3>
      <a href="{basePage}/resources" class="view-all">View all</a>
    </div>
    <div class="resource-chips">
      {#each resourceCounts as { kind, label, count } (kind)}
        <a href="{basePage}/resources?kind={kind}" class="resource-chip">
          {#if resourceIconMapping[kind]}
            <svelte:component this={resourceIconMapping[kind]} size="12px" />
          {/if}
          <span class="font-medium">{count}</span>
          <span class="text-fg-secondary">{label}{count !== 1 ? "s" : ""}</span>
        </a>
      {/each}
    </div>
  </section>
{/if}

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .view-all {
    @apply text-xs text-primary-500 no-underline;
  }
  .view-all:hover {
    @apply text-primary-600;
  }
  .resource-chips {
    @apply flex flex-wrap gap-2;
  }
  .resource-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle no-underline text-inherit;
  }
  .resource-chip:hover {
    @apply border-primary-500 text-primary-600;
  }
</style>
