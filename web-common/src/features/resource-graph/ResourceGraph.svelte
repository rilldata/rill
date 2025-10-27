<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceGraphCanvas from "./ResourceGraphCanvas.svelte";
  import {
    partitionResourcesByMetrics,
    partitionResourcesBySeeds,
    type ResourceGraphGrouping,
  } from "./build-resource-graph";
  import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let resources: V1Resource[] | undefined;
  export let isLoading = false;
  export let error: string | null = null;
  export let seeds: string[] | undefined;

  $: normalizedResources = resources ?? [];
  const KIND_ALIASES: Record<string, ResourceKind> = {
    metrics: ResourceKind.MetricsView,
    metric: ResourceKind.MetricsView,
    metricsview: ResourceKind.MetricsView,
    dashboard: ResourceKind.Explore,
    explore: ResourceKind.Explore,
    model: ResourceKind.Model,
    source: ResourceKind.Source,
    canvas: ResourceKind.Canvas,
  };

  function normalizeSeed(s: string): string | V1ResourceName {
    const idx = s.indexOf(":");
    if (idx === -1) {
      return { kind: ResourceKind.MetricsView, name: s };
    }
    const kindPart = s.slice(0, idx);
    const namePart = s.slice(idx + 1);
    if (kindPart.includes(".")) {
      return { kind: kindPart, name: namePart };
    }
    const mapped = KIND_ALIASES[kindPart.trim().toLowerCase()];
    if (mapped) return { kind: mapped, name: namePart };
    return s;
  }

  $: normalizedSeeds = (seeds ?? []).map((s) => normalizeSeed(s));

  $: resourceGroups = (normalizedSeeds && normalizedSeeds.length)
    ? partitionResourcesBySeeds(normalizedResources, normalizedSeeds)
    : partitionResourcesByMetrics(normalizedResources);
  $: hasGraphs = resourceGroups.length > 0;

  // Expanded state (fills the graph-wrapper area, not fullscreen)
  let expandedGroup: ResourceGraphGrouping | null = null;

  // When the URL seeds change, re-open the first seeded graph in expanded view
  let lastSeedsSignature = "";
  $: {
    const signature = (seeds ?? []).join("|");
    if (signature !== lastSeedsSignature) {
      lastSeedsSignature = signature;
      // If seeds are provided, open the first group; otherwise clear
      expandedGroup = (seeds && seeds.length && resourceGroups.length)
        ? resourceGroups[0]
        : null;
    }
  }

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
  <div class="graph-root">
    <div class={"graph-grid " + (expandedGroup ? 'blur-[1px] pointer-events-none' : '')}>
      {#each resourceGroups as group, index (group.id)}
        <ResourceGraphCanvas
          resources={group.resources}
          title={formatGroupTitle(group, index)}
          on:expand={() => (expandedGroup = group)}
        />
      {/each}
    </div>

    {#if expandedGroup}
      <div class="graph-overlay">
        <div class="graph-overlay-header">
          <div class="graph-overlay-title">{formatGroupTitle(expandedGroup, resourceGroups.findIndex((g) => g.id === expandedGroup?.id))}</div>
          <button class="overlay-close" on:click={() => (expandedGroup = null)} aria-label="Close expanded graph">✕</button>
        </div>
        <div class="graph-overlay-body">
          <ResourceGraphCanvas
            resources={expandedGroup.resources}
            title={null}
            showControls={true}
            enableExpand={false}
            fillParent={true}
          />
        </div>
      </div>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .graph-root {
    @apply relative h-full w-full overflow-auto;
  }

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

  .graph-overlay {
    @apply absolute inset-0 z-10 flex flex-col rounded-lg border border-gray-200 bg-white;
  }

  .graph-overlay-header {
    @apply flex items-center justify-between border-b border-gray-200 p-3;
  }

  .graph-overlay-title {
    @apply text-sm font-semibold text-foreground;
  }

  .overlay-close {
    @apply h-7 w-7 rounded border border-gray-300 bg-white text-sm text-gray-600 hover:bg-gray-50 hover:text-gray-800;
    line-height: 1.25rem;
  }

  .graph-overlay-body {
    @apply flex-1 overflow-hidden;
  }
</style>
