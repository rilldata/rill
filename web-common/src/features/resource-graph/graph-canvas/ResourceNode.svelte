<script lang="ts">
  import { Handle, Position, useSvelteFlow } from "@xyflow/svelte";
  import { resourceShorthandMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import type { ResourceNodeData } from "../shared/types";
  import { TEST_FAILURE_MARKER } from "../shared/resource-status";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { getInspectStore, openInspect } from "./inspect-store";

  export let id: string;
  export let type: string;
  export let data: ResourceNodeData;
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
  export let isConnectable: boolean = true;
  export let positionAbsoluteX = 0;
  export let positionAbsoluteY = 0;

  // XYFlow injects these props for layout; declaring them above prevents
  // "unknown prop" warnings. The reactive reference silences the Svelte
  // "unused export property" warning without generating meaningful work.
  // prettier-ignore
  // eslint-disable-next-line @typescript-eslint/no-unused-expressions
  $: [type, height, sourcePosition, targetPosition, dragHandle, parentId, dragging, zIndex, selectable, deletable, draggable, positionAbsoluteX, positionAbsoluteY];

  const inspectStore = getInspectStore();
  const DEFAULT_COLOR = "#6B7280";

  $: kind = data?.kind;
  $: color =
    kind && resourceShorthandMapping[kind]
      ? `var(--${resourceShorthandMapping[kind]})`
      : DEFAULT_COLOR;
  $: reconcileError = data?.resource?.meta?.reconcileError ?? "";
  $: isTestOnlyError =
    !!reconcileError && reconcileError.includes(TEST_FAILURE_MARKER);
  $: hasError = !!reconcileError && !isTestOnlyError;
  $: isPending =
    data?.resource?.meta?.reconcileStatus &&
    data.resource.meta.reconcileStatus !==
      V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: routeHighlighted = data?.routeHighlighted === true;

  const { fitView } = useSvelteFlow();

  let nodeEl: HTMLDivElement;

  function getNodeRect(): {
    x: number;
    y: number;
    width: number;
    height: number;
  } {
    if (!nodeEl) return { x: 0, y: 0, width: 0, height: 0 };
    const nodeRect = nodeEl.getBoundingClientRect();
    const container = nodeEl.closest(".graph-container");
    const containerRect = container?.getBoundingClientRect() ?? nodeRect;
    return {
      x: nodeRect.left - containerRect.left,
      y: nodeRect.top - containerRect.top,
      width: nodeRect.width,
      height: nodeRect.height,
    };
  }

  function handleClick(e?: MouseEvent) {
    if (e && (e.metaKey || e.ctrlKey || e.shiftKey)) return;
    openInspect(inspectStore, data, getNodeRect());
  }

  function handleDoubleClick() {
    fitView({ nodes: [{ id }], duration: 300, padding: 0.5 });
  }
</script>

<div
  bind:this={nodeEl}
  class="node"
  class:selected
  class:route-highlighted={routeHighlighted}
  class:error={hasError}
  class:warned={isTestOnlyError}
  class:pending={isPending}
  class:root={data?.isRoot}
  style:--node-accent={color}
  style:width={width ? `${width}px` : undefined}
  style:height={height ? `${height}px` : undefined}
  data-kind={kind}
  onclick={handleClick}
  ondblclick={handleDoubleClick}
  role="button"
  tabindex="0"
  onkeydown={(e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleClick();
    }
  }}
>
  <Handle
    id="target"
    type="target"
    position={Position.Left}
    isConnectable={isConnectable ?? true}
  />
  <Handle
    id="source"
    type="source"
    position={Position.Right}
    isConnectable={isConnectable ?? true}
  />

  <div class="title-row">
    {#if kind}<ResourceTypeBadge {kind} showLabel={false} />{/if}
    <p class="title" title={data?.label}>{data?.label}</p>
    {#if isPending}
      <LoadingSpinner size="0.7em" />
    {/if}
  </div>
</div>

<style lang="postcss">
  .node {
    @apply relative border flex items-center rounded-lg bg-surface-subtle px-2 py-1.5 cursor-pointer shadow overflow-hidden;
    max-width: 320px;
    border-color: var(--border);
    transition:
      box-shadow 120ms ease,
      border-color 120ms ease,
      transform 120ms ease,
      background 120ms ease;
  }

  .node.selected {
    @apply shadow-md;
    border-width: 2px;
    background-color: color-mix(
      in srgb,
      var(--node-accent) 14%,
      var(--surface-background, #ffffff)
    );
    border-color: color-mix(in srgb, var(--node-accent) 50%, var(--border));
  }

  .node.route-highlighted {
    border-color: color-mix(in srgb, var(--node-accent) 50%, var(--border));
    background-color: color-mix(
      in srgb,
      var(--node-accent) 4%,
      var(--surface-background, #ffffff)
    );
  }

  .node.error {
    border-color: var(--color-red-400);
  }

  .node.warned {
    border-color: var(--color-amber-400);
  }

  .node.pending {
    opacity: 0.7;
  }

  .title-row {
    @apply flex items-center gap-x-1.5 min-w-0;
  }

  .title {
    @apply font-normal text-xs leading-snug truncate flex-1 min-w-0;
  }

  /* Hide handle dots */
  :global(.svelte-flow__node[data-id]) :global(.svelte-flow__handle) {
    width: 1px;
    height: 1px;
    min-width: 1px;
    min-height: 1px;
    background: transparent;
    border: none;
    box-shadow: none;
    opacity: 0;
  }
</style>
