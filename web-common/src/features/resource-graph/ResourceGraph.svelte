<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import ResourceGraphCanvas from "./ResourceGraphCanvas.svelte";
  import {
    partitionResourcesByMetrics,
    partitionResourcesBySeeds,
    type ResourceGraphGrouping,
  } from "./build-resource-graph";
  import { coerceResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { expandSeedsByKind, isKindToken } from "./seed-utils";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils";

  export let resources: V1Resource[] | undefined;
  export let isLoading = false;
  export let error: string | null = null;
  export let seeds: string[] | undefined;

  $: normalizedResources = resources ?? [];
  $: normalizedSeeds = expandSeedsByKind(seeds, normalizedResources, coerceResourceKind);

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

  /* Inline expansion: span across all columns and set a taller height.
   * 700px on mobile provides adequate viewing space without excessive scrolling.
   * 860px on md+ screens accommodates larger displays and more complex graphs. */
  .grid-item.expanded {
    @apply col-span-full h-[700px] md:h-[860px];
  }
</style>
