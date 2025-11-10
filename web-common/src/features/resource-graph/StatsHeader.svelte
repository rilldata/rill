<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import type { V1Resource, V1ResourceName } from "@rilldata/web-common/runtime-client";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { partitionResourcesByMetrics, partitionResourcesBySeeds } from "./build-resource-graph";

  export let seeds: string[] | undefined;

  $: ({ instanceId } = $runtime);
  $: resourcesQuery = createRuntimeServiceListResources(instanceId, undefined, {
    query: { retry: 2, refetchOnMount: true, refetchOnWindowFocus: false, enabled: !!instanceId },
  });
  $: resources = $resourcesQuery.data?.resources ?? [];

  // Mirror sidebar logic to coerce models that are defined-as-source into Source for display
  function coerceKind(res: V1Resource): ResourceKind | undefined {
    const raw = res.meta?.name?.kind as ResourceKind | undefined;
    if (raw === ResourceKind.Model) {
      try {
        const name = res.meta?.name?.name;
        const resultTable = (res as any)?.model?.state?.resultTable;
        const definedAsSource = Boolean((res as any)?.model?.spec?.definedAsSource);
        if (name && name === resultTable && definedAsSource) return ResourceKind.Source;
      } catch {}
    }
    return raw;
  }

  const ALLOWED = new Set<ResourceKind>([
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
  ]);

  // Seed normalization (copied from ResourceGraph.svelte)
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
    if (idx === -1) return { kind: ResourceKind.MetricsView, name: s } as V1ResourceName;
    const kindPart = s.slice(0, idx);
    const namePart = s.slice(idx + 1);
    if (kindPart.includes(".")) return { kind: kindPart, name: namePart } as V1ResourceName;
    const mapped = KIND_ALIASES[kindPart.trim().toLowerCase()];
    return mapped ? ({ kind: mapped, name: namePart } as V1ResourceName) : s;
  }

  $: normalizedSeeds = (seeds ?? []).map((s) => normalizeSeed(s));

  // Counts
  $: allowedResources = resources.filter((r) => ALLOWED.has(coerceKind(r) as ResourceKind));
  $: sourcesCount = allowedResources.filter((r) => coerceKind(r) === ResourceKind.Source).length;
  $: modelsCount = allowedResources.filter((r) => coerceKind(r) === ResourceKind.Model).length;
  $: metricsCount = allowedResources.filter((r) => coerceKind(r) === ResourceKind.MetricsView).length;
  $: dashboardsCount = allowedResources.filter((r) => coerceKind(r) === ResourceKind.Explore).length;
  $: errorCount = allowedResources.filter((r) => !!r.meta?.reconcileError).length;

  // Graph grouping for counts
  $: groups = (normalizedSeeds && normalizedSeeds.length)
    ? partitionResourcesBySeeds(resources, normalizedSeeds)
    : partitionResourcesByMetrics(resources);
  $: graphsCount = groups.length;
  $: singleNodeGraphs = groups.filter((g) => (g.resources?.length ?? 0) === 1).length;
</script>

<div class="stats" aria-label="Project graph summary">
  <span>{sourcesCount} sources</span>
  <span>• {modelsCount} models</span>
  <span>• {metricsCount} metrics</span>
  <span>• {dashboardsCount} dashboards</span>
  <span>• {graphsCount} graphs</span>
  <span>• {singleNodeGraphs} singletons</span>
  {#if errorCount}
    <span class="errors">• {errorCount} errors</span>
  {/if}
  {#if $resourcesQuery.isLoading}
    <span class="loading">• loading…</span>
  {/if}
  {#if $resourcesQuery.error}
    <span class="errors">• failed to load</span>
  {/if}
  <!-- Align end marker for assistive tech -->
  <span class="sr-only">end of graph summary</span>
  </div>

<style lang="postcss">
  .stats {
    @apply ml-3 inline-flex items-center gap-x-2 text-xs text-gray-600;
  }
  .errors {
    @apply text-red-600;
  }
  .loading {
    @apply text-gray-400 italic;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
