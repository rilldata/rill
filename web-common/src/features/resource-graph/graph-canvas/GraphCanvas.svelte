<script lang="ts">
  import {
    Background,
    Controls,
    SvelteFlow,
    type Edge,
    type Node,
    type NodeTypes,
  } from "@xyflow/svelte";
  import "@xyflow/svelte/dist/base.css";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { writable } from "svelte/store";
  import { onMount, onDestroy } from "svelte";
  import { buildResourceGraph } from "./graph-builder";
  import {
    traverseUpstream,
    traverseDownstream,
  } from "../shared/traversal/graph-traversal";
  import ResourceNode from "./ResourceNode.svelte";
  import type { ResourceNodeData } from "../shared/types";
  import { UI_CONFIG, EDGE_CONFIG, FIT_VIEW_CONFIG } from "../shared/config";

  export let resources: V1Resource[] = [];
  export let title: string | null = null;
  // Fine-grained title rendering: base label + error count with conditional coloring
  export let titleLabel: string | null = null;
  export let titleErrorCount: number | null = null;
  export let anchorError: boolean = false;
  // Emphasize particular nodes (e.g., the root/seed node for this graph)
  export let rootNodeIds: string[] | undefined = undefined;
  // Unique flow id to isolate multiple SvelteFlow instances
  export let flowId: string | undefined = undefined;
  // Props and events for expansion control
  export let showControls = false;
  // Controls bar: toggle visibility of the lock/interactive button
  export let showLock = true;
  export let enableExpand = true;
  export let fillParent = false;
  // Fit view configuration - allows customization of how the graph is centered and zoomed
  export let fitViewPadding: number = FIT_VIEW_CONFIG.PADDING;
  export let fitViewMinZoom: number = FIT_VIEW_CONFIG.MIN_ZOOM;
  export let fitViewMaxZoom: number = FIT_VIEW_CONFIG.MAX_ZOOM;
  export let onExpand: () => void = () => {};
  export let showNodeActions = true;

  let hasNodes = false;
  const nodesStore = writable<Node<ResourceNodeData>[]>([]);
  const edgesStore = writable<Edge[]>([]);
  const edgesViewStore = writable<Edge[]>([]);
  let flowKey = "";
  let containerKey = "";
  let containerEl: HTMLDivElement | null = null;
  let ro: ResizeObserver | null = null;

  onMount(() => {
    if (typeof ResizeObserver !== "undefined") {
      ro = new ResizeObserver((entries) => {
        const entry = entries[0];
        if (!entry) return;
        const { width, height } = entry.contentRect;
        const next = `${Math.round(width)}x${Math.round(height)}`;
        if (next !== containerKey) containerKey = next;
      });
      if (containerEl) ro.observe(containerEl);
    }
  });

  onDestroy(() => {
    try {
      ro?.disconnect();
    } catch (error) {
      console.debug(
        "[ResourceGraphCanvas] Failed to disconnect observer",
        error,
      );
    }
    ro = null;
  });

  // Tie Svelte Flow theme to the app theme
  import { themeControl } from "../../themes/theme-control";
  // Derive Svelte Flow color mode from global theme
  $: flowColorMode = ($themeControl === "dark" ? "dark" : "light") as
    | "dark"
    | "light";

  // Use inline height to avoid Tailwind class generation issues with dynamic arbitrary values
  $: containerInlineHeight = fillParent
    ? "100%"
    : `${UI_CONFIG.CARD_HEIGHT_PX}px`;

  const nodeTypes: NodeTypes = {
    "resource-node": ResourceNode as NodeTypes[string],
  };

  const edgeOptions = {
    type: "smoothstep",
    style: EDGE_CONFIG.DEFAULT_STYLE,
    // Small offset so edges clear nodes slightly
    pathOptions: { offset: 3, borderRadius: 4 },
  } as const;

  /**
   * Calculate dynamic edge offset based on node positions to create smoother, straighter routes.
   * Uses smaller offsets for nearly-vertical edges and larger offsets for edges spanning more distance.
   */
  function calculateEdgeOffset(
    sourceNode: Node<ResourceNodeData> | undefined,
    targetNode: Node<ResourceNodeData> | undefined,
  ): number {
    if (!sourceNode || !targetNode) return EDGE_CONFIG.DEFAULT_OFFSET;

    // Calculate center x and handle y positions
    const sx = (sourceNode.position?.x ?? 0) + (sourceNode.width ?? 0) / 2;
    const sy = (sourceNode.position?.y ?? 0) + (sourceNode.height ?? 0); // bottom handle
    const tx = (targetNode.position?.x ?? 0) + (targetNode.width ?? 0) / 2;
    const ty = targetNode.position?.y ?? 0; // top handle

    const dx = Math.abs(tx - sx);
    const dy = Math.abs(ty - sy);

    // For nearly-vertical edges, use minimal offset; otherwise scale with distance
    if (dx < EDGE_CONFIG.VERTICAL_THRESHOLD_PX) return EDGE_CONFIG.MIN_OFFSET;
    return Math.max(
      EDGE_CONFIG.MIN_OFFSET,
      Math.min(
        EDGE_CONFIG.MAX_OFFSET,
        Math.round(dy / EDGE_CONFIG.OFFSET_SCALING_FACTOR),
      ),
    );
  }

  /**
   * Apply edge styling and positioning based on highlighted paths.
   * Highlighted edges use emphasized styling; others are dimmed when selection exists.
   */
  function styleEdges(
    edges: Edge[],
    nodes: Node<ResourceNodeData>[],
    highlightedEdgeIds: Set<string>,
  ): Edge[] {
    const nodeMap = new Map(nodes.map((n) => [n.id, n]));

    return edges.map((e) => {
      const sourceNode = nodeMap.get(e.source);
      const targetNode = nodeMap.get(e.target);
      const offset = calculateEdgeOffset(sourceNode, targetNode);
      const isHighlighted = highlightedEdgeIds.has(e.id);

      return {
        ...e,
        style: isHighlighted
          ? EDGE_CONFIG.HIGHLIGHT_STYLE
          : EDGE_CONFIG.DIM_STYLE,
        pathOptions: { offset, borderRadius: EDGE_CONFIG.BORDER_RADIUS },
      };
    });
  }

  /**
   * Apply default edge positioning when no nodes are selected.
   */
  function styleEdgesDefault(
    edges: Edge[],
    nodes: Node<ResourceNodeData>[],
  ): Edge[] {
    const nodeMap = new Map(nodes.map((n) => [n.id, n]));

    return edges.map((e) => {
      const sourceNode = nodeMap.get(e.source);
      const targetNode = nodeMap.get(e.target);
      const offset = calculateEdgeOffset(sourceNode, targetNode);

      return {
        ...e,
        pathOptions: { offset, borderRadius: EDGE_CONFIG.BORDER_RADIUS },
      };
    });
  }

  // Handle pane click (background) to deselect all nodes
  function handlePaneClick() {
    nodesStore.update((nds) =>
      nds.map((n) => ({
        ...n,
        selected: false,
      })),
    );
  }

  // Reactively compute highlighted edges tracing strictly upstream and downstream
  // from the selected node(s). We explore both directions only at the start node(s):
  // upstream explores only incoming edges (sources) and continues going up;
  // downstream explores only outgoing edges (dependents) and continues going down.
  $: (function updateHighlightedEdges() {
    const nodes = $nodesStore as Node<ResourceNodeData>[];
    const edges = $edgesStore as Edge[];
    const selectedNodes = nodes.filter((n) => n.selected);
    const selectedIds = new Set(selectedNodes.map((n) => n.id));

    // No selection: apply default styling and clear highlights
    if (!selectedIds.size) {
      edgesViewStore.set(styleEdgesDefault(edges, nodes));
      nodesStore.update((nds) =>
        nds.map((n) => ({
          ...n,
          data: { ...n.data, routeHighlighted: false },
        })),
      );
      return;
    }

    // Traverse paths from selected nodes
    const upstream = traverseUpstream(selectedIds, edges);
    const downstream = traverseDownstream(selectedIds, edges);

    const highlightedEdgeIds = new Set<string>([
      ...upstream.edgeIds,
      ...downstream.edgeIds,
    ]);
    const highlightedNodeIds = new Set<string>([
      ...upstream.visited,
      ...downstream.visited,
    ]);

    // Apply highlighted styling to edges
    edgesViewStore.set(styleEdges(edges, nodes, highlightedEdgeIds));

    // Mark nodes along the traced paths as highlighted
    nodesStore.update((nds) =>
      nds.map((n) => ({
        ...n,
        data: { ...n.data, routeHighlighted: highlightedNodeIds.has(n.id) },
      })),
    );
  })();

  $: {
    const rootSet = new Set(rootNodeIds ?? []);
    const graph = buildResourceGraph(resources ?? [], {
      positionNs: flowId,
      ignoreCache: true,
    });
    const nodeIds = new Set(graph.nodes.map((n) => n.id));
    const filteredEdges = graph.edges.filter(
      (e) => nodeIds.has(e.source) && nodeIds.has(e.target),
    );
    const nodesWithRoots = (graph.nodes as Node<ResourceNodeData>[]).map(
      (node) => ({
        ...node,
        data: { ...node.data, isRoot: rootSet.has(node.id), showNodeActions },
      }),
    );
    nodesStore.set(nodesWithRoots);
    edgesStore.set(filteredEdges);
    hasNodes = nodesWithRoots.length > 0;
    // Build a signature of the current graph to force SvelteFlow to remount and refit when graph changes
    try {
      const nodeSig = nodesWithRoots
        .map((n) => n.id)
        .sort()
        .join(",");
      const edgeSig = filteredEdges
        .map((e) => e.id || `${e.source}->${e.target}`)
        .sort()
        .join(",");
      flowKey = `${flowId ?? "flow"}|${fillParent ? "E" : "N"}|n:${nodeSig}|e:${edgeSig}|c:${containerKey}`;
    } catch {
      flowKey = `${flowId ?? "flow"}|${fillParent ? "E" : "N"}|${Date.now()}`;
    }
    // Debug logging (only in development)
    if (import.meta.env.DEV) {
      console.log("ResourceGraph graph", {
        title,
        nodes: nodesWithRoots.map((n) => n.id),
        edges: filteredEdges.map((e) => ({
          id: e.id,
          source: e.source,
          target: e.target,
        })),
      });
    }
  }
