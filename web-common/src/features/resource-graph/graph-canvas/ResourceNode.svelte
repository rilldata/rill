<script lang="ts">
  import { Handle, Position } from "@xyflow/svelte";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceIconMapping,
    resourceShorthandMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { ResourceNodeData } from "../shared/types";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import ConditionalTooltip from "@rilldata/web-common/components/tooltip/ConditionalTooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceNodeActions from "./ResourceNodeActions.svelte";

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

  // XYFlow injects these props for layout, but we only need them for typing support.
  const ensureFlowProps = (..._args: unknown[]) => {};
  $: ensureFlowProps(
    id,
    type,
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
    positionAbsoluteX,
    positionAbsoluteY,
  );

  $: showActions = data?.showNodeActions !== false;

  const DEFAULT_COLOR = "#6B7280";
  const DEFAULT_ICON = resourceIconMapping[ResourceKind.Model];

  $: kind = data?.kind;
  $: icon =
    kind && resourceIconMapping[kind]
      ? resourceIconMapping[kind]
      : DEFAULT_ICON;
  $: color =
    kind && resourceShorthandMapping[kind]
      ? `var(--${resourceShorthandMapping[kind]})`
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
</script>

<ConditionalTooltip
  showTooltip={hasError}
  location="top"
  distance={8}
  activeDelay={150}
>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="node"
    class:selected
    class:route-highlighted={routeHighlighted}
    class:error={hasError}
    class:root={data?.isRoot}
    style:--node-accent={color}
    style:width={width ? `${width}px` : undefined}
    style:height={height ? `${height}px` : undefined}
    data-kind={kind}
    role="button"
    tabindex="0"
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
      <svelte:component this={icon} size="16px" {color} />
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
        <p class="status" class:error={hasError}>{effectiveStatusLabel}</p>
      {/if}
    </div>
    {#if showActions}
      <div class="actions-trigger">
        <ResourceNodeActions {data} />
      </div>
    {/if}
  </div>
  <TooltipContent slot="tooltip-content" maxWidth="420px" variant="light">
    <div class="error-tooltip-content">
      <span class="error-tooltip-title">Error</span>
      <pre class="error-tooltip-message">{data?.resource?.meta
          ?.reconcileError}</pre>
    </div>
  </TooltipContent>
</ConditionalTooltip>

<style lang="postcss">
  .node {
    @apply relative border flex items-center gap-x-2 rounded-lg border bg-surface-subtle px-2.5 py-1.5 cursor-pointer shadow-sm;
    border-color: color-mix(in srgb, var(--node-accent) 60%, transparent);
    transition:
      box-shadow 120ms ease,
      border-color 120ms ease,
      transform 120ms ease,
      background 120ms ease;
  }

  .node.root {
    border-color: color-mix(in srgb, var(--node-accent) 65%, transparent);
    box-shadow:
      0 0 0 2px color-mix(in srgb, var(--node-accent) 35%, transparent),
      0 8px 18px rgba(15, 23, 42, 0.12);
    background-color: color-mix(
      in srgb,
      var(--node-accent) 8%,
      var(--surface-background, #ffffff)
    );
  }

  .node.selected {
    @apply shadow border-2;
    border-color: var(--node-accent);
  }

  .node.error {
    @apply border-red-300;
  }

  .icon-wrapper {
    @apply flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md;
  }

  .details {
    @apply flex flex-col gap-y-0.5 min-w-0;
  }

  .title {
    @apply font-medium text-sm leading-snug truncate;
  }

  .meta {
    @apply text-xs text-fg-secondary capitalize;
  }

  .status {
    @apply text-xs text-fg-secondary italic;
  }

  .status.error {
    @apply not-italic text-red-600;
  }

  .actions-trigger {
    @apply absolute right-1 top-1;
  }

  /* Error tooltip styling */
  .error-tooltip-content {
    @apply flex flex-col gap-y-1;
  }

  .error-tooltip-title {
    @apply text-xs font-semibold uppercase tracking-wide text-red-500;
  }

  .error-tooltip-message {
    @apply max-h-[200px] overflow-auto whitespace-pre-wrap text-xs;
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
