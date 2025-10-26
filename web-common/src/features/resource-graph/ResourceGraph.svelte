<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceGraphCanvas from "./ResourceGraphCanvas.svelte";
  import {
    partitionResourcesByMetrics,
    type ResourceGraphGrouping,
  } from "./build-resource-graph";

  export let resources: V1Resource[] | undefined;
  export let isLoading = false;
  export let error: string | null = null;

  $: normalizedResources = resources ?? [];
  $: resourceGroups = partitionResourcesByMetrics(
    normalizedResources,
  );
  $: hasGraphs = resourceGroups.length > 0;

  const formatGroupTitle = (group: ResourceGraphGrouping, index: number) => {
    const baseLabel = group.label ?? `Graph ${index + 1}`;
    const count = group.resources.length;
    const errorCount = group.resources.filter((r) => !!r.meta?.reconcileError)
      .length;
    const errorSuffix = errorCount
      ? ` • ${errorCount} error${errorCount === 1 ? "" : "s"}`
      : "";
    return `${baseLabel} - ${count} resource${count === 1 ? "" : "s"}${errorSuffix}`;
  };
</script>

{#if isLoading}
  <div class="state">
    <div class="loading-state">
      <DelayedSpinner isLoading={isLoading} size="1.5rem" />
      <p>Loading project graph...</p>
    </div>
  </div>
{:else if error}
  <div class="state error">
    <p>{error}</p>
  </div>
{:else if !hasGraphs}
  <div class="state">
    <p>No resources found.</p>
  </div>
{:else}
  <div class="graph-grid">
    {#each resourceGroups as group, index (group.id)}
      <ResourceGraphCanvas
        resources={group.resources}
        title={formatGroupTitle(group, index)}
      />
    {/each}
  </div>
{/if}

<style lang="postcss">
  .graph-grid {
    @apply grid gap-6;
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }

  @media (min-width: 1024px) {
    .graph-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  .state {
    @apply flex h-full w-full items-center justify-center text-sm text-gray-500;
  }

  .state.error {
    @apply text-red-500;
  }

  .loading-state {
    @apply flex items-center gap-x-3;
  }
</style>
