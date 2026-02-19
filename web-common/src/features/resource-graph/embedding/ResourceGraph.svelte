<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import GraphCanvas from "../graph-canvas/GraphCanvas.svelte";
  import GraphOverlay from "./GraphOverlay.svelte";
  import {
    partitionResourcesByMetrics,
    partitionResourcesBySeeds,
    type ResourceGraphGrouping,
  } from "../graph-canvas/graph-builder";
  import {
    coerceResourceKind,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    expandSeedsByKind,
    isKindToken,
    tokenForKind,
    tokenForSeedString,
  } from "../navigation/seed-parser";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils";
  import SummaryGraph from "../summary/SummaryGraph.svelte";
  import { onDestroy } from "svelte";
  import { UI_CONFIG, FIT_VIEW_CONFIG } from "../shared/config";

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

  // New props for modularity
  export let onExpandedChange: ((id: string | null) => void) | null = null;
  export let expandedId: string | null = null; // Controlled mode
  export let overlayMode: "inline" | "fullscreen" | "modal" = "inline";
  export let gridColumns: number = UI_CONFIG.DEFAULT_GRID_COLUMNS;
  export let expandedHeightMobile: string = UI_CONFIG.EXPANDED_HEIGHT_MOBILE;
  export let expandedHeightDesktop: string = UI_CONFIG.EXPANDED_HEIGHT_DESKTOP;

  type SummaryMemo = {
    connectors: number;
    sources: number;
    metrics: number;
    models: number;
    dashboards: number;
    resources: V1Resource[];
    activeToken:
      | "connectors"
      | "metrics"
      | "sources"
      | "models"
      | "dashboards"
      | null;
  };
  function summaryEquals(a: SummaryMemo, b: SummaryMemo) {
    return (
      a.connectors === b.connectors &&
      a.sources === b.sources &&
      a.metrics === b.metrics &&
      a.models === b.models &&
      a.dashboards === b.dashboards &&
      a.resources === b.resources &&
      a.activeToken === b.activeToken
    );
  }

  // Fit view configuration for better centering
  export let fitViewPadding: number = FIT_VIEW_CONFIG.PADDING;
  export let fitViewMinZoom: number = FIT_VIEW_CONFIG.MIN_ZOOM;
  export let fitViewMaxZoom: number = FIT_VIEW_CONFIG.MAX_ZOOM;

  $: normalizedResources = resources ?? [];
  $: normalizedSeeds = expandSeedsByKind(
    seeds,
    normalizedResources,
    coerceResourceKind,
  );

  // Determine if we're filtering by a specific kind (e.g., ?kind=metrics)
  // This is used to filter out groups that don't contain any resource of the filtered kind
  $: filterKind = (function (): ResourceKind | undefined {
    const rawSeeds = seeds ?? [];
    // Only apply kind filter if all seeds are kind tokens (e.g., ["metrics"] or ["sources"])
    if (rawSeeds.length === 0) return undefined;
    for (const raw of rawSeeds) {
      const kind = isKindToken((raw || "").toLowerCase());
      if (!kind) return undefined; // Mixed seeds, no single kind filter
    }
    // All seeds are kind tokens - return the first one's kind
    return isKindToken((rawSeeds[0] || "").toLowerCase());
  })();

  // Determine which overview node should be highlighted based on current seeds
  $: overviewActiveToken = (function ():
    | "connectors"
    | "metrics"
    | "sources"
    | "models"
    | "dashboards"
    | null {
    const rawSeeds = seeds ?? [];
    for (const raw of rawSeeds) {
      const token = tokenForSeedString(raw);
      if (token) return token;
    }

    const normalized = normalizedSeeds ?? [];
    if (normalized.length) {
      const first = normalized[0];
      if (typeof first === "string") {
        return tokenForSeedString(first);
      } else {
        return tokenForKind(first.kind as ResourceKind | string | undefined);
      }
    }
    return null;
  })();

  $: resourceGroups =
    normalizedSeeds && normalizedSeeds.length
      ? partitionResourcesBySeeds(
          normalizedResources,
          normalizedSeeds,
          filterKind,
        )
      : partitionResourcesByMetrics(normalizedResources);
  $: visibleResourceGroups =
    typeof maxGroups === "number" && maxGroups >= 0
      ? resourceGroups.slice(0, maxGroups)
      : resourceGroups;
  $: hasGraphs = visibleResourceGroups.length > 0;

  // Brief loading indicator when URL seeds change (e.g., via Overview node clicks)
  let seedTransitionLoading = false;
  let seedTransitionTimer: ReturnType<typeof setTimeout> | null = null;

  // Cleanup timer on component destroy
  onDestroy(() => {
    if (seedTransitionTimer) {
      clearTimeout(seedTransitionTimer);
      seedTransitionTimer = null;
    }
  });

  // Compute resource counts for the summary graph header.
  // We compute directly in a single pass rather than using filter().length for performance.
  // This is more efficient (O(n) instead of O(4n)) and clearer in intent.
  $: ({
    connectorsCount,
    sourcesCount,
    modelsCount,
    metricsCount,
    dashboardsCount,
  } = (function computeCounts() {
    let connectors = 0,
      sources = 0,
      models = 0,
      metrics = 0,
      dashboards = 0;
    for (const r of normalizedResources) {
      if (r?.meta?.hidden) continue;
      const k = coerceResourceKind(r);
      if (!k) continue;
      if (k === ResourceKind.Connector) connectors++;
      else if (k === ResourceKind.Source) sources++;
      else if (k === ResourceKind.Model) models++;
      else if (k === ResourceKind.MetricsView) metrics++;
      else if (k === ResourceKind.Explore) dashboards++;
    }
    return {
      connectorsCount: connectors,
      sourcesCount: sources,
      modelsCount: models,
      metricsCount: metrics,
      dashboardsCount: dashboards,
    };
  })());

  // Memoization wrapper for summary data to avoid Svelte reactivity issues with Set/object equality.
  // Without this, the SummaryGraph component would re-render on every resource array change
  // even if counts haven't actually changed. The summaryEquals function does shallow comparison
  // of counts while checking resources array reference equality.
  let summaryMemo: SummaryMemo = {
    connectors: 0,
    sources: 0,
    models: 0,
    metrics: 0,
    dashboards: 0,
    resources: normalizedResources,
    activeToken: null,
  };
  $: {
    const nextSummary: SummaryMemo = {
      connectors: connectorsCount,
      sources: sourcesCount,
      metrics: metricsCount,
      models: modelsCount,
      dashboards: dashboardsCount,
      resources: normalizedResources,
      activeToken: overviewActiveToken,
    };
    // Only update memo if values actually changed (avoids unnecessary child re-renders)
    if (!summaryEquals(summaryMemo, nextSummary)) {
      summaryMemo = nextSummary;
    }
  }

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
    const errorCount = group.resources.filter(
      (r) => !!r.meta?.reconcileError,
    ).length;
    const anchor = anchorForGroup(group);
    const anchorError = !!anchor?.meta?.reconcileError;
    const labelWithCount = `${baseLabel} - ${count} resource${count === 1 ? "" : "s"}`;
    return { labelWithCount, errorCount, anchorError };
  }

  function groupRootNodeIds(
    group: ResourceGraphGrouping,
  ): string[] | undefined {
    const anchor = anchorForGroup(group);
    const anchorId = anchor ? resourceId(anchor) : group.id;
    return anchorId ? [anchorId] : undefined;
  }

  // Expansion state management with controlled/uncontrolled mode support
  let internalExpandedId: string | null = null;

  // Determine if we're in controlled mode (external expandedId prop provided)
  $: isControlledMode = expandedId !== null || onExpandedChange !== null;

  // Derive current expanded ID for template usage (computed from props/state)
  $: currentExpandedId = isControlledMode ? expandedId : internalExpandedId;

  // When the URL seeds change, re-open the first seeded graph in expanded view
  let lastSeedsSignature = "";
  $: areKindOnlySeeds =
    seeds && seeds.length
      ? seeds.every((s) => Boolean(isKindToken((s || "").toLowerCase())))
      : false;

  // Track URL sync state for the expanded query param.
  // undefined -> follow $page value, string/null -> explicit override from client-side history updates.
  let manualExpandedParam: string | null | undefined = undefined;
  let lastExpandedSyncUrl = "";

  // Pull expanded param from the current $page URL (decoded).
  $: expandedParamFromUrl = syncExpandedParam
    ? $page.url.searchParams.get("expanded") || null
    : null;

  // When the page URL actually changes (e.g., navigation), clear any manual override.
  $: if (syncExpandedParam) {
    const currentUrlString = $page.url.toString();
    if (currentUrlString !== lastExpandedSyncUrl) {
      lastExpandedSyncUrl = currentUrlString;
      manualExpandedParam = undefined;
    }
  } else {
    manualExpandedParam = undefined;
    lastExpandedSyncUrl = "";
  }

  // Effective expanded param includes manual overrides made via history.replaceState.
  $: effectiveExpandedParam = syncExpandedParam
    ? manualExpandedParam !== undefined
      ? manualExpandedParam
      : expandedParamFromUrl
    : null;

  // Auto-expand logic when seeds change
  $: {
    const signature = (seeds ?? []).join("|");
    if (signature !== lastSeedsSignature) {
      // Show a short loading state to indicate graphs are updating
      seedTransitionLoading = true;
      if (seedTransitionTimer) clearTimeout(seedTransitionTimer);
      seedTransitionTimer = setTimeout(
        () => (seedTransitionLoading = false),
        500,
      );

      lastSeedsSignature = signature;

      // Only auto-expand in uncontrolled mode
      const isUncontrolled = !isControlledMode;
      if (isUncontrolled && !effectiveExpandedParam) {
        if (
          seeds &&
          seeds.length &&
          visibleResourceGroups.length &&
          !areKindOnlySeeds
        ) {
          internalExpandedId = visibleResourceGroups[0]?.id ?? null;
        } else {
          internalExpandedId = null;
        }
      }
    }
  }

  // Sync with URL expanded param (in uncontrolled mode with URL sync enabled)
  $: if (
    !isControlledMode &&
    syncExpandedParam &&
    effectiveExpandedParam !== internalExpandedId
  ) {
    internalExpandedId = effectiveExpandedParam;
  }

  /**
   * Handle expansion change - calls callback or updates URL
   */
  function handleExpandChange(id: string | null) {
    // Update internal state if uncontrolled
    if (!isControlledMode) {
      internalExpandedId = id;
    }

    // Call user callback if provided
    if (onExpandedChange) {
      onExpandedChange(id);
    }
    // Otherwise sync with URL if enabled
    else if (syncExpandedParam) {
      setExpandedInUrl(id);
    }
  }

  function setExpandedInUrl(id: string | null) {
    manualExpandedParam = id;
    try {
      if (typeof window !== "undefined") {
        const currentUrl = new URL(window.location.href);
        if (id) {
          const newUrl = copyWithAdditionalArguments(
            currentUrl,
            { expanded: id },
            {},
          );
          window.history.replaceState(
            window.history.state,
            "",
            newUrl.toString(),
          );
        } else {
          const newUrl = copyWithAdditionalArguments(
            currentUrl,
            {},
            { expanded: true },
          );
          window.history.replaceState(
            window.history.state,
            "",
            newUrl.toString(),
          );
        }
        return;
      }
    } catch (error) {
      console.debug("[ResourceGraph] Skipped history update fallback", error);
    }
    // Fallback to SvelteKit navigation if direct history manipulation fails
    const currentUrl = new URL($page.url);
    if (id) {
      const newUrl = copyWithAdditionalArguments(
        currentUrl,
        { expanded: id },
        {},
      );
      goto(newUrl.pathname + newUrl.search, {
        replaceState: true,
        noScroll: true,
      });
    } else {
      const newUrl = copyWithAdditionalArguments(
        currentUrl,
        {},
        { expanded: true },
      );
      goto(newUrl.pathname + newUrl.search, {
        replaceState: true,
        noScroll: true,
      });
    }
  }
