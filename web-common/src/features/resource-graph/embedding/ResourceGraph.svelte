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
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import {
    ALLOWED_FOR_GRAPH,
    expandSeedsByKind,
    isKindToken,
    normalizeSeed,
    tokenForKind,
    tokenForSeedString,
    type KindToken,
  } from "../navigation/seed-parser";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils";
  import ResourceNodeSelector from "../summary/ResourceNodeSelector.svelte";
  import { onDestroy } from "svelte";
  import { UI_CONFIG, FIT_VIEW_CONFIG } from "../shared/config";
  import type {
    ResourceStatusFilter,
    ResourceStatusFilterValue,
  } from "../shared/types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { RefreshCw } from "lucide-svelte";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";

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
  export let searchQuery = "";
  export let statusFilter: ResourceStatusFilter = [];
  export let showNodeActions = true;

  // New props for modularity
  export let onExpandedChange: ((id: string | null) => void) | null = null;
  export let expandedId: string | null = null; // Controlled mode
  export let overlayMode: "inline" | "fullscreen" | "modal" = "inline";
  export let gridColumns: number = UI_CONFIG.DEFAULT_GRID_COLUMNS;
  export let expandedHeightMobile: string = UI_CONFIG.EXPANDED_HEIGHT_MOBILE;
  export let expandedHeightDesktop: string = UI_CONFIG.EXPANDED_HEIGHT_DESKTOP;

  // Sidebar layout mode
  export let layout: "grid" | "sidebar" = "grid";
  export let selectedGroupId: string | null = null;
  export let onSelectedGroupChange: ((id: string | null) => void) | null = null;

  // Toolbar callbacks (sidebar layout)
  export let onRefreshAll: (() => void) | null = null;
  export let statusFilterOptions: {
    label: string;
    value: ResourceStatusFilterValue;
  }[] = [];
  export let onStatusToggle:
    | ((value: ResourceStatusFilterValue) => void)
    | null = null;
  export let onClearFilters: (() => void) | null = null;

  type SummaryMemo = {
    connector: number;
    sources: number;
    models: number;
    metrics: number;
    dashboards: number;
    resources: V1Resource[];
    activeToken: KindToken | null;
  };
  function summaryEquals(a: SummaryMemo, b: SummaryMemo) {
    return (
      a.connector === b.connector &&
      a.sources === b.sources &&
      a.models === b.models &&
      a.metrics === b.metrics &&
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

  // Derive active resource ID for the node selector dropdown.
  // If there's exactly one non-kind-token seed, use its fully qualified ID.
  $: activeResourceIdForSelector = (function (): string | null {
    if (!normalizedSeeds || normalizedSeeds.length !== 1) return null;
    const first = normalizedSeeds[0];
    if (typeof first === "string") return null;
    return first.kind && first.name ? `${first.kind}:${first.name}` : null;
  })();

  // Determine if we're filtering by a specific kind (e.g., ?kind=metrics)
  // This is used to filter out groups that don't contain any resource of the filtered kind
  // Special case: "dashboards" includes both Explore and Canvas
  $: filterKind = (function (): ResourceKind | "dashboards" | undefined {
    const rawSeeds = seeds ?? [];
    // Only apply kind filter if all seeds are kind tokens (e.g., ["metrics"] or ["sources"])
    if (rawSeeds.length === 0) return undefined;
    for (const raw of rawSeeds) {
      const kind = isKindToken((raw || "").toLowerCase());
      if (!kind) return undefined; // Mixed seeds, no single kind filter
    }
    // Check if it's the dashboards token (which includes both Explore and Canvas)
    const firstSeed = (rawSeeds[0] || "").toLowerCase();
    if (firstSeed === "dashboards" || firstSeed === "dashboard") {
      return "dashboards"; // Special token to indicate both Explore and Canvas
    }
    // All seeds are kind tokens - return the first one's kind
    return isKindToken(firstSeed);
  })();

  // Determine which overview node should be highlighted based on current seeds
  // For Canvas with MetricsView seeds, prioritize the Canvas token (dashboards) over MetricsView tokens
  $: overviewActiveToken = (function (): KindToken | null {
    const rawSeeds = seeds ?? [];

    // Check the first seed first - this should be the anchor resource (e.g., Canvas)
    // This ensures Canvas/Explore tokens are prioritized over MetricsView tokens
    if (rawSeeds.length > 0) {
      const firstToken = tokenForSeedString(rawSeeds[0]);
      if (firstToken) return firstToken;
    }

    // Fall back to checking all seeds if first seed didn't yield a token
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

  // If seeds were provided (e.g., kind filter like "models"), use seed-based partitioning.
  // When the kind has no resources, normalizedSeeds will be empty — return [] instead of
  // falling back to partitionResourcesByMetrics, which would show unrelated metric views.
  $: hasExplicitSeeds = seeds && seeds.length > 0;
  $: resourceGroups =
    normalizedSeeds && normalizedSeeds.length
      ? partitionResourcesBySeeds(
          normalizedResources,
          normalizedSeeds,
          filterKind,
        )
      : hasExplicitSeeds
        ? []
        : partitionResourcesByMetrics(normalizedResources);
  // Filter groups by search query and status
  $: filteredResourceGroups = (() => {
    let groups = resourceGroups;

    // Filter by search query (matches resource names)
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase().trim();
      groups = groups.filter((group) =>
        group.resources.some((r) =>
          r.meta?.name?.name?.toLowerCase().includes(query),
        ),
      );
    }

    // Filter by status (multi-select: show groups matching ANY selected status)
    if (statusFilter.length > 0) {
      groups = groups.filter((group) =>
        group.resources.some((r) => {
          if (
            statusFilter.includes("pending") &&
            r.meta?.reconcileStatus &&
            r.meta.reconcileStatus !== "RECONCILE_STATUS_IDLE"
          ) {
            return true;
          }
          if (statusFilter.includes("errored") && !!r.meta?.reconcileError) {
            return true;
          }
          if (
            statusFilter.includes("ok") &&
            r.meta?.reconcileStatus === "RECONCILE_STATUS_IDLE" &&
            !r.meta?.reconcileError
          ) {
            return true;
          }
          return false;
        }),
      );
    }

    return groups;
  })();

  $: visibleResourceGroups =
    typeof maxGroups === "number" && maxGroups >= 0
      ? filteredResourceGroups.slice(0, maxGroups)
      : filteredResourceGroups;
  $: hasGraphs = visibleResourceGroups.length > 0;

  // Whether any filters are active (kind, status, or tree search)
  $: hasActiveFilters =
    statusFilter.length > 0 || treeSearchQuery.trim().length > 0;

  function handleClearFilters() {
    treeSearchQuery = "";
    onClearFilters?.();
  }

  // --- Sidebar selection state ---
  let treeSearchQuery = "";

  // All resources organized by kind for the tree dropdown
  type ResourceDropdownEntry = {
    name: string;
    kind: ResourceKind;
    status: "ok" | "pending" | "errored";
  };
  type ResourceDropdownSection = {
    kind: ResourceKind;
    label: string;
    entries: ResourceDropdownEntry[];
  };

  const DROPDOWN_SECTION_ORDER: ResourceKind[] = [
    ResourceKind.Connector,
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
  ];

  const DROPDOWN_SECTION_LABELS: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.Connector]: "OLAP Connector",
    [ResourceKind.Source]: "Source Models",
    [ResourceKind.Model]: "Models",
    [ResourceKind.MetricsView]: "Metric Views",
    [ResourceKind.Explore]: "Explore Dashboards",
    [ResourceKind.Canvas]: "Canvas Dashboards",
  };

  function getResourceStatus(r: V1Resource): "ok" | "pending" | "errored" {
    if (r.meta?.reconcileError) return "errored";
    if (
      r.meta?.reconcileStatus &&
      r.meta.reconcileStatus !== "RECONCILE_STATUS_IDLE"
    )
      return "pending";
    return "ok";
  }

  $: allResourceSections = (function (): ResourceDropdownSection[] {
    const grouped = new Map<ResourceKind, ResourceDropdownEntry[]>();

    for (const r of normalizedResources) {
      const kind = coerceResourceKind(r);
      if (!kind || !ALLOWED_FOR_GRAPH.has(kind)) continue;
      if (r.meta?.hidden && kind !== ResourceKind.Connector) continue;
      const name = r.meta?.name?.name;
      if (!name) continue;

      const entries = grouped.get(kind) ?? [];
      entries.push({ name, kind, status: getResourceStatus(r) });
      grouped.set(kind, entries);
    }

    const result: ResourceDropdownSection[] = [];
    for (const kind of DROPDOWN_SECTION_ORDER) {
      const entries = grouped.get(kind);
      if (!entries?.length) continue;
      entries.sort((a, b) => a.name.localeCompare(b.name));
      result.push({
        kind,
        label: DROPDOWN_SECTION_LABELS[kind] ?? kind,
        entries,
      });
    }
    return result;
  })();

  $: filteredResourceSections = (function (): ResourceDropdownSection[] {
    const query = treeSearchQuery.toLowerCase().trim();
    const hasSearch = query.length > 0;
    const hasStatus = statusFilter.length > 0;
    if (!hasSearch && !hasStatus) return allResourceSections;
    return allResourceSections
      .map((section) => ({
        ...section,
        entries: section.entries.filter((e) => {
          if (hasSearch && !e.name.toLowerCase().includes(query)) return false;
          if (hasStatus && !statusFilter.includes(e.status)) return false;
          return true;
        }),
      }))
      .filter((section) => section.entries.length > 0);
  })();

  function handleResourceSelect(entry: ResourceDropdownEntry) {
    const groupId = `${entry.kind}:${entry.name}`;
    if (isSidebarControlled) {
      onSelectedGroupChange?.(groupId);
    } else {
      internalSelectedGroupId = groupId;
    }
  }

  let internalSelectedGroupId: string | null = null;
  $: isSidebarControlled =
    selectedGroupId !== null || onSelectedGroupChange !== null;

  // Resolve selectedGroupId (which may be a short name like "orders") to a
  // fully qualified group ID (like "rill.runtime.v1.MetricsView:orders")
  function resolveGroupId(
    id: string | null,
    groups: ResourceGraphGrouping[],
  ): string | null {
    if (!id) return null;
    // Exact match
    if (groups.some((g) => g.id === id)) return id;
    // Match by name suffix (group.id = "rill.runtime.v1.Kind:name")
    const match = groups.find((g) => g.id.endsWith(`:${id}`));
    if (match) return match.id;
    // Try normalizing shorthand format (e.g., "model:name" → "rill.runtime.v1.Model:name")
    if (id.includes(":")) {
      const normalized = normalizeSeed(id);
      if (
        typeof normalized !== "string" &&
        normalized.kind &&
        normalized.name
      ) {
        const fqId = `${normalized.kind}:${normalized.name}`;
        const fqMatch = groups.find((g) => g.id === fqId);
        if (fqMatch) return fqMatch.id;
      }
    }
    // Match by label
    const labelMatch = groups.find((g) => g.label === id);
    return labelMatch?.id ?? null;
  }

  // Resolve controlled prop separately to avoid cyclical dependency
  $: resolvedControlledId = resolveGroupId(
    selectedGroupId,
    filteredResourceGroups,
  );

  // Auto-select first group when none selected (controlled path)
  $: if (
    layout === "sidebar" &&
    isSidebarControlled &&
    !resolvedControlledId &&
    filteredResourceGroups.length > 0
  ) {
    onSelectedGroupChange?.(filteredResourceGroups[0].id);
  }

  // Auto-select first group when none selected (uncontrolled path)
  $: if (
    layout === "sidebar" &&
    !isSidebarControlled &&
    !internalSelectedGroupId &&
    filteredResourceGroups.length > 0
  ) {
    internalSelectedGroupId = filteredResourceGroups[0].id;
  }

  // Fallback when selected group is removed by filters (controlled path)
  $: if (
    layout === "sidebar" &&
    isSidebarControlled &&
    resolvedControlledId &&
    !filteredResourceGroups.some((g) => g.id === resolvedControlledId)
  ) {
    onSelectedGroupChange?.(filteredResourceGroups[0]?.id ?? null);
  }

  // Fallback when selected group is removed by filters (uncontrolled path)
  $: if (
    layout === "sidebar" &&
    !isSidebarControlled &&
    internalSelectedGroupId &&
    !filteredResourceGroups.some((g) => g.id === internalSelectedGroupId)
  ) {
    internalSelectedGroupId = filteredResourceGroups[0]?.id ?? null;
  }

  // Effective selected ID for rendering — purely derived, no writes
  $: effectiveSelectedGroupId = isSidebarControlled
    ? resolvedControlledId
    : internalSelectedGroupId;

  $: selectedGroup =
    layout === "sidebar"
      ? (filteredResourceGroups.find(
          (g) => g.id === effectiveSelectedGroupId,
        ) ?? null)
      : null;

  // Display label for the breadcrumb trigger
  $: selectedGroupIsConnector =
    effectiveSelectedGroupId?.includes("Connector") ?? false;
  $: breadcrumbLabel = selectedGroupIsConnector
    ? `Full DAG · ${selectedGroup?.label ?? "OLAP"}`
    : (selectedGroup?.label ?? "Select resource");

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
    connectorCount,
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
      const k = coerceResourceKind(r);
      if (!k) continue;
      // Allow connectors even if hidden; GraphContainer pre-filters to OLAP only
      if (r?.meta?.hidden && k !== ResourceKind.Connector) continue;
      if (k === ResourceKind.Connector) connectors++;
      else if (k === ResourceKind.Source) sources++;
      else if (k === ResourceKind.Model) models++;
      else if (k === ResourceKind.MetricsView) metrics++;
      else if (k === ResourceKind.Explore || k === ResourceKind.Canvas)
        dashboards++;
    }
    return {
      connectorCount: connectors,
      sourcesCount: sources,
      modelsCount: models,
      metricsCount: metrics,
      dashboardsCount: dashboards,
    };
  })());

  // Memoization wrapper for summary data to avoid Svelte reactivity issues with Set/object equality.
  // Without this, the kind selector would re-render on every resource array change
  // even if counts haven't actually changed. The summaryEquals function does shallow comparison
  // of counts while checking resources array reference equality.
  let summaryMemo: SummaryMemo = {
    connector: 0,
    sources: 0,
    models: 0,
    metrics: 0,
    dashboards: 0,
    resources: normalizedResources,
    activeToken: null,
  };
  $: {
    const nextSummary: SummaryMemo = {
      connector: connectorCount,
      sources: sourcesCount,
      models: modelsCount,
      metrics: metricsCount,
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
  {#if layout === "sidebar"}
    <!-- Sidebar layout: toolbar always visible, content varies -->
    <div class="graph-toolbar-bar" class:nav-collapsed={!$navigationOpen}>
      <div class="breadcrumb">
        <DropdownMenu.Root>
          <DropdownMenu.Trigger asChild let:builder>
            <button
              class="text-fg-muted px-[5px] py-1 max-w-fit"
              use:builder.action
              {...builder}
            >
              <span class="gap-x-1.5 items-center font-medium flex">
                <span class="truncate">{breadcrumbLabel}</span>
                <CaretDownIcon size="10px" />
              </span>
            </button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="start" class="w-64">
            <div class="tree-search-wrapper">
              <input
                class="tree-search-input"
                type="text"
                placeholder="Filter resources..."
                bind:value={treeSearchQuery}
                on:keydown|stopPropagation
              />
            </div>
            <div class="tree-dropdown-list">
              {#each filteredResourceSections as section, sIdx}
                {#if sIdx > 0}
                  <DropdownMenu.Separator />
                {/if}
                <div class="section-header">
                  <ResourceTypeBadge kind={section.kind} />
                  <span class="text-[10px] text-fg-muted"
                    >{section.entries.length}</span
                  >
                </div>
                {#each section.entries as entry}
                  {@const entryId = `${entry.kind}:${entry.name}`}
                  {@const isConnectorEntry =
                    entry.kind === ResourceKind.Connector}
                  <DropdownMenu.Item
                    class="flex items-center gap-x-2 cursor-pointer {effectiveSelectedGroupId ===
                    entryId
                      ? 'font-semibold'
                      : ''}"
                    on:click={() => handleResourceSelect(entry)}
                  >
                    <svelte:component
                      this={resourceIconMapping[entry.kind]}
                      size="12px"
                    />
                    <span class="flex-1 truncate text-xs">
                      {isConnectorEntry
                        ? `Full DAG · ${entry.name}`
                        : entry.name}
                    </span>
                    <span class="status-dot {entry.status}"></span>
                  </DropdownMenu.Item>
                {/each}
              {/each}
              {#if filteredResourceSections.length === 0}
                <div class="px-3 py-2 text-xs text-fg-muted">
                  No resources match.
                </div>
              {/if}
            </div>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>

      <div class="toolbar-right">
        {#if hasActiveFilters}
          <button class="clear-link" on:click={handleClearFilters}
            >Clear Filter</button
          >
        {/if}
        {#if statusFilterOptions.length > 0}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger asChild let:builder>
              <Button builders={[builder]} type="tertiary">
                {#if statusFilter.length === 0}
                  All statuses
                {:else}
                  {statusFilter.length} status{statusFilter.length > 1
                    ? "es"
                    : ""}
                {/if}
              </Button>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end" class="w-40">
              {#each statusFilterOptions as opt}
                <DropdownMenu.CheckboxItem
                  checked={statusFilter.includes(opt.value)}
                  onCheckedChange={() => onStatusToggle?.(opt.value)}
                >
                  {opt.label}
                </DropdownMenu.CheckboxItem>
              {/each}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}
        {#if onRefreshAll}
          <Button type="secondary" onClick={onRefreshAll}>
            <RefreshCw size="14" />
            <span>Refresh all sources and models</span>
          </Button>
        {/if}
      </div>
    </div>
    <div class="sidebar-main">
      {#if error}
        <div class="state error">
          <p>{error}</p>
        </div>
      {:else if isLoading || seedTransitionLoading}
        <div class="state">
          <div class="loading-state">
            <DelayedSpinner isLoading={true} size="1.5rem" />
            <p>
              {isLoading ? "Loading project graph..." : "Updating graphs..."}
            </p>
          </div>
        </div>
      {:else if selectedGroup}
        <GraphCanvas
          flowId={selectedGroup.id}
          resources={selectedGroup.resources}
          title={null}
          titleLabel={null}
          titleErrorCount={null}
          anchorError={false}
          rootNodeIds={groupRootNodeIds(selectedGroup)}
          {showControls}
          {showNodeActions}
          showLock={false}
          fillParent={true}
          enableExpand={false}
          {fitViewPadding}
          {fitViewMinZoom}
          {fitViewMaxZoom}
        />
      {:else}
        <div class="state">
          <p>No DAGs match filters.</p>
        </div>
      {/if}
    </div>
  {:else if error}
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
    <div class="graph-toolbar">
      <div></div>
      {#if showSummary}
        <slot
          name="summary"
          connector={connectorCount}
          sources={sourcesCount}
          metrics={metricsCount}
          models={modelsCount}
          dashboards={dashboardsCount}
        >
          <ResourceNodeSelector
            resources={normalizedResources}
            activeResourceId={activeResourceIdForSelector}
          />
        </slot>
      {/if}
    </div>

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
              {showNodeActions}
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
                {showNodeActions}
                showLock={true}
                fillParent={true}
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
  }

  .graph-toolbar-bar {
    @apply flex items-center justify-between px-4 h-11 flex-none gap-x-2;
    transition: padding-left 300ms ease-in-out;
  }

  .graph-toolbar-bar.nav-collapsed {
    padding-left: 44px;
  }

  .breadcrumb {
    @apply flex items-center gap-x-1.5 min-w-0;
  }

  .breadcrumb :global(button),
  .breadcrumb :global(a) {
    @apply bg-transparent border-none cursor-pointer transition-colors;
  }

  .breadcrumb :global(button:hover),
  .breadcrumb :global(a:hover) {
    @apply text-fg-primary;
  }

  .breadcrumb :global(button[data-state="open"]) {
    @apply bg-gray-100 rounded-[2px] text-fg-primary;
  }

  .toolbar-right {
    @apply flex items-center gap-x-2;
  }

  .clear-link {
    @apply text-xs text-primary-500 cursor-pointer;
  }

  .clear-link:hover {
    @apply text-primary-600;
  }

  .tree-search-wrapper {
    @apply px-2 pb-1.5 pt-0.5 border-b;
  }

  .tree-search-input {
    @apply w-full text-xs px-2 py-1.5 rounded border bg-transparent text-fg-primary;
    @apply outline-none;
  }

  .tree-search-input::placeholder {
    @apply text-fg-muted;
  }

  .tree-search-input:focus {
    @apply border-primary-300 ring-1 ring-primary-300;
  }

  .tree-dropdown-list {
    @apply max-h-72 overflow-y-auto;
  }

  .section-header {
    @apply flex items-center justify-between px-2 py-1.5;
  }

  .sidebar-main {
    @apply flex-1 min-w-0 h-full;
  }

  .status-dot {
    @apply flex-shrink-0 rounded-full;
    width: 6px;
    height: 6px;
  }

  .status-dot.ok {
    @apply bg-green-500;
  }

  .status-dot.pending {
    @apply bg-yellow-500;
  }

  .status-dot.errored {
    @apply bg-red-500;
  }

  .resource-graph-grid {
    @apply grid gap-4 flex-1 min-h-0;
    grid-template-columns: repeat(1, minmax(0, 1fr));
    grid-auto-rows: 1fr;
  }

  @media (min-width: 1024px) {
    .resource-graph-grid {
      grid-template-columns: repeat(var(--grid-columns, 3), minmax(0, 1fr));
    }
  }

  .grid-item {
    @apply relative h-full;
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

  .graph-toolbar {
    @apply flex items-end justify-between;
  }
</style>
