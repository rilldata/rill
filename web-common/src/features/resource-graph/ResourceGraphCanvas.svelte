<script lang="ts">
  import {
    Background,
    Controls,
    SvelteFlow,
    type Edge,
    type Node,
  } from "@xyflow/svelte";
  import '@xyflow/svelte/dist/base.css';
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { writable } from "svelte/store";
  import { onMount, onDestroy } from "svelte";
  import { buildResourceGraph } from "./build-resource-graph";
  import { traverseUpstream, traverseDownstream } from "./graph-traversal";
  import ResourceNode from "./ResourceNode.svelte";
  import type { ResourceNodeData } from "./types";

  export let resources: V1Resource[] = [];
  export let title: string | null = null;
  // Fine-grained title rendering: base label + error count with conditional coloring
  export let titleLabel: string | null = null;
  export let titleErrorCount: number | null = null;
  export let anchorError: boolean = false;
  // Preselect specific nodes by id on initial render (e.g., the seeded anchor)
  export let preselectNodeIds: string[] | undefined = undefined;
  // Unique flow id to isolate multiple SvelteFlow instances
  export let flowId: string | undefined = undefined;

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
    try { ro?.disconnect(); } catch {}
    ro = null;
  });

  // Props and events for expansion control
  export let showControls = false;
  // Controls bar: toggle visibility of the lock/interactive button
  export let showLock = true;
  export let enableExpand = true;
  export let fillParent = false;
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher<{ expand: void }>();
  // Tie Svelte Flow theme to the app theme
  import { themeControl } from "../themes/theme-control";
  // Derive Svelte Flow color mode from global theme
  $: flowColorMode = ($themeControl === "dark" ? "dark" : "light") as
    | "dark"
    | "light";

  // Layout constants
  const CARD_HEIGHT_PX = 260; // Sized to fit 3x3 grid comfortably on standard displays
  const EDGE_BORDER_RADIUS = 6; // Rounded corners for edge paths

  // Edge offset calculation constants
  const DEFAULT_EDGE_OFFSET = 8; // Default offset when nodes are moderately spaced
  const MIN_EDGE_OFFSET = 4; // Minimal offset for nearly-vertical edges
  const MAX_EDGE_OFFSET = 18; // Maximum offset for widely-spaced nodes
  const VERTICAL_EDGE_THRESHOLD_PX = 12; // Treat edge as vertical if horizontal distance < this
  const EDGE_OFFSET_SCALING_FACTOR = 10; // Divides vertical distance to compute dynamic offset

  // Shrink card height so 3x3 fits comfortably
  $: containerHeightClass = fillParent ? "h-full" : `h-[${CARD_HEIGHT_PX}px]`;

  const nodeTypes = {
    "resource-node": ResourceNode,
  };

  const edgeOptions = {
    type: "smoothstep",
    style: "stroke:#b1b1b7;stroke-width:1px;opacity:0.85;",
    // Small offset so edges clear nodes slightly
    pathOptions: { offset: 3, borderRadius: 4 },
  } as const;

  const HIGHLIGHT_EDGE_STYLE = "stroke:#3b82f6;stroke-width:2px;opacity:1;";
  const DIM_EDGE_STYLE = "stroke:#b1b1b7;stroke-width:1px;opacity:0.25;";

  /**
   * Calculate dynamic edge offset based on node positions to create smoother, straighter routes.
   * Uses smaller offsets for nearly-vertical edges and larger offsets for edges spanning more distance.
   */
  function calculateEdgeOffset(
    sourceNode: Node<ResourceNodeData> | undefined,
    targetNode: Node<ResourceNodeData> | undefined
  ): number {
    if (!sourceNode || !targetNode) return DEFAULT_EDGE_OFFSET;

    // Calculate center x and handle y positions
    const sx = (sourceNode.position?.x ?? 0) + (sourceNode.width ?? 0) / 2;
    const sy = (sourceNode.position?.y ?? 0) + (sourceNode.height ?? 0); // bottom handle
    const tx = (targetNode.position?.x ?? 0) + (targetNode.width ?? 0) / 2;
    const ty = (targetNode.position?.y ?? 0); // top handle

    const dx = Math.abs(tx - sx);
    const dy = Math.abs(ty - sy);

    // For nearly-vertical edges, use minimal offset; otherwise scale with distance
    if (dx < VERTICAL_EDGE_THRESHOLD_PX) return MIN_EDGE_OFFSET;
    return Math.max(
      MIN_EDGE_OFFSET,
      Math.min(MAX_EDGE_OFFSET, Math.round(dy / EDGE_OFFSET_SCALING_FACTOR))
    );
  }

  /**
   * Apply edge styling and positioning based on highlighted paths.
   * Highlighted edges use emphasized styling; others are dimmed when selection exists.
   */
  function styleEdges(
    edges: Edge[],
    nodes: Node<ResourceNodeData>[],
    highlightedEdgeIds: Set<string>
  ): Edge[] {
    const nodeMap = new Map(nodes.map((n) => [n.id, n]));

    return edges.map((e) => {
      const sourceNode = nodeMap.get(e.source);
      const targetNode = nodeMap.get(e.target);
      const offset = calculateEdgeOffset(sourceNode, targetNode);
      const isHighlighted = highlightedEdgeIds.has(e.id);

      return {
        ...e,
        style: isHighlighted ? HIGHLIGHT_EDGE_STYLE : DIM_EDGE_STYLE,
        pathOptions: { offset, borderRadius: EDGE_BORDER_RADIUS },
      };
    });
  }

  /**
   * Apply default edge positioning when no nodes are selected.
   */
  function styleEdgesDefault(edges: Edge[], nodes: Node<ResourceNodeData>[]): Edge[] {
    const nodeMap = new Map(nodes.map((n) => [n.id, n]));

    return edges.map((e) => {
      const sourceNode = nodeMap.get(e.source);
      const targetNode = nodeMap.get(e.target);
      const offset = calculateEdgeOffset(sourceNode, targetNode);

      return { ...e, pathOptions: { offset, borderRadius: EDGE_BORDER_RADIUS } };
    });
  }

  // Reactively compute highlighted edges tracing strictly upstream and downstream
  // from the selected node(s). We explore both directions only at the start node(s):
  // upstream explores only incoming edges (sources) and continues going up;
  // downstream explores only outgoing edges (dependents) and continues going down.
  $: (function updateHighlightedEdges() {
    const nodes = $nodesStore as Node<ResourceNodeData>[];
    const edges = $edgesStore as Edge[];
    const selectedIds = new Set(nodes.filter((n) => n.selected).map((n) => n.id));

    // No selection: apply default styling and clear highlights
    if (!selectedIds.size) {
      edgesViewStore.set(styleEdgesDefault(edges, nodes));
      nodesStore.update((nds) => nds.map((n) => ({
        ...n,
        data: { ...n.data, routeHighlighted: false },
      })));
      return;
    }

    // Traverse paths from selected nodes
    const upstream = traverseUpstream(selectedIds, edges);
    const downstream = traverseDownstream(selectedIds, edges);

    const highlightedEdgeIds = new Set<string>([...upstream.edgeIds, ...downstream.edgeIds]);
    const highlightedNodeIds = new Set<string>([...upstream.visited, ...downstream.visited]);

    // Apply highlighted styling to edges
    edgesViewStore.set(styleEdges(edges, nodes, highlightedEdgeIds));

    // Mark nodes along the traced paths as highlighted
    nodesStore.update((nds) => nds.map((n) => ({
      ...n,
      data: { ...n.data, routeHighlighted: highlightedNodeIds.has(n.id) },
    })));
  })();

  $: {
    const graph = buildResourceGraph(resources ?? [], { positionNs: flowId, ignoreCache: true });
    const nodeIds = new Set(graph.nodes.map((n) => n.id));
    const filteredEdges = graph.edges.filter(
      (e) => nodeIds.has(e.source) && nodeIds.has(e.target),
    );
    nodesStore.set(graph.nodes as Node<ResourceNodeData>[]);
    edgesStore.set(filteredEdges);
    hasNodes = graph.nodes.length > 0;
    // Build a signature of the current graph to force SvelteFlow to remount and refit when graph changes
    try {
      const nodeSig = graph.nodes.map((n) => n.id).sort().join(",");
      const edgeSig = filteredEdges
        .map((e) => e.id || `${e.source}->${e.target}`)
        .sort()
        .join(",");
      flowKey = `${flowId ?? 'flow'}|${fillParent ? 'E' : 'N'}|n:${nodeSig}|e:${edgeSig}|c:${containerKey}`;
    } catch {
      flowKey = `${flowId ?? 'flow'}|${fillParent ? 'E' : 'N'}|${Date.now()}`;
    }
    // Debug logging (only in development)
    if (import.meta.env.DEV) {
      console.log("ResourceGraph graph", {
        title,
        nodes: graph.nodes.map((n) => n.id),
        edges: filteredEdges.map((e) => ({ id: e.id, source: e.source, target: e.target })),
      });
    }
  }

  // Apply preselection for seeded anchors (runs when preselectNodeIds or nodes change)
  $: (function applyPreselection() {
    const ids = new Set(preselectNodeIds ?? []);
    // Always set selected based on current ids; if empty, clear selection
    nodesStore.update((nds) => nds.map((n) => ({ ...n, selected: ids.has(n.id) })));
  })();

