<script lang="ts">
  import { Background, SvelteFlow, type Node, type Edge } from "@xyflow/svelte";
  import "@xyflow/svelte/dist/base.css";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { coerceResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SummaryCountNode from "./SummaryCountNode.svelte";
  import { writable } from "svelte/store";
  import { onMount, onDestroy } from "svelte";
  import { themeControl } from "../themes/theme-control";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { ALLOWED_FOR_GRAPH } from "./seed-utils";
  import { goto } from "$app/navigation";

  export let sources = 0;
  export let metrics = 0;
  export let models = 0;
  export let dashboards = 0;
  // Full list of resources (for selection panel)
  export let resources: V1Resource[] = [];
  // Active token to highlight: 'sources' | 'metrics' | 'models' | 'dashboards'
  export let activeToken:
    | "sources"
    | "metrics"
    | "models"
    | "dashboards"
    | null = null;

  let containerEl: HTMLDivElement | null = null;
  let containerKey = "";
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
    } catch {}
    ro = null;
  });

  // Stores for SvelteFlow
  const nodesStore = writable<Node[]>([]);
  const edgesStore = writable<Edge[]>([]);

  // Build filtered lists by kind
  let visible: V1Resource[] = [];
  $: visible = (resources || []).filter((r) => {
    if (r?.meta?.hidden) return false;
    const k = coerceResourceKind(r);
    return (
      k === ResourceKind.Source ||
      k === ResourceKind.Model ||
      k === ResourceKind.MetricsView ||
      k === ResourceKind.Explore
    );
  });
  function listFor(kind: ResourceKind): V1Resource[] {
    return visible
      .filter((r) => coerceResourceKind(r) === kind)
      .sort((a, b) =>
        (a.meta?.name?.name || "").localeCompare(b.meta?.name?.name || ""),
      );
  }
  $: srcList = listFor(ResourceKind.Source);
  $: mtrList = listFor(ResourceKind.MetricsView);
  $: mdlList = listFor(ResourceKind.Model);
  $: dshList = listFor(ResourceKind.Explore);

  function navigateTokenForNode(id: string) {
    let token: "sources" | "metrics" | "models" | "dashboards" | null = null;
    if (id === "sources") token = "sources";
    else if (id === "metrics") token = "metrics";
    else if (id === "models") token = "models";
    else if (id === "dashboards") token = "dashboards";
    if (token) goto(`/graph?seed=${token}`);
  }
  function toSeed(kind: ResourceKind, name: string) {
    const k =
      kind === ResourceKind.MetricsView
        ? "metrics"
        : kind === ResourceKind.Explore
          ? "dashboard"
          : kind === ResourceKind.Model
            ? "model"
            : "source";
    return `${k}:${name}`;
  }
  function openGraph(kind: ResourceKind, name: string) {
    const seed = toSeed(kind, name);
    goto(`/graph?seed=${encodeURIComponent(seed)}`);
  }

  // Build nodes spaced across the available width
  function buildNodes(
    width: number,
    counts: {
      sources: number;
      metrics: number;
      models: number;
      dashboards: number;
    },
    token: "sources" | "metrics" | "models" | "dashboards" | null,
  ) {
    const pad = 40;
    const eff = Math.max(120, width - pad * 2);
    const step = Math.floor(eff / 3);
    const y = 60; // center larger nodes vertically in taller canvas
    const { sources, metrics, models, dashboards } = counts;
    const isActive = (key: "sources" | "metrics" | "models" | "dashboards") =>
      token === key;
    return [
      {
        id: "sources",
        position: { x: pad + step * 0, y },
        type: "summary-count",
        selected: isActive("sources"),
        data: {
          label: "Sources",
          count: sources,
          kind: ResourceKind.Source,
          active: isActive("sources"),
        },
      },
      {
        id: "metrics",
        position: { x: pad + step * 1, y },
        type: "summary-count",
        selected: isActive("metrics"),
        data: {
          label: "Metrics",
          count: metrics,
          kind: ResourceKind.MetricsView,
          active: isActive("metrics"),
        },
      },
      {
        id: "models",
        position: { x: pad + step * 2, y },
        type: "summary-count",
        selected: isActive("models"),
        data: {
          label: "Models",
          count: models,
          kind: ResourceKind.Model,
          active: isActive("models"),
        },
      },
      {
        id: "dashboards",
        position: { x: pad + step * 3, y },
        type: "summary-count",
        selected: isActive("dashboards"),
        data: {
          label: "Dashboards",
          count: dashboards,
          kind: ResourceKind.Explore,
          active: isActive("dashboards"),
        },
      },
    ] satisfies Node[];
  }

  function buildEdges() {
    const shared = {
      type: "straight",
      sourceHandle: "out",
      targetHandle: "in",
    } as const;
    return [
      { id: "e1", source: "sources", target: "metrics", ...shared },
      { id: "e2", source: "metrics", target: "models", ...shared },
      { id: "e3", source: "models", target: "dashboards", ...shared },
    ] satisfies Edge[];
  }

  // Recompute on size or counts change
  $: {
    const width = containerEl?.clientWidth ?? 800;
    nodesStore.set(
      buildNodes(width, { sources, metrics, models, dashboards }, activeToken),
    );
    edgesStore.set(buildEdges());
  }

  const nodeTypes = { "summary-count": SummaryCountNode } as const;
  $: flowColorMode = ($themeControl === "dark" ? "dark" : "light") as
    | "dark"
    | "light";
  $: flowKey = `overview|${sources}|${metrics}|${models}|${dashboards}|${containerKey}|${flowColorMode}`;

  const edgeOptions = {
    type: "straight",
    style: "stroke:#b1b1b7;stroke-width:1.5px;opacity:0.95;",
  } as const;
</script>

<section class="summary-graph" aria-label="Resource summary graph">
  <div class="title">Overview</div>
  <div class="canvas" bind:this={containerEl}>
    {#key flowKey}
      <SvelteFlow
        id="overview-flow"
        nodes={nodesStore}
        edges={edgesStore}
        {nodeTypes}
        colorMode={flowColorMode}
        proOptions={{ hideAttribution: true }}
        fitView
        fitViewOptions={{
          padding: 0.25,
          minZoom: 0.25,
          maxZoom: 1.2,
          duration: 150,
        }}
        preventScrolling={true}
        zoomOnScroll={false}
        panOnScroll={false}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
        onlyRenderVisibleElements
        defaultEdgeOptions={edgeOptions}
        on:nodeClick={(e) =>
          navigateTokenForNode(e.detail?.node?.id || e.detail?.id)}
      >
        <Background gap={20} />
      </SvelteFlow>
    {/key}
  </div>
</section>

<style lang="postcss">
  .summary-graph {
    @apply mb-4 w-full;
  }
  .title {
    @apply text-sm font-semibold text-foreground mb-2;
  }
  .canvas {
    @apply relative w-full overflow-hidden rounded-lg border;
    border-color: var(--border, #e5e7eb);
    background-color: var(--surface, #ffffff);
    height: 260px;
  }
</style>
