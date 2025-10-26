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
  import { buildResourceGraph } from "./build-resource-graph";
  import ResourceNode from "./ResourceNode.svelte";
  import type { ResourceNodeData } from "./types";

  export let resources: V1Resource[] = [];
  export let title: string | null = null;

  let hasNodes = false;
  const nodesStore = writable<Node<ResourceNodeData>[]>([]);
  const edgesStore = writable<Edge[]>([]);
  const edgesViewStore = writable<Edge[]>([]);

  const nodeTypes = {
    "resource-node": ResourceNode,
  };

  const edgeOptions = {
    type: "smoothstep",
    style: "stroke:#b1b1b7;stroke-width:1px;opacity:0.85;",
    pathOptions: { offset: 36, borderRadius: 8 },
  } as const;

  const HIGHLIGHT_EDGE_STYLE = "stroke:#3b82f6;stroke-width:2px;opacity:1;";
  const DIM_EDGE_STYLE = "stroke:#b1b1b7;stroke-width:1px;opacity:0.25;";

  // Reactively compute highlighted edges into a writable store that SvelteFlow can mutate.
  $: (function updateHighlightedEdges() {
    const nodes = $nodesStore as Node<ResourceNodeData>[];
    const edges = $edgesStore as Edge[];
    const selectedIds = new Set(nodes.filter((n) => n.selected).map((n) => n.id));
    if (!selectedIds.size) {
      edgesViewStore.set(edges);
      return;
    }

    const neighbors = new Map<string, Set<string>>();
    for (const e of edges) {
      if (!neighbors.has(e.source)) neighbors.set(e.source, new Set());
      if (!neighbors.has(e.target)) neighbors.set(e.target, new Set());
      neighbors.get(e.source)!.add(e.target);
      neighbors.get(e.target)!.add(e.source);
    }

    const visited = new Set<string>();
    const queue: string[] = Array.from(selectedIds);
    while (queue.length) {
      const id = queue.shift()!;
      if (visited.has(id)) continue;
      visited.add(id);
      const nbrs = neighbors.get(id);
      if (!nbrs) continue;
      for (const nb of nbrs) if (!visited.has(nb)) queue.push(nb);
    }

    const highlighted = new Set<string>();
    for (const e of edges) {
      if (visited.has(e.source) && visited.has(e.target)) highlighted.add(e.id);
    }

    edgesViewStore.set(
      edges.map((e) => ({
        ...e,
        style: highlighted.has(e.id) ? HIGHLIGHT_EDGE_STYLE : DIM_EDGE_STYLE,
      })),
    );
  })();

  $: {
    const graph = buildResourceGraph(resources ?? []);
    const nodeIds = new Set(graph.nodes.map((n) => n.id));
    const filteredEdges = graph.edges.filter(
      (e) => nodeIds.has(e.source) && nodeIds.has(e.target),
    );
    nodesStore.set(graph.nodes as Node<ResourceNodeData>[]);
    edgesStore.set(filteredEdges);
    hasNodes = graph.nodes.length > 0;
    console.log("ResourceGraph graph", {
      title,
      nodes: graph.nodes.map((n) => n.id),
      edges: filteredEdges.map((e) => ({ id: e.id, source: e.source, target: e.target })),
    });
  }

  // $: {
  //   const graph = buildResourceGraph(resources ?? []);
  //   const nodeIds = new Set(graph.nodes.map((node) => node.id));
  //   const filteredEdges = graph.edges.filter((edge) => {
  //     const hasSource = nodeIds.has(edge.source);
  //     const hasTarget = nodeIds.has(edge.target);
  //     if (!hasSource || !hasTarget) {
  //       console.warn("Filtered dangling edge", edge, { hasSource, hasTarget });
  //     }
  //     return hasSource && hasTarget;
  //   });
  //   nodesStore.set(graph.nodes);
  //   edgesStore.set(filteredEdges);
  //   hasNodes = graph.nodes.length > 0;
  //   console.log("ResourceGraph graph", {
  //     title,
  //     nodes: graph.nodes.map((node) => node.id),
  //     edges: filteredEdges.map((edge) => ({
  //       id: edge.id,
  //       source: edge.source,
  //       target: edge.target,
  //     })),
  //   });
  // }

  // No manual fit to view; use Svelte Flow's fitView prop instead.
</script>

<section class="graph-instance">
  {#if title}
    <h2 class="graph-title">{title}</h2>
  {/if}

  {#if hasNodes}
    <div class="graph-container">
      <SvelteFlow
        nodes={nodesStore}
        edges={edgesViewStore}
        nodeTypes={nodeTypes}
        proOptions={{ hideAttribution: true }}
        fitView
        fitViewOptions={{ padding: 0.2, minZoom: 0.1, maxZoom: 1.25, duration: 200 }}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable
        selectionOnDrag
        onlyRenderVisibleElements={false}
        defaultEdgeOptions={edgeOptions}
      >
        <Background gap={24} />
      </SvelteFlow>
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
    @apply h-[420px] w-full overflow-hidden rounded-lg border border-gray-200 bg-white;
  }

  .state {
    @apply flex h-[160px] w-full items-center justify-center rounded-lg border border-dashed border-gray-200 bg-white text-sm text-gray-500;
  }
</style>
