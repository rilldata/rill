<script lang="ts">
  import { Handle, Position, NodeToolbar } from "@xyflow/svelte";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { ResourceNodeData } from "./types";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { goto } from "$app/navigation";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";

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

  let showError = false;
  function handleClick() {
    if (hasError && data?.resource?.meta?.reconcileError) {
      showError = !showError;
    }
  }

  $: resourceName = data?.resource?.meta?.name?.name ?? "";
  $: resourceKind = kind; // already normalized ResourceKind
  $: artifact = resourceName && resourceKind
    ? fileArtifacts.findFileArtifact(resourceKind, resourceName)
    : undefined;

  function openFile(e?: MouseEvent) {
    e?.stopPropagation();
    if (artifact?.path) {
      try {
        const key = artifact.path;
        const raw = localStorage.getItem(key);
        const obj = raw ? JSON.parse(raw) || {} : {};
        obj.view = "code";
        localStorage.setItem(key, JSON.stringify(obj));
      } catch {
        // ignore storage issues; fall back to default behavior
      }
      goto(`/files${artifact.path}`);
    }
  }
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
  on:click={handleClick}
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

  <NodeToolbar isVisible={selected && !hasError && !!artifact?.path} position={Position.Top} align="center" offset={2}>
    <button
      class="toolbar-open-btn"
      aria-label="Open in code"
      title={`Open ${artifact?.path}`}
      on:click|stopPropagation={openFile}
    >
      <ExternalLink size="10px" />
    </button>
  </NodeToolbar>

  {#if showError && hasError}
    <div class="error-popover" role="alert">
      <div class="error-popover-header">
        <span class="error-title">Error</span>
        <div class="error-actions">
          {#if artifact?.path}
            <a
              href={`/files${artifact.path}`}
              class="error-open"
              on:click|stopPropagation={openFile}
              title={`Open ${artifact.path}`}
              >Open YAML</a
            >
          {/if}
          <button class="error-close" aria-label="Close error" on:click|stopPropagation={() => (showError = false)}>
            ✕
          </button>
        </div>
      </div>
      <pre class="error-message" title={data?.resource?.meta?.reconcileError}>{data?.resource?.meta?.reconcileError}</pre>
    </div>
  {/if}
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

  .error-popover {
    @apply absolute -top-2 left-1/2 z-20 w-[420px] max-w-[70vw] -translate-y-full -translate-x-1/2 rounded-md border border-red-300 bg-red-50 shadow-lg;
  }

  .error-popover-header {
    @apply flex items-center justify-between px-3 py-2 border-b border-red-200 gap-2;
  }

  .error-title {
    @apply text-xs font-semibold text-red-700 uppercase tracking-wide;
  }

  .error-close {
    @apply h-6 w-6 rounded border border-red-300 bg-white text-xs text-red-600 hover:bg-red-50 hover:text-red-700;
    line-height: 1rem;
  }

  .error-actions {
    @apply flex items-center gap-2;
  }

  .error-open {
    @apply text-xs text-red-700 underline hover:text-red-800;
  }

  .error-message {
    @apply m-3 max-h-[40vh] overflow-auto whitespace-pre-wrap text-xs text-red-700;
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

  .toolbar-open-btn {
    @apply h-6 w-6 rounded-sm border border-transparent bg-white text-gray-600 hover:bg-gray-50 hover:text-gray-800 flex items-center justify-center shadow-sm ring-1 ring-black/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-400/60;
  }

  .toolbar-open-btn :global(svg) {
    width: 12px;
    height: 12px;
    display: block;
  }

  .toolbar-open-btn :global(svg path) {
    fill: currentColor;
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
