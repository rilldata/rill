<script lang="ts">
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import GraphCanvas from "../graph-canvas/GraphCanvas.svelte";
  import GraphOverlay from "./GraphOverlay.svelte";
  import {
    partitionResourcesByMetrics,
    partitionResourcesBySeeds,
    buildResourceGraph,
    type ResourceGraphGrouping,
  } from "../graph-canvas/graph-builder";
  import type { Edge, Node } from "@xyflow/svelte";
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
  import {
    UI_CONFIG,
    FIT_VIEW_CONFIG,
    PERFORMANCE_CONFIG,
    RESOURCE_SECTION_ORDER,
    RESOURCE_SECTION_LABELS,
  } from "../shared/config";
  import type {
    ResourceNodeData,
    ResourceStatusFilter,
    ResourceStatusFilterValue,
  } from "../shared/types";
  import { getResourceStatus } from "../shared/resource-status";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import {
    FilterIcon,
    SearchIcon,
    XIcon,
  } from "lucide-svelte";

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
  export let showIsolatedResources = false;
  export let onShowIsolatedChange: ((value: boolean) => void) | null = null;
  $: onShowIsolatedChange?.(showIsolatedResources);

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
  export let onSelectAll: (() => void) | null = null;
  export let hasUrlFilters = false;
  export let flushToolbar = false;
  export let showTitle = true;
  export let showToolbar = true;

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

  // Track container width for dynamic multi-tree row wrapping
  let sidebarMainWidth = 0;

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

  // Detect sprawl mode: seeds contain only a single kind token ("dashboards" or "metrics")
  // and no specific resource is selected via controlled prop
  $: sprawlSeed = (() => {
    if (selectedGroupId) return null;
    const rawSeeds = seeds ?? [];
    if (rawSeeds.length !== 1) return null;
    const s = rawSeeds[0]?.toLowerCase();
    if (s === "dashboards" || s === "metrics") return s;
    return null;
  })();
  $: isSprawlMode = !!sprawlSeed;

  // Trees for sprawl mode: use seed-based (directed) partitioning so each
  // metrics view / dashboard gets its own independent tree. This allows
  // status filtering to hide/show individual trees correctly.
  $: sprawlTreeGroups = (() => {
    if (!isSprawlMode) return [];
    const seedKind =
      sprawlSeed === "dashboards"
        ? ResourceKind.Explore
        : ResourceKind.MetricsView;
    // Build one seed per anchor resource of the relevant kind
    const anchorSeeds: { kind: string; name: string }[] = [];
    for (const r of normalizedResources) {
      const kind = coerceResourceKind(r);
      if (sprawlSeed === "dashboards") {
        if (kind !== ResourceKind.Explore && kind !== ResourceKind.Canvas)
          continue;
      } else {
        if (kind !== seedKind) continue;
      }
      if (r.meta?.hidden) continue;
      const name = r.meta?.name?.name;
      const rKind = r.meta?.name?.kind;
      if (!name || !rKind) continue;
      anchorSeeds.push({ kind: rKind, name });
    }
    if (!anchorSeeds.length) return [];
    return partitionResourcesBySeeds(normalizedResources, anchorSeeds);
  })();

  // If seeds were provided (e.g., kind filter like "models"), use seed-based partitioning.
  // When the kind has no resources, normalizedSeeds will be empty — return [] instead of
  // falling back to partitionResourcesByMetrics, which would show unrelated metric views.
  $: hasExplicitSeeds = seeds && seeds.length > 0;
  $: resourceGroups = isSprawlMode
    ? sprawlTreeGroups
    : normalizedSeeds && normalizedSeeds.length
      ? partitionResourcesBySeeds(
          normalizedResources,
          normalizedSeeds,
          filterKind,
        )
      : hasExplicitSeeds
        ? []
        : partitionResourcesByMetrics(normalizedResources);

  // Compute IDs of all resources reachable upstream from dashboard anchors.
  // Only follows refs (dependencies), not reverse refs, so orphaned resources
  // sharing a connector with a dashboard tree are correctly excluded.
  $: dashboardConnectedIds = (() => {
    const idToResource = new Map<string, V1Resource>();
    for (const r of normalizedResources) {
      const id = `${r.meta?.name?.kind}:${r.meta?.name?.name}`;
      idToResource.set(id, r);
    }

    const connected = new Set<string>();
    const queue: string[] = [];
    for (const r of normalizedResources) {
      const kind = coerceResourceKind(r);
      if (
        kind === ResourceKind.Explore ||
        kind === ResourceKind.Canvas ||
        kind === ResourceKind.MetricsView
      ) {
        queue.push(`${r.meta?.name?.kind}:${r.meta?.name?.name}`);
      }
    }

    while (queue.length) {
      const id = queue.pop()!;
      if (connected.has(id)) continue;
      connected.add(id);
      const r = idToResource.get(id);
      for (const ref of r?.meta?.refs ?? []) {
        const refId = `${ref.kind}:${ref.name}`;
        if (!connected.has(refId)) queue.push(refId);
      }
    }

    return connected;
  })();

  // Filter groups by search query and status
  $: filteredResourceGroups = (() => {
    let groups = resourceGroups;

    // Filter by external search query (matches resource names)
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase().trim();
      groups = groups.filter((group) =>
        group.resources.some((r) =>
          r.meta?.name?.name?.toLowerCase().includes(query),
        ),
      );
    }

    // Filter by toolbar search (matches resource names within trees)
    if (treeSearchQuery.trim()) {
      const query = treeSearchQuery.toLowerCase().trim();
      groups = groups.filter((group) =>
        group.resources.some((r) =>
          r.meta?.name?.name?.toLowerCase().includes(query),
        ),
      );
    }

    // Filter trees by status: show full tree if any resource in it matches
    if (statusFilter.length > 0) {
      groups = groups.filter((group) =>
        group.resources.some((r) =>
          statusFilter.includes(getResourceStatus(r)),
        ),
      );
    }

    // Hide isolated resources (not reachable from any Explore/Canvas/MetricsView)
    if (!showIsolatedResources) {
      groups = groups
        .map((group) => ({
          ...group,
          resources: group.resources.filter((r) => {
            const id = `${r.meta?.name?.kind}:${r.meta?.name?.name}`;
            return dashboardConnectedIds.has(id);
          }),
        }))
        .filter((group) => group.resources.length > 0);
    }

    return groups;
  })();

  // Build combined layout for sprawl mode: merge all filtered groups'
  // resources into one graph so shared nodes (metrics views, models) appear
  // once and naturally connect the trees together.
  $: sprawlLayout = (() => {
    if (!isSprawlMode || !filteredResourceGroups.length) return null;
    const seen = new Set<string>();
    const merged: V1Resource[] = [];
    for (const group of filteredResourceGroups) {
      for (const r of group.resources) {
        const id = `${r.meta?.name?.kind}:${r.meta?.name?.name}`;
        if (seen.has(id)) continue;
        seen.add(id);
        merged.push(r);
      }
    }
    return buildResourceGraph(merged, {
      positionNs: "sprawl",
      ignoreCache: true,
    });
  })() as { nodes: Node<ResourceNodeData>[]; edges: Edge[] } | null;

  $: visibleResourceGroups =
    typeof maxGroups === "number" && maxGroups >= 0
      ? filteredResourceGroups.slice(0, maxGroups)
      : filteredResourceGroups;
  $: hasGraphs = visibleResourceGroups.length > 0;

  // Whether any filters are active (URL params, status, or tree search)
  $: activeFilterCount =
    statusFilter.length +
    (!showIsolatedResources ? 0 : 0); // extend as needed
  $: hasActiveFilters =
    hasUrlFilters ||
    statusFilter.length > 0 ||
    treeSearchQuery.trim().length > 0;

  function handleClearFilters() {
    treeSearchQuery = "";
    onClearFilters?.();
  }

  // --- Sidebar selection state ---
  function handleSearchComboBlur(e: FocusEvent) {
    const related = e.relatedTarget as globalThis.Node | null;
    if (!(e.currentTarget as globalThis.Node)?.contains(related)) {
      resourceDropdownOpen = false;
    }
  }

  let resourceDropdownOpen = false;
  let statusDropdownOpen = false;
  let filterDropdownOpen = false;
  let searchExpanded = false;
  let treeSearchQuery = "";

  // Sync search bar with URL-selected resource on initial navigation
  let lastSyncedGroupId: string | null = null;
  $: if (selectedGroupId && selectedGroupId !== lastSyncedGroupId) {
    // Extract short name from group ID (e.g. "rill.runtime.v1.Model:orders" -> "orders")
    const name = selectedGroupId.includes(":")
      ? (selectedGroupId.split(":").pop() ?? selectedGroupId)
      : selectedGroupId;
    treeSearchQuery = name;
    searchExpanded = true;
    lastSyncedGroupId = selectedGroupId;
  }

  // All resources organized by kind for the tree dropdown
  type ResourceDropdownEntry = {
    name: string;
    kind: ResourceKind;
    status: "ok" | "pending" | "warning" | "errored";
  };
  type ResourceDropdownSection = {
    kind: ResourceKind;
    label: string;
    entries: ResourceDropdownEntry[];
  };

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
    for (const kind of RESOURCE_SECTION_ORDER) {
      const entries = grouped.get(kind);
      if (!entries?.length) continue;
      entries.sort((a, b) => a.name.localeCompare(b.name));
      result.push({
        kind,
        label: RESOURCE_SECTION_LABELS[kind] ?? kind,
        entries,
      });
    }
    return result;
  })();

  function handleResourceSelect(entry: ResourceDropdownEntry) {
    const groupId = `${entry.kind}:${entry.name}`;
    internalSelectedGroupId = groupId;
    onSelectedGroupChange?.(groupId);
  }

  function handleSelectAll() {
    if (onSelectAll) {
      onSelectAll();
    } else {
      // Fallback: select the first connector entry (shows full DAG)
      const connectorSection = allResourceSections.find(
        (s) => s.kind === ResourceKind.Connector,
      );
      if (connectorSection?.entries[0]) {
        handleResourceSelect(connectorSection.entries[0]);
      }
    }
  }

  let internalSelectedGroupId: string | null = null;

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

  // Auto-select first group when none selected.
  // Uses internal state only — never calls onSelectedGroupChange, so the URL
  // stays clean until the user explicitly clicks a resource.
  $: if (
    layout === "sidebar" &&
    !resolvedControlledId &&
    !internalSelectedGroupId &&
    filteredResourceGroups.length > 0
  ) {
    internalSelectedGroupId = filteredResourceGroups[0].id;
  }

  // Fallback when selected group is removed by filters
  $: if (
    layout === "sidebar" &&
    (resolvedControlledId || internalSelectedGroupId) &&
    !filteredResourceGroups.some(
      (g) => g.id === resolvedControlledId || g.id === internalSelectedGroupId,
    )
  ) {
    internalSelectedGroupId = filteredResourceGroups[0]?.id ?? null;
  }

  // Effective selected ID: controlled (URL) takes precedence, falls back to internal
  $: effectiveSelectedGroupId = resolvedControlledId ?? internalSelectedGroupId;

  $: selectedGroup =
    layout === "sidebar"
      ? (filteredResourceGroups.find(
          (g) => g.id === effectiveSelectedGroupId,
        ) ?? null)
      : null;

  $: selectedGroupIsConnector =
    effectiveSelectedGroupId?.includes("Connector") ?? false;

  // Brief loading indicator when URL seeds change (e.g., via Overview node clicks)
  let seedTransitionLoading = false;
  let seedTransitionTimer: ReturnType<typeof setTimeout> | null = null;

  // Lazy rendering: only mount GraphCanvas when grid item is near the viewport.
  // Uses IntersectionObserver with a generous rootMargin so graphs mount before
  // the user scrolls to them, avoiding visible pop-in.
  const LAZY_ROOT_MARGIN = "200px";
  let visibleGroupIds = new Set<string>();
  let prevGroupIdKey = "";

  function lazyObserve(node: HTMLElement, groupId: string) {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !visibleGroupIds.has(groupId)) {
          visibleGroupIds = new Set([...visibleGroupIds, groupId]);
        }
      },
      { rootMargin: LAZY_ROOT_MARGIN },
    );
    observer.observe(node);
    return {
      destroy() {
        observer.disconnect();
      },
    };
  }

  // Reset visible set only when the set of group IDs changes (not on resource-content updates)
  $: {
    const key = visibleResourceGroups.map((g) => g.id).join(",");
    if (key !== prevGroupIdKey) {
      prevGroupIdKey = key;
      visibleGroupIds = new Set<string>();
    }
  }

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

  // Inspect panel state is scoped per GraphCanvas instance via Svelte context.
  // URL changes cause the graph to re-render, which naturally resets the panel.

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
    const signature = JSON.stringify(seeds ?? []);
    if (signature !== lastSeedsSignature) {
      // Show a short loading state to indicate graphs are updating.
      // Skip the transition on the very first render (lastSeedsSignature is "")
      // to avoid a 500ms blank screen on initial page load.
      const isFirstRender = lastSeedsSignature === "";
      if (!isFirstRender) {
        seedTransitionLoading = true;
        if (seedTransitionTimer) clearTimeout(seedTransitionTimer);
        seedTransitionTimer = setTimeout(
          () => (seedTransitionLoading = false),
          PERFORMANCE_CONFIG.SEED_TRANSITION_DELAY_MS,
        );
      }

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
    <section class="flex flex-col gap-y-3 flex-1 min-h-0">
    {#if showToolbar}
      <!-- Row 1: Title + Refresh -->
      {#if showTitle}
        <div
          class="graph-title-bar"
          class:nav-collapsed={!$navigationOpen}
          class:flush-toolbar={flushToolbar}
        >
          <h2 class="text-lg font-medium">Resource Graph (DAG)</h2>
          {#if onRefreshAll}
            <Button
              type="secondary"
              large
              class="shrink-0 whitespace-nowrap"
              onClick={onRefreshAll}
            >
              <span class="hidden lg:inline">Refresh all sources and models</span>
              <span class="lg:hidden">Refresh all</span>
            </Button>
          {/if}
        </div>
      {/if}
      <!-- Row 2: Filter + search -->
      <div
        class="graph-toolbar-bar"
        class:nav-collapsed={!$navigationOpen}
        class:flush-toolbar={flushToolbar}
      >
        <!-- Filter dropdown -->
        <DropdownMenu.Root bind:open={filterDropdownOpen}>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <button
                {...props}
                class="filter-trigger"
              >
                <FilterIcon size="14px" />
                <span>Filter</span>
                {#if statusFilter.length > 0}
                  <span class="filter-badge">{statusFilter.length}</span>
                {/if}
              </button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="start" class="w-52">
            {#if statusFilterOptions.length > 0}
              <DropdownMenu.Group>
                <DropdownMenu.Label class="uppercase text-[10px] tracking-wide"
                  >Status</DropdownMenu.Label
                >
                {#each statusFilterOptions as opt (opt.value)}
                  <DropdownMenu.CheckboxItem
                    closeOnSelect={false}
                    checked={statusFilter.includes(opt.value)}
                    onCheckedChange={() => onStatusToggle?.(opt.value)}
                  >
                    {opt.label}
                  </DropdownMenu.CheckboxItem>
                {/each}
              </DropdownMenu.Group>
              <DropdownMenu.Separator />
            {/if}
            <DropdownMenu.Group>
              <DropdownMenu.Label class="uppercase text-[10px] tracking-wide"
                >Visibility</DropdownMenu.Label
              >
              <DropdownMenu.CheckboxItem
                closeOnSelect={false}
                checked={!showIsolatedResources}
                onCheckedChange={() => {
                  showIsolatedResources = !showIsolatedResources;
                  onShowIsolatedChange?.(showIsolatedResources);
                }}
              >
                Hide isolated
              </DropdownMenu.CheckboxItem>
            </DropdownMenu.Group>
          </DropdownMenu.Content>
        </DropdownMenu.Root>

        <div class="flex-1"></div>

        <!-- Search icon / expandable search -->
        {#if searchExpanded}
          <div class="flex items-center w-56 h-9 shrink-0">
            <Search
              bind:value={treeSearchQuery}
              placeholder="Search resources..."
              large
              autofocus={true}
              showBorderOnFocus={false}
              retainValueOnMount
            />
            <button
              class="h-9 w-9 flex items-center justify-center text-fg-primary shrink-0"
              onclick={() => {
                searchExpanded = false;
                treeSearchQuery = "";
              }}
            >
              <XIcon size="14px" />
            </button>
          </div>
        {:else}
          <button
            class="toolbar-icon-btn"
            onclick={() => (searchExpanded = true)}
          >
            <SearchIcon size="14px" />
          </button>
        {/if}

      </div>

      <!-- Divider -->
      <hr
        class="border-t border-gray-200 my-0"
        class:flush-toolbar={flushToolbar}
      />

      <!-- Filter pills row -->
      {#if hasActiveFilters}
        <div
          class="filter-pills-row"
          class:flush-toolbar={flushToolbar}
        >
          <div class="filter-pills-scroll">
            {#if statusFilter.length > 0}
              <button
                class="filter-pill"
                onclick={() => onClearFilters?.()}
              >
                <span>Status = {statusFilter
                  .map((s) => statusFilterOptions.find((o) => o.value === s)?.label ?? s)
                  .join(", ")}</span>
                <XIcon size="10px" />
              </button>
            {/if}
            {#if showIsolatedResources}
              <button
                class="filter-pill"
                onclick={() => {
                  showIsolatedResources = false;
                  onShowIsolatedChange?.(false);
                }}
              >
                <span>Show isolated</span>
                <XIcon size="10px" />
              </button>
            {/if}
          </div>
          <button
            class="filter-pills-clear"
            onclick={handleClearFilters}
          >
            Clear all
          </button>
        </div>
      {/if}
    {/if}
    <div class="sidebar-main" bind:clientWidth={sidebarMainWidth}>
      {#if error}
        <div class="state error">
          <p>{error}</p>
        </div>
      {:else if isLoading || seedTransitionLoading}
        <div class="state">
          <div class="loading-state">
            <LoadingSpinner size="1.5rem" />
            <p>
              {isLoading ? "Loading project graph..." : "Updating graphs..."}
            </p>
          </div>
        </div>
      {:else if isSprawlMode && sprawlLayout}
        <GraphCanvas
          flowId="sprawl"
          resources={normalizedResources}
          precomputedNodes={sprawlLayout.nodes}
          precomputedEdges={sprawlLayout.edges}
          title={null}
          titleLabel={null}
          titleErrorCount={null}
          anchorError={false}
          {showControls}
          {showNodeActions}
          showLock={false}
          fillParent={true}
          enableExpand={false}
          {fitViewPadding}
          {fitViewMinZoom}
          {fitViewMaxZoom}
        />
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
          <p>No resources match the current filters.</p>
        </div>
      {/if}
    </div>
    </section>
  {:else if error}
    <div class="state error">
      <p>{error}</p>
    </div>
  {:else if isLoading || seedTransitionLoading}
    <div class="state">
      <div class="loading-state">
        <LoadingSpinner size="1.5rem" />
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
        {@const parts = groupTitleParts(group, index)}
        {@const isVisible = isExpanded || visibleGroupIds.has(group.id)}
        <div
          class="grid-item"
          class:expanded={isExpanded}
          class:hidden={isHidden}
          use:lazyObserve={group.id}
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
              titleLabel={showCardTitles ? parts.labelWithCount : null}
              titleErrorCount={showCardTitles ? parts.errorCount : null}
              anchorError={showCardTitles ? parts.anchorError : false}
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
          {:else if isVisible}
            <!-- Collapsed card view (lazy-mounted when near viewport) -->
            <slot name="graph-item" {group} {index}>
              <GraphCanvas
                flowId={group.id}
                resources={group.resources}
                title={null}
                titleLabel={showCardTitles ? parts.labelWithCount : null}
                titleErrorCount={showCardTitles ? parts.errorCount : null}
                anchorError={showCardTitles ? parts.anchorError : false}
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
    @apply flex flex-row items-center min-h-[3rem] flex-none px-2;
    transition: padding-left 300ms ease-in-out;
  }

  .graph-toolbar-bar.nav-collapsed {
    padding-left: 44px;
  }

  .graph-toolbar-bar.flush-toolbar {
    @apply px-0;
  }

  .filter-trigger {
    @apply flex items-center gap-1.5 px-4 py-1.5 rounded-sm bg-primary-50 text-sm text-primary-600;
  }
  :global(.dark) .filter-trigger {
    @apply bg-surface-background text-primary-500;
  }
  .filter-trigger:hover {
    @apply bg-primary-100;
  }
  :global(.dark) .filter-trigger:hover {
    @apply bg-surface-muted;
  }

  .filter-badge {
    @apply text-[10px] font-semibold bg-primary-500 text-white rounded-full w-4 h-4 flex items-center justify-center;
  }

  .filter-pills-row {
    @apply flex items-center min-h-7 relative px-2;
  }

  .filter-pills-row.flush-toolbar {
    @apply px-0;
  }

  .filter-pills-scroll {
    @apply flex items-center gap-1.5 flex-1 min-w-0 overflow-hidden;
  }

  .filter-pills-clear {
    @apply shrink-0 text-xs text-fg-primary hover:underline whitespace-nowrap pl-2 pr-1;
  }

  .filter-pill {
    @apply flex items-center gap-1.5 text-xs font-medium text-fg-primary border border-gray-300 rounded-sm px-2 py-1 whitespace-nowrap shrink-0;
  }
  .filter-pill:hover {
    @apply bg-surface-hover;
  }

  .toolbar-icon-btn {
    @apply h-9 w-9 flex items-center justify-center text-fg-primary;
  }

  .graph-title-bar {
    @apply flex items-center justify-between h-9 flex-none px-2 pt-3 pb-1;
    transition: padding-left 300ms ease-in-out;
  }

  .graph-title-bar.nav-collapsed {
    padding-left: 44px;
  }

  .graph-title-bar.flush-toolbar {
    @apply px-0;
  }

  .search-combo {
    @apply relative flex-1 min-w-0 flex items-center;
  }

  .search-combo-dropdown {
    @apply absolute top-full left-0 right-0 z-50 mt-1 rounded-md border bg-popover shadow-md;
    min-width: 24rem;
  }

  .combo-item {
    @apply flex w-full items-center gap-x-2 cursor-pointer rounded-sm py-1.5 px-2 text-xs text-left;
    &:hover {
      @apply bg-surface-hover;
    }
  }

  .combo-separator {
    @apply my-1 h-px bg-border;
  }

  .tree-dropdown-list {
    @apply max-h-72 overflow-y-auto overflow-x-hidden p-1;
  }

  .section-header {
    @apply flex items-center justify-between px-2 py-1.5;
  }

  .sidebar-main {
    @apply flex-1 min-w-0 h-full;
  }

  .status-dot {
    @apply flex-shrink-0;
    width: 6px;
    height: 6px;
  }

  .status-dot.ok {
    @apply rounded-full bg-green-500;
  }

  .status-dot.pending {
    @apply rounded-full border border-yellow-500;
    background: transparent;
  }

  .status-dot.warning {
    @apply rounded-full bg-yellow-500;
  }

  .status-dot.errored {
    @apply bg-red-500;
    border-radius: 1px;
    transform: rotate(45deg);
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
    min-height: 200px;
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
