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

  // Shrink card height so 3x3 fits comfortably
  $: containerHeightClass = fillParent ? "h-full" : "h-[260px]";

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

  // Reactively compute highlighted edges tracing strictly upstream and downstream
  // from the selected node(s). We explore both directions only at the start node(s):
  // upstream explores only incoming edges (sources) and continues going up;
  // downstream explores only outgoing edges (dependents) and continues going down.
  $: (function updateHighlightedEdges() {
    const nodes = $nodesStore as Node<ResourceNodeData>[];
    const edges = $edgesStore as Edge[];
    const selectedIds = new Set(nodes.filter((n) => n.selected).map((n) => n.id));
    if (!selectedIds.size) {
      const nodeMap = new Map(nodes.map((n) => [n.id, n]));
      edgesViewStore.set(
        edges.map((e) => {
          const s = nodeMap.get(e.source);
          const t = nodeMap.get(e.target);
          let offset = 8;
          if (s && t) {
            const sx = (s.position?.x ?? 0) + (s.width ?? 0) / 2;
            const sy = (s.position?.y ?? 0) + (s.height ?? 0);
            const tx = (t.position?.x ?? 0) + (t.width ?? 0) / 2;
            const ty = (t.position?.y ?? 0);
            const dx = Math.abs(tx - sx);
            const dy = Math.abs(ty - sy);
            if (dx < 12) offset = 4; else offset = Math.max(6, Math.min(18, Math.round(dy / 10)));
          }
          return { ...e, pathOptions: { offset, borderRadius: 6 } };
        }),
      );
      // clear route highlight flags if nothing is selected
      nodesStore.update((nds) => nds.map((n) => ({
        ...n,
        data: { ...n.data, routeHighlighted: false },
      })));
      return;
    }

    // Upstream traversal (incoming edges only)
    const upstreamVisited = new Set<string>();
    const upstreamEdgeIds = new Set<string>();
    const upQueue: string[] = Array.from(selectedIds);
    while (upQueue.length) {
      const curr = upQueue.shift()!;
      if (upstreamVisited.has(curr)) continue;
      upstreamVisited.add(curr);
      for (const e of edges) {
        if (e.target === curr) {
          upstreamEdgeIds.add(e.id);
          if (!upstreamVisited.has(e.source)) upQueue.push(e.source);
        }
      }
    }

    // Downstream traversal (outgoing edges only)
    const downstreamVisited = new Set<string>();
    const downstreamEdgeIds = new Set<string>();
    const downQueue: string[] = Array.from(selectedIds);
    while (downQueue.length) {
      const curr = downQueue.shift()!;
      if (downstreamVisited.has(curr)) continue;
      downstreamVisited.add(curr);
      for (const e of edges) {
        if (e.source === curr) {
          downstreamEdgeIds.add(e.id);
          if (!downstreamVisited.has(e.target)) downQueue.push(e.target);
        }
      }
    }

    const highlighted = new Set<string>([...upstreamEdgeIds, ...downstreamEdgeIds]);
    const highlightNodeIds = new Set<string>([
      ...upstreamVisited,
      ...downstreamVisited,
    ]);

    // Compute a dynamic offset for smoother, straighter routes.
    const nodeMap = new Map(nodes.map((n) => [n.id, n]));
    edgesViewStore.set(
      edges.map((e) => {
        const s = nodeMap.get(e.source);
        const t = nodeMap.get(e.target);
        let offset = 8; // default
        if (s && t) {
          const sx = (s.position?.x ?? 0) + (s.width ?? 0) / 2;
          const sy = (s.position?.y ?? 0) + (s.height ?? 0); // bottom handle
          const tx = (t.position?.x ?? 0) + (t.width ?? 0) / 2;
          const ty = (t.position?.y ?? 0); // top handle
          const dx = Math.abs(tx - sx);
          const dy = Math.abs(ty - sy);
          // If almost vertical, keep offset tiny; if further apart, allow a bit more to avoid kinks
          if (dx < 12) offset = 4;
          else offset = Math.max(6, Math.min(18, Math.round(dy / 10)));
        }
        return {
          ...e,
          style: highlighted.has(e.id) ? HIGHLIGHT_EDGE_STYLE : DIM_EDGE_STYLE,
          pathOptions: { offset, borderRadius: 6 },
        };
      }),
    );

    // Mark nodes along the traced paths as highlighted
    nodesStore.update((nds) => nds.map((n) => ({
      ...n,
      data: { ...n.data, routeHighlighted: highlightNodeIds.has(n.id) },
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
    console.log("ResourceGraph graph", {
      title,
      nodes: graph.nodes.map((n) => n.id),
      edges: filteredEdges.map((e) => ({ id: e.id, source: e.source, target: e.target })),
    });
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