</script>

<section class="graph-instance">
  {#if hasNodes}
    <div
      class="graph-container"
      class:h-full={fillParent}
      class:no-border={fillParent}
      bind:this={containerEl}
      style:height={containerInlineHeight}
    >
      {#if titleLabel != null}
        <div class="graph-watermark">
          <span class:text-red-600={anchorError}>{titleLabel}</span>
          {#if titleErrorCount && titleErrorCount > 0}
            <span class="text-red-600">
              • {titleErrorCount} error{titleErrorCount === 1 ? "" : "s"}
            </span>
          {/if}
        </div>
      {:else if title}
        <div class="graph-watermark">{title}</div>
      {/if}
      {#if enableExpand}
        <button
          class="expand-btn"
          aria-label="Expand graph"
          title="Expand"
          on:click={() => onExpand()}
        >
          ⤢
        </button>
      {/if}

      {#key flowKey}
        <SvelteFlow
          id={flowId}
          nodes={nodesStore}
          edges={edgesViewStore}
          {nodeTypes}
          colorMode={flowColorMode}
          proOptions={{ hideAttribution: true }}
          fitView
          fitViewOptions={{
            padding: fitViewPadding,
            minZoom: fitViewMinZoom,
            maxZoom: fitViewMaxZoom,
            duration: 0,
          }}
          preventScrolling={false}
          zoomOnScroll={false}
          panOnScroll={false}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable
          selectionOnDrag
          onlyRenderVisibleElements={false}
          defaultEdgeOptions={edgeOptions}
          on:paneclick={handlePaneClick}
        >
          <Background gap={24} />
          {#if showControls}
            <Controls position="top-right" {showLock} />
          {/if}
        </SvelteFlow>
      {/key}
    </div>
  {:else}
    <div class="state">
      <p>No resources found.</p>
    </div>
  {/if}
</section>

<style lang="postcss">
  .graph-instance {
    @apply flex h-full flex-col;
  }

  .graph-container {
    @apply relative w-full overflow-hidden rounded-lg border;
  }

  .graph-container.no-border {
    @apply border-0 rounded-none;
  }

  .state {
    @apply flex h-[160px] w-full items-center justify-center rounded-lg border border-dashed text-sm text-fg-muted;
  }

  .expand-btn {
    @apply absolute right-2 top-2 z-20 h-7 w-7 rounded border bg-surface-subtle text-sm text-fg-secondary;
    line-height: 1.25rem;
  }

  .expand-btn:hover {
    @apply bg-surface-muted text-fg-primary;
  }

  .graph-watermark {
    @apply absolute bottom-3 right-3 z-10 pointer-events-none;
    @apply text-xs font-semibold leading-tight text-fg-secondary opacity-70;
  }

  /* Override xyflow pane background to match app theme - scoped to this component */
  .graph-container :global(.svelte-flow .svelte-flow__pane) {
    background-color: var(--surface-background, #ffffff);
  }
</style>
