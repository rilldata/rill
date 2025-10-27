<script lang="ts">
  import { Handle, Position } from "@xyflow/svelte";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { ResourceNodeData } from "./types";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";

  export let data: ResourceNodeData;
  export let selected = false;

  // Accept Svelte Flow injected props (we don't use them directly, but this avoids warnings).
  export let id: string;
  export let type: string;
  export let width: number | undefined = undefined;
  export let height: number | undefined = undefined;
  export let draggable = false;
  export let dragHandle: string | undefined = undefined;
  export let dragging = false;
  export let selectable = true;
  export let deletable = true;
  export let isConnectable = true;
  export let sourcePosition: Position | undefined = undefined;
  export let targetPosition: Position | undefined = undefined;
  export let positionAbsoluteX = 0;
  export let positionAbsoluteY = 0;

  // Props injected by Svelte Flow (unused but need to exist to silence warnings).
  export let zIndex = 0;
  export let parentId: string | undefined = undefined;

  const DEFAULT_COLOR = "#6B7280";
  const DEFAULT_ICON = resourceIconMapping[ResourceKind.Model];

  $: kind = data?.kind;
  $: icon =
    kind && resourceIconMapping[kind]
      ? resourceIconMapping[kind]
      : DEFAULT_ICON;
  $: color =
    kind && resourceColorMapping[kind]
      ? resourceColorMapping[kind]
      : DEFAULT_COLOR;
  $: reconcileStatus = data?.resource?.meta?.reconcileStatus;
  $: hasError = !!data?.resource?.meta?.reconcileError;
  $: isIdle = reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: statusLabel =
    reconcileStatus && !isIdle
      ? reconcileStatus
          ?.replace("RECONCILE_STATUS_", "")
          ?.toLowerCase()
          ?.replaceAll("_", " ")
      : undefined;
  $: effectiveStatusLabel = hasError ? "error" : statusLabel;
  $: routeHighlighted = (data as any)?.routeHighlighted === true;
  $: void [
    id,
    type,
    width,
    height,
    draggable,
    dragHandle,
    dragging,
    selectable,
    deletable,
    sourcePosition,
    targetPosition,
    positionAbsoluteX,
    positionAbsoluteY,
    zIndex,
    parentId,
  ];
</script>

<div
  class="node"
  class:selected
  class:route-highlighted={routeHighlighted}
  class:error={hasError}
  style={`--node-accent:${color}`}
  data-kind={kind}
>
  <Handle
    id="target"
    type="target"
    position={Position.Top}
    isConnectable={isConnectable ?? true}
  />
  <Handle
    id="source"
    type="source"
    position={Position.Bottom}
    isConnectable={isConnectable ?? true}
  />
  <div class="icon-wrapper" style={`background:${color}20`}>
    <svelte:component this={icon} size="20px" color={color} />
  </div>
  <div class="details">
    <p class="title" title={data?.label}>{data?.label}</p>
    <p class="meta">
      {#if kind}
        {displayResourceKind(kind)}
      {:else}
        Unknown
      {/if}
    </p>
    {#if effectiveStatusLabel}
      <p class="status" class:error={hasError} title={hasError ? data?.resource?.meta?.reconcileError : undefined}>
        {effectiveStatusLabel}
      </p>
    {/if}
  </div>
</div>

<style lang="postcss">
  .node {
    @apply relative flex items-center gap-x-3 rounded-lg border border-accent bg-surface px-3 py-2 cursor-pointer shadow-sm;
    border-color: color-mix(in srgb, var(--node-accent) 60%, transparent);
    transition: box-shadow 120ms ease, border-color 120ms ease,
      transform 120ms ease, background 120ms ease;
  }

  .node.selected {
    @apply shadow border-2;
    border-color: var(--node-accent);
    transform: translateY(-1px);
  }

  .node.route-highlighted {
    @apply shadow border-2;
    border-color: var(--node-accent);
    transform: translateY(-1px);
  }

  .node.error {
    @apply border-red-300;
  }

  .icon-wrapper {
    @apply flex h-10 w-10 items-center justify-center rounded-md;
  }

  .details {
    @apply flex flex-col gap-y-0.5;
  }

  .title {
    @apply font-medium text-sm leading-snug truncate;
  }

  .meta {
    @apply text-xs text-gray-500 capitalize;
  }

  .status {
    @apply text-xs text-gray-400 italic;
  }

  .status.error {
    @apply not-italic text-red-600;
  }

  /* Make handles small circular dots, tinted by the node accent */
  :global(.svelte-flow__node[data-id]) :global(.svelte-flow__handle) {
    width: 6px;
    height: 6px;
    min-width: 6px;
    min-height: 6px;
    border-radius: 9999px;
    background-color: color-mix(in srgb, var(--node-accent) 18%, #ffffff);
    border: 1px solid color-mix(in srgb, var(--node-accent) 55%, #b1b1b7);
    box-shadow: 0 0 0 1px #ffffff;
    opacity: 1;
  }
</style>
