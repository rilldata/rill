<script lang="ts">
  import { Handle, Position } from "@xyflow/svelte";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  import { goto } from "$app/navigation";

  export let id: string;
  export let type: string;
  export let data: {
    label: string;
    count: number;
    kind: ResourceKind;
    active?: boolean;
  };
  export let selected: boolean = false;
  export let width: number | undefined = undefined;
  export let height: number | undefined = undefined;
  export let sourcePosition: Position | undefined = undefined;
  export let targetPosition: Position | undefined = undefined;
  export let dragHandle: string | undefined = undefined;
  export let parentId: string | undefined = undefined;
  export let dragging: boolean = false;
  export let zIndex = 0;
  export let selectable: boolean = true;
  export let deletable: boolean = true;
  export let draggable: boolean = false;
  export let isConnectable: boolean = false;
  export let positionAbsoluteX = 0;
  export let positionAbsoluteY = 0;

  // XYFlow passes these props when rendering custom nodes.
  const ensureFlowProps = (..._args: unknown[]) => {};
  $: ensureFlowProps(
    id,
    type,
    width,
    height,
    sourcePosition,
    targetPosition,
    dragHandle,
    parentId,
    dragging,
    zIndex,
    selectable,
    deletable,
    draggable,
    isConnectable,
    positionAbsoluteX,
    positionAbsoluteY,
  );

  $: color = resourceColorMapping[data?.kind] || "#6B7280";
  $: Icon = resourceIconMapping[data?.kind] || null;
  $: label = data?.label ?? "";
  $: count = data?.count ?? 0;
  $: isActive = data?.active ?? selected;
  $: isEmpty = count === 0;

  function navigateByKind() {
    // Don't navigate if there are no resources of this kind
    if (isEmpty) return;

    const kind = data?.kind;
    let token: string | null = null;
    if (kind === ResourceKind.Source) token = "sources";
    else if (kind === ResourceKind.MetricsView) token = "metrics";
    else if (kind === ResourceKind.Model) token = "models";
    else if (kind === ResourceKind.Explore) token = "dashboards";
    if (token) goto(`/graph?kind=${token}`);
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  class="summary-node"
  class:active={isActive}
  class:empty={isEmpty}
  style={`--summary-accent:${color}`}
  on:click={navigateByKind}
  role="button"
  tabindex={isEmpty ? -1 : 0}
  aria-disabled={isEmpty}
  on:keydown={(e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      navigateByKind();
    }
  }}
>
  <!-- connection points for flow edges (left/right) -->
  <Handle
    id="in"
    type="target"
    position={Position.Left}
    isConnectable={false}
  />
  <Handle
    id="out"
    type="source"
    position={Position.Right}
    isConnectable={false}
  />
  <div class="icon">
    {#if Icon}
      <svelte:component this={Icon} size="32px" {color} />
    {/if}
  </div>
  <div class="content">
    <div class="label" title={label}>{label}</div>
    <div class="count" aria-label={`${label} count`}>{count}</div>
  </div>
</div>

<style lang="postcss">
  .summary-node {
    @apply relative flex items-center gap-4 rounded-lg border px-5 py-4 shadow-sm min-w-[280px];
    background-color: var(--surface, #ffffff);
    border-color: color-mix(
      in srgb,
      var(--summary-accent, #94a3b8) 35%,
      var(--border, #e5e7eb)
    );
  }
  .summary-node.active {
    @apply border-2;
    border-color: color-mix(
      in srgb,
      var(--summary-accent, #94a3b8) 70%,
      var(--border, #94a3b8)
    );
    box-shadow: 0 0 0 1px
      color-mix(in srgb, var(--summary-accent, #94a3b8) 30%, transparent);
    background-color: color-mix(
      in srgb,
      var(--summary-accent, #94a3b8) 15%,
      var(--surface, #ffffff)
    );
  }
  .summary-node.empty {
    @apply opacity-50 cursor-default;
  }
  .summary-node:not(.empty):hover {
    @apply cursor-pointer;
    border-color: color-mix(
      in srgb,
      var(--summary-accent, #94a3b8) 50%,
      var(--border, #e5e7eb)
    );
  }
  .icon {
    @apply flex h-12 w-12 items-center justify-center rounded-md;
    background-color: color-mix(
      in srgb,
      var(--summary-accent, #94a3b8) 16%,
      transparent
    );
  }
  .content {
    @apply flex items-baseline gap-2;
  }
  .label {
    @apply text-base font-medium;
    color: var(--muted-foreground, #4b5563);
  }
  .count {
    @apply text-3xl font-semibold leading-tight;
    color: var(--foreground, #111827);
  }
</style>