</script>

<div
  class="graph-root"
  style={`--graph-expanded-height-mobile:${expandedHeightMobile};--graph-expanded-height-desktop:${expandedHeightDesktop};`}
>
  {#if showSummary && currentExpandedId === null}
    <slot
      name="summary"
      connectors={connectorsCount}
      sources={sourcesCount}
      {metricsCount}
      {modelsCount}
      dashboards={dashboardsCount}
    >
      <div class="top-summary">
        <SummaryGraph
          connectors={summaryMemo.connectors}
          sources={summaryMemo.sources}
          metrics={summaryMemo.metrics}
          models={summaryMemo.models}
          dashboards={summaryMemo.dashboards}
          resources={summaryMemo.resources}
          activeToken={summaryMemo.activeToken}
        />
      </div>
      {#if hasGraphs}
        <div class="graph-section-title">All Graphs</div>
      {/if}
    </slot>
  {/if}

  {#if error}
    <div class="state error">
      <p>{error}</p>
    </div>
  {:else if isLoading || seedTransitionLoading}
    <div class="state">
      <div class="loading-state">
        <DelayedSpinner isLoading={true} size="1.5rem" />
        <p>{isLoading ? "Loading project graph..." : "Updating graphs..."}</p>
      </div>
    </div>
  {:else if !hasGraphs}
    <slot name="empty-state">
      <div class="state">
        <p>No resources found.</p>
      </div>
    </slot>
  {:else}
    {@const hasExpandedItem = currentExpandedId !== null}
    <div
      class="resource-graph-grid"
      class:has-expanded={hasExpandedItem}
      style:--grid-columns={gridColumns}
    >
      {#each visibleResourceGroups as group, index (group.id)}
        {@const isExpanded = currentExpandedId === group.id}
        {@const isHidden = hasExpandedItem && !isExpanded}
        <div
          class="grid-item"
          class:expanded={isExpanded}
          class:hidden={isHidden}
        >
          {#if isExpanded && overlayMode !== "inline"}
            <!-- Fullscreen or modal overlay -->
            <GraphOverlay
              {group}
              open={isExpanded}
              mode={overlayMode}
              {showControls}
              onClose={() => handleExpandChange(null)}
            />
          {:else if isExpanded}
            <!-- Inline expansion within grid -->
            <GraphCanvas
              flowId={group.id}
              resources={group.resources}
              title={null}
              titleLabel={showCardTitles
                ? groupTitleParts(group, index).labelWithCount
                : null}
              titleErrorCount={showCardTitles
                ? groupTitleParts(group, index).errorCount
                : null}
              anchorError={showCardTitles
                ? groupTitleParts(group, index).anchorError
                : false}
              rootNodeIds={groupRootNodeIds(group)}
              {showControls}
              showLock={false}
              fillParent={true}
              enableExpand={enableExpansion}
              {fitViewPadding}
              {fitViewMinZoom}
              {fitViewMaxZoom}
              onExpand={() => handleExpandChange(null)}
            />
          {:else}
            <!-- Collapsed card view -->
            <slot name="graph-item" {group} {index}>
              <GraphCanvas
                flowId={group.id}
                resources={group.resources}
                title={null}
                titleLabel={showCardTitles
                  ? groupTitleParts(group, index).labelWithCount
                  : null}
                titleErrorCount={showCardTitles
                  ? groupTitleParts(group, index).errorCount
                  : null}
                anchorError={showCardTitles
                  ? groupTitleParts(group, index).anchorError
                  : false}
                rootNodeIds={groupRootNodeIds(group)}
                showControls={false}
                showLock={true}
                fillParent={false}
                enableExpand={enableExpansion}
                {fitViewPadding}
                {fitViewMinZoom}
                {fitViewMaxZoom}
                onExpand={() => handleExpandChange(group.id)}
              />
            </slot>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  .graph-root {
    @apply relative h-full w-full overflow-auto flex flex-col min-h-0;
    --graph-card-height: 260px;
  }

  .resource-graph-grid {
    @apply grid gap-4 flex-1 min-h-0;
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }

  @media (min-width: 1024px) {
    .resource-graph-grid {
      grid-template-columns: repeat(var(--grid-columns, 3), minmax(0, 1fr));
    }
  }

  .grid-item {
    @apply relative;
  }

  .grid-item.hidden {
    display: none;
  }

  .grid-item.expanded {
    @apply col-span-full h-full;
  }

  .resource-graph-grid.has-expanded {
    @apply flex-1;
  }

  .state {
    @apply flex h-full w-full items-center justify-center text-sm text-fg-secondary;
  }

  .state.error {
    @apply text-red-500;
  }

  .loading-state {
    @apply flex items-center gap-x-3;
  }

  .top-summary {
    @apply mb-2;
  }
  .graph-section-title {
    @apply text-sm font-semibold text-fg-primary mt-4 mb-2;
  }
</style>
