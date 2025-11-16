<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceGraphCanvas from "./ResourceGraphCanvas.svelte";
  import {
    partitionResourcesByMetrics,
    partitionResourcesBySeeds,
    type ResourceGraphGrouping,
  } from "./build-resource-graph";
  import { coerceResourceKind, ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { expandSeedsByKind, isKindToken } from "./seed-utils";
  import { ALLOWED_FOR_GRAPH } from "./seed-utils";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils";
  import SummaryCountsGraph from "./SummaryCountsGraph.svelte";
  import { KIND_ALIASES } from "./seed-utils";

  export let resources: V1Resource[] | undefined;
  export let isLoading = false;
  export let error: string | null = null;
  export let seeds: string[] | undefined;
  export let syncExpandedParam = true;
  export let showSummary = true;
  export let showCardTitles = true;
  export let maxGroups: number | null = null;
  export let showControls = true;
  export let enableExpansion = true;

  $: normalizedResources = resources ?? [];
  $: normalizedSeeds = expandSeedsByKind(seeds, normalizedResources, coerceResourceKind);

  function tokenForKind(kind?: ResourceKind | string | null) {
    if (!kind) return null;
    const key = `${kind}`.toLowerCase();
    if (key.includes("source")) return "sources";
    if (key.includes("model")) return "models";
    if (key.includes("metricsview") || key.includes("metric")) return "metrics";
    if (key.includes("explore") || key.includes("dashboard")) return "dashboards";
    return null;
  }

  function tokenForSeedString(seed?: string | null) {
    if (!seed) return null;
    const normalized = seed.trim().toLowerCase();
    if (!normalized) return null;
    const tokenKind = isKindToken(normalized);
    if (tokenKind) return tokenForKind(tokenKind);
    const idx = normalized.indexOf(":");
    if (idx !== -1) {
      const kindPart = normalized.slice(0, idx);
      const mapped = KIND_ALIASES[kindPart];
      if (mapped) return tokenForKind(mapped);
      return tokenForKind(kindPart);
    }
    return null;
  }

  // Determine which overview node should be highlighted based on current seeds
  $: overviewActiveToken = (function () {
    const normalized = normalizedSeeds ?? [];
    if (normalized.length) {
      const first = normalized[0];
      if (typeof first === "string") {
        const token = tokenForSeedString(first);
        if (token) return token;
      } else {
        const token = tokenForKind(first.kind as ResourceKind | string | undefined);
        if (token) return token;
      }
    }
    return tokenForSeedString(seeds?.[0]) ?? null;
  })();

  $: resourceGroups = (normalizedSeeds && normalizedSeeds.length)
    ? partitionResourcesBySeeds(normalizedResources, normalizedSeeds)
    : partitionResourcesByMetrics(normalizedResources);
  $: visibleResourceGroups =
    typeof maxGroups === "number" && maxGroups >= 0
      ? resourceGroups.slice(0, maxGroups)
      : resourceGroups;
  $: hasGraphs = visibleResourceGroups.length > 0;
  $: singleGraphMode =
    !syncExpandedParam && maxGroups === 1 && visibleResourceGroups.length === 1;

  // Brief loading indicator when URL seeds change (e.g., via Overview node clicks)
  let seedTransitionLoading = false;
  let seedTransitionTimer: any = null;

  // Counts for the summary counts graph. Compute directly to avoid any cross-module Set equality pitfalls.
  $: ({ sourcesCount, modelsCount, metricsCount, dashboardsCount } = (function computeCounts() {
    let sources = 0, models = 0, metrics = 0, dashboards = 0;
    for (const r of normalizedResources) {
      if (r?.meta?.hidden) continue;
      const k = coerceResourceKind(r);
      if (!k) continue;
      if (k === ResourceKind.Source) sources++;
      else if (k === ResourceKind.Model) models++;
      else if (k === ResourceKind.MetricsView) metrics++;
      else if (k === ResourceKind.Explore) dashboards++;
    }
    return { sourcesCount: sources, modelsCount: models, metricsCount: metrics, dashboardsCount: dashboards };
  })());

  // Helpers to build title fragments with anchor error awareness
  function resourceId(res?: V1Resource | null): string | null {
    const kind = res?.meta?.name?.kind;
    const name = res?.meta?.name?.name;
    if (!kind || !name) return null;
    return `${kind}:${name}`;
  }

  function anchorForGroup(group: ResourceGraphGrouping): V1Resource | null {
    const rid = group.id;
    const found = group.resources.find((r) => resourceId(r) === rid);
    return found ?? null;
  }
  
  function groupTitleParts(group: ResourceGraphGrouping, index: number) {
    const baseLabel = group.label ?? `Graph ${index + 1}`;
    const count = group.resources.length;
    const errorCount = group.resources.filter((r) => !!r.meta?.reconcileError).length;
    const anchor = anchorForGroup(group);
    const anchorError = !!anchor?.meta?.reconcileError;
    const labelWithCount = `${baseLabel} - ${count} resource${count === 1 ? "" : "s"}`;
    return { labelWithCount, errorCount, anchorError };
  }

  function groupRootNodeIds(group: ResourceGraphGrouping): string[] | undefined {
    const anchor = anchorForGroup(group);
    const anchorId = anchor ? resourceId(anchor) : group.id;
    return anchorId ? [anchorId] : undefined;
  }

  // Expanded state (fills the graph-wrapper area, not fullscreen)
  let expandedGroup: ResourceGraphGrouping | null = null;
  let rootEl: HTMLDivElement | null = null;
  // Keep refs to each grid item so we can scroll it into view on expand
  const groupElMap = new Map<string, HTMLElement>();
  function registerGroupEl(node: HTMLElement, id: string) {
    groupElMap.set(id, node);
    return {
      destroy() {
        // Only delete if the same node (avoid race during diffing)
        if (groupElMap.get(id) === node) groupElMap.delete(id);
      },
    };
  }

  // When the URL seeds change, re-open the first seeded graph in expanded view
  let lastSeedsSignature = "";
  $: areKindOnlySeeds = (seeds && seeds.length)
    ? seeds.every((s) => Boolean(isKindToken((s || "").toLowerCase())))
    : false;
  // Read expanded param from URL
  $: expandedParam = syncExpandedParam
    ? $page.url.searchParams.get("expanded") || null
    : null;
  $: {
    const signature = (seeds ?? []).join("|");
    if (signature !== lastSeedsSignature) {
      // Show a short loading state to indicate graphs are updating
      seedTransitionLoading = true;
      if (seedTransitionTimer) clearTimeout(seedTransitionTimer);
      seedTransitionTimer = setTimeout(() => (seedTransitionLoading = false), 500);

      lastSeedsSignature = signature;
      // If seeds are kind-tokens only (e.g., metrics/sources/models/dashboards),
      // do not auto-expand. Only auto-expand when a specific resource seed is present.
      // But if URL has an explicit expanded param, that takes precedence and is handled below.
      if (!expandedParam) {
        if (
          seeds &&
          seeds.length &&
          visibleResourceGroups.length &&
          !areKindOnlySeeds
        ) {
          expandedGroup = visibleResourceGroups[0];
        } else {
          expandedGroup = null;
        }
      }
    }
  }

  // Apply expanded param from URL to select the group inline (only when it differs)
  $: if (syncExpandedParam && expandedParam && expandedParam !== expandedGroup?.id) {
    const match = visibleResourceGroups.find((g) => g.id === expandedParam);
    if (match) expandedGroup = match;
  }

  $: if (
    !expandedParam &&
    expandedGroup &&
    !visibleResourceGroups.find((g) => g.id === expandedGroup?.id)
  ) {
    expandedGroup = visibleResourceGroups[0] ?? null;
  }

  function setExpandedInUrl(id: string | null) {
    if (!syncExpandedParam) return;
    try {
      if (typeof window !== "undefined") {
        const currentUrl = new URL(window.location.href);
        if (id) {
          const newUrl = copyWithAdditionalArguments(currentUrl, { expanded: id }, {});
          window.history.replaceState(window.history.state, "", newUrl.toString());
        } else {
          const newUrl = copyWithAdditionalArguments(currentUrl, {}, { expanded: true });
          window.history.replaceState(window.history.state, "", newUrl.toString());
        }
        return;
      }
    } catch {}
    // Fallback to SvelteKit navigation if direct history manipulation fails
    const currentUrl = new URL($page.url);
    if (id) {
      const newUrl = copyWithAdditionalArguments(currentUrl, { expanded: id }, {});
      goto(newUrl.pathname + newUrl.search, { replaceState: true, noScroll: true });
    } else {
      const newUrl = copyWithAdditionalArguments(currentUrl, {}, { expanded: true });
      goto(newUrl.pathname + newUrl.search, { replaceState: true, noScroll: true });
    }
  }

  // No overlay: expanded graph renders inline within the grid
</script>

<div class="graph-root" bind:this={rootEl}>
  {#if showSummary}
    <div class="top-summary">
      <SummaryCountsGraph
        sources={sourcesCount}
        metrics={metricsCount}
        models={modelsCount}
        dashboards={dashboardsCount}
        resources={normalizedResources}
        activeToken={overviewActiveToken}
      />
    </div>
    {#if hasGraphs}
      <div class="graph-section-title">All Graphs</div>
    {/if}
  {/if}

  {#if error}
    <div class="state error">
      <p>{error}</p>
    </div>
  {:else if isLoading || seedTransitionLoading}
    <div class="state">
      <div class="loading-state">
        <DelayedSpinner isLoading={true} size="1.5rem" />
        <p>{isLoading ? 'Loading project graph...' : 'Updating graphs...'}</p>
      </div>
    </div>
  {:else if !hasGraphs}
    <div class="state">
      <p>No resources found.</p>
    </div>
  {:else}
    <div class={singleGraphMode ? 'graph-grid single' : 'graph-grid'}>
      {#each visibleResourceGroups as group, index (group.id)}
        <div class={expandedGroup?.id === group.id ? 'grid-item expanded' : 'grid-item'} use:registerGroupEl={group.id}>
          <ResourceGraphCanvas
            flowId={group.id}
            resources={group.resources}
            title={null}
            titleLabel={showCardTitles ? groupTitleParts(group, index).labelWithCount : null}
            titleErrorCount={showCardTitles ? groupTitleParts(group, index).errorCount : null}
            anchorError={showCardTitles ? groupTitleParts(group, index).anchorError : false}
            rootNodeIds={groupRootNodeIds(group)}
            showControls={showControls && expandedGroup?.id === group.id}
            showLock={expandedGroup?.id === group.id ? false : true}
            fillParent={expandedGroup?.id === group.id}
            enableExpand={enableExpansion}
            on:expand={enableExpansion
              ? async () => {
                  const willExpand = expandedGroup?.id !== group.id;
                  expandedGroup = willExpand ? group : null;
                  setExpandedInUrl(willExpand ? group.id : null);
                  if (willExpand) {
                    await Promise.resolve();
                    const el = groupElMap.get(group.id);
                    try {
                      el?.scrollIntoView({ behavior: 'smooth', block: 'start', inline: 'nearest' });
                    } catch {}
                  }
                }
              : undefined}
          />
        </div>
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  .graph-root {
    @apply relative h-full w-full overflow-auto;
  }

  .graph-grid {
    @apply grid gap-4;
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }

  @media (min-width: 1024px) {
    .graph-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
  }

  .graph-grid.single {
    @apply flex flex-col h-full;
  }

  .graph-grid.single .grid-item,
  .graph-grid.single .grid-item.expanded {
    @apply flex-1 h-full;
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

  .top-summary { @apply mb-2; }
  .graph-section-title { @apply text-sm font-semibold text-foreground mt-4 mb-2; }

  /* Inline expansion: span across all columns and set a taller height.
   * 700px on mobile provides adequate viewing space without excessive scrolling.
   * 860px on md+ screens accommodates larger displays and more complex graphs. */
  .grid-item.expanded {
    @apply col-span-full h-[700px] md:h-[860px];
  }
</style>