</script>

<section class="graph-instance">
  {#if titleLabel != null}
    <h2 class="graph-title">
      <span class={anchorError ? 'text-red-600' : ''}>{titleLabel}</span>
      {#if titleErrorCount && titleErrorCount > 0}
        <span class={anchorError ? 'text-red-600' : 'text-red-600'}>
          {' '}
          • {titleErrorCount} error{titleErrorCount === 1 ? '' : 's'}
        </span>
      {/if}
    </h2>
  {:else if title}
    <h2 class="graph-title">{title}</h2>
  {/if}

  {#if hasNodes}
    <div class={"graph-container " + containerHeightClass} bind:this={containerEl}>
      {#if enableExpand}
        <button
          class="expand-btn"
          aria-label="Expand graph"
          title="Expand"
          on:click={() => dispatch('expand')}
        >
          ⤢
        </button>
      {/if}

      {#key flowKey}
        <SvelteFlow
          id={flowId}
          nodes={nodesStore}
          edges={edgesViewStore}
          nodeTypes={nodeTypes}
          colorMode={flowColorMode}
          proOptions={{ hideAttribution: true }}
          fitView
          fitViewOptions={{ padding: 0.22, minZoom: 0.05, maxZoom: 1.25, duration: 200 }}
          preventScrolling={false}
          zoomOnScroll={false}
          panOnScroll={false}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable
          selectionOnDrag
          onlyRenderVisibleElements={false}
          defaultEdgeOptions={edgeOptions}
        >
          <Background gap={24} />
          {#if showControls}
            <Controls position="top-right" showLock={showLock} />
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
    @apply flex h-full flex-col gap-y-3;
  }

  .graph-title {
    @apply text-sm font-semibold text-foreground;
  }

  .graph-container {
    @apply relative w-full overflow-hidden rounded-lg border border-gray-200 bg-white;
  }

  .state {
    @apply flex h-[160px] w-full items-center justify-center rounded-lg border border-dashed border-gray-200 bg-white text-sm text-gray-500;
  }

  .expand-btn {
    @apply absolute right-2 top-2 z-20 h-7 w-7 rounded border bg-surface text-sm text-muted-foreground hover:bg-muted hover:text-foreground;
    line-height: 1.25rem;
  }
</style>
