<script lang="ts">
  import { page } from "$app/stores";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";
  import { countByKind, pluralizeKind } from "./overview-utils";
  import OverviewCard from "./OverviewCard.svelte";

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  $: resources = useResources(instanceId);
  $: allResources = $resources.data?.resources ?? [];
  $: resourceCounts = countByKind(allResources);
</script>

<OverviewCard title="Resources" viewAllHref="{basePage}/resources">
  {#if $resources.isLoading}
    <p class="text-sm text-fg-secondary">Loading resources...</p>
  {:else if resourceCounts.length > 0}
    <div class="chips">
      {#each resourceCounts as { kind, label, count } (kind)}
        <a href="{basePage}/resources?kind={kind}" class="chip">
          {#if resourceIconMapping[kind]}
            <svelte:component this={resourceIconMapping[kind]} size="12px" />
          {/if}
          <span class="font-medium">{count}</span>
          <span class="text-fg-secondary">{pluralizeKind(label, count)}</span>
        </a>
      {/each}
    </div>
  {:else}
    <p class="text-sm text-fg-secondary">No resources found.</p>
  {/if}
</OverviewCard>

<style lang="postcss">
  .chips {
    @apply flex flex-wrap gap-2;
  }
  .chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle no-underline text-inherit;
  }
  .chip:hover {
    @apply border-primary-500 text-primary-600;
  }
</style>
