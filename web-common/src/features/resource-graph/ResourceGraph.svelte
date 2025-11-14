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
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";

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

  // Coerce models that are defined-as-source into Source for selection (matches graph display)
  function coerceKindForSelection(res: V1Resource): ResourceKind | undefined {
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

  const ALLOWED_FOR_GRAPH = new Set<ResourceKind>([
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
  ]);

  function isKindToken(s: string): ResourceKind | undefined {
    const key = s.trim().toLowerCase();
    switch (key) {
      case "metrics":
      case "metric":
      case "metricsview":
        return ResourceKind.MetricsView;
      case "dashboards":
      case "dashboard":
      case "explore":
      case "explores":
        return ResourceKind.Explore;
      case "models":
      case "model":
        return ResourceKind.Model;
      case "sources":
      case "source":
        return ResourceKind.Source;
      default:
        return undefined;
    }
  }

  function expandSeedsByKind(
    seedStrings: string[] | undefined,
    resList: V1Resource[],
  ): (string | V1ResourceName)[] {
    const input = seedStrings ?? [];
    const expanded: (string | V1ResourceName)[] = [];
    const seen = new Set<string>(); // de-dupe by id "kind:name"

    // Helper to push a normalized seed and avoid duplicates
    const pushSeed = (s: string | V1ResourceName) => {
      const id = typeof s === "string" ? s : `${s.kind}:${s.name}`;
      if (seen.has(id)) return;
      seen.add(id);
      expanded.push(s);
    };

    // Visible resources only, to align with graph rendering
    const visible = resList.filter(
      (r) => ALLOWED_FOR_GRAPH.has(coerceKindForSelection(r) as ResourceKind) && !r.meta?.hidden,
    );

    for (const raw of input) {
      if (!raw) continue;
      if (raw.includes(":")) {
        // Explicit seed, keep as-is after normalization
        pushSeed(normalizeSeed(raw));
        continue;
      }
      const kindToken = isKindToken(raw);
      if (!kindToken) {
        // Name-only, defaults to metrics view name
        pushSeed(normalizeSeed(raw));
        continue;
      }
      // Expand: one seed per visible resource of this kind
      for (const r of visible) {
        if (coerceKindForSelection(r) !== kindToken) continue;
        const name = r.meta?.name?.name;
        const kind = r.meta?.name?.kind; // use actual runtime kind for matching ids
        if (!name || !kind) continue;
        pushSeed({ kind, name });
      }
    }

    return expanded;
  }

  $: normalizedSeeds = expandSeedsByKind(seeds, normalizedResources);

  $: resourceGroups = (normalizedSeeds && normalizedSeeds.length)
    ? partitionResourcesBySeeds(normalizedResources, normalizedSeeds)
    : partitionResourcesByMetrics(normalizedResources);
  $: hasGraphs = resourceGroups.length > 0;

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
  $: expandedParam = $page.url.searchParams.get("expanded") || null;
  $: {
    const signature = (seeds ?? []).join("|");
    if (signature !== lastSeedsSignature) {
      lastSeedsSignature = signature;
      // If seeds are kind-tokens only (e.g., metrics/sources/models/dashboards),
      // do not auto-expand. Only auto-expand when a specific resource seed is present.
      // But if URL has an explicit expanded param, that takes precedence and is handled below.
      if (!expandedParam) {
        if (seeds && seeds.length && resourceGroups.length && !areKindOnlySeeds) {
          expandedGroup = resourceGroups[0];
        } else {
          expandedGroup = null;
        }
      }
    }
  }

  // Apply expanded param from URL to select the group inline (only when it differs)
  $: {
    if (expandedParam && expandedParam !== expandedGroup?.id) {
      const match = resourceGroups.find((g) => g.id === expandedParam);
      if (match) expandedGroup = match;
    }
  }

  function setExpandedInUrl(id: string | null) {
    try {
      if (typeof window !== "undefined") {
        const url = new URL(window.location.href);
        if (!id) url.searchParams.delete("expanded");
        else url.searchParams.set("expanded", id);
        const qs = url.searchParams.toString();
        const newUrl = qs ? `${url.pathname}?${qs}${url.hash ?? ''}` : `${url.pathname}${url.hash ?? ''}`;
        window.history.replaceState(window.history.state, "", newUrl);
        return;
      }
    } catch {}
    // Fallback to SvelteKit navigation if direct history manipulation fails
    const url = new URL($page.url);
    if (!id) url.searchParams.delete("expanded");
    else url.searchParams.set("expanded", id);
    const qs = url.searchParams.toString();
    const newPath = qs ? `${url.pathname}?${qs}` : url.pathname;
    goto(newPath, { replaceState: true, noScroll: true });
  }

  // No overlay: expanded graph renders inline within the grid
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
  <div class="graph-root" bind:this={rootEl}>
    <div class="graph-grid">
      {#each resourceGroups as group, index (group.id)}
        <div class={expandedGroup?.id === group.id ? 'grid-item expanded' : 'grid-item'} use:registerGroupEl={group.id}>
          <ResourceGraphCanvas
            flowId={group.id}
            resources={group.resources}
            title={null}
            titleLabel={groupTitleParts(group, index).labelWithCount}
            titleErrorCount={groupTitleParts(group, index).errorCount}
          anchorError={groupTitleParts(group, index).anchorError}
            showControls={expandedGroup?.id === group.id}
            showLock={expandedGroup?.id === group.id ? false : true}
            fillParent={expandedGroup?.id === group.id}
            on:expand={async () => {
              // Toggle inline expansion and sync to URL
              const willExpand = expandedGroup?.id !== group.id;
              expandedGroup = willExpand ? group : null;
              setExpandedInUrl(willExpand ? group.id : null);
              // After DOM updates, scroll expanded item to top of the scroll container
              if (willExpand) {
                await Promise.resolve();
                const el = groupElMap.get(group.id);
                try {
                  el?.scrollIntoView({ behavior: 'smooth', block: 'start', inline: 'nearest' });
                } catch {}
              }
            }}
          />
        </div>
      {/each}
    </div>
  </div>
{/if}

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

  .state {
    @apply flex h-full w-full items-center justify-center text-sm text-gray-500;
  }

  .state.error {
    @apply text-red-500;
  }

  .loading-state {
    @apply flex items-center gap-x-3;
  }

  /* Inline expansion: span across all columns and set a taller height */
  .grid-item.expanded {
    @apply col-span-full h-[700px] md:h-[860px];
  }
</style>
