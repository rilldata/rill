<script lang="ts">
  import {
    Background,
    SvelteFlow,
    type Node,
    type Edge,
    type NodeTypes,
  } from "@xyflow/svelte";
  import "@xyflow/svelte/dist/base.css";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SummaryNode from "./SummaryNode.svelte";
  import { writable } from "svelte/store";
  import { onMount, onDestroy } from "svelte";
  import { themeControl } from "../../themes/theme-control";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { goto } from "$app/navigation";

  export let metrics = 0;
  export let models = 0;
  export let dashboards = 0;
  // Full list of resources (for selection panel)
  export let resources: V1Resource[] = [];
  // Active token to highlight: 'metrics' | 'models' | 'dashboards'
  export let activeToken:
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
    } catch (error) {
      console.debug(
        "[SummaryCountsGraph] Failed to disconnect observer",
        error,
      );
    }
    ro = null;
  });

  // Stores for SvelteFlow
  const nodesStore = writable<Node[]>([]);
  const edgesStore = writable<Edge[]>([]);

  function navigateTokenForNode(id: string) {
    let token: "metrics" | "models" | "dashboards" | null = null;
    let count = 0;

    if (id === "metrics") {
      token = "metrics";
      count = metrics;
    } else if (id === "models") {
      token = "models";
      count = models;
    } else if (id === "dashboards") {
      token = "dashboards";
      count = dashboards;
    }

    // Only navigate if there are resources of this kind
    if (token && count > 0) {
      goto(`/graph?kind=${token}`);
    }
  }

  // Build nodes spaced across the available width
  // Sources and Models are merged (Source is deprecated)
  function buildNodes(
    width: number,
    counts: {
      metrics: number;
      models: number;
      dashboards: number;
    },
    token: "metrics" | "models" | "dashboards" | null,
  ) {
    const pad = 40;
    const eff = Math.max(120, width - pad * 2);
    const step = Math.floor(eff / 2);
    const y = 60; // center larger nodes vertically in taller canvas
    const { metrics, models, dashboards } = counts;
    const isActive = (key: "metrics" | "models" | "dashboards") =>
      token === key;
    return [
      {
        id: "models",
        position: { x: pad + step * 0, y },
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
        id: "dashboards",
        position: { x: pad + step * 2, y },
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
      { id: "e1", source: "models", target: "metrics", ...shared },
      { id: "e2", source: "metrics", target: "dashboards", ...shared },
    ] satisfies Edge[];
  }

  // Recompute on size or counts change
  $: {
    const width = containerEl?.clientWidth ?? 800;
    nodesStore.set(
      buildNodes(width, { metrics, models, dashboards }, activeToken),
    );
    edgesStore.set(buildEdges());
  }

  const nodeTypes: NodeTypes = {
    "summary-count": SummaryNode as NodeTypes[string],
  };
  $: flowColorMode = ($themeControl === "dark" ? "dark" : "light") as
    | "dark"
    | "light";
  $: flowKey = `overview|${models}|${metrics}|${dashboards}|${containerKey}|${flowColorMode}`;

  const edgeOptions = {
    type: "straight",
    style: "stroke:#b1b1b7;stroke-width:1.5px;opacity:0.95;",
  } as const;
</script>

<section
  class="summary-graph"
  aria-label="Resource summary graph"
  data-total-resources={resources.length}
>
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
          duration: 0,
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
    @apply text-sm font-semibold text-fg-primary mb-2;
  }
  .canvas {
    @apply relative w-full overflow-hidden rounded-lg border;
    border-color: var(--border, #e5e7eb);
    background-color: var(--surface-background, #ffffff);
    height: 260px;
  }
</style>
