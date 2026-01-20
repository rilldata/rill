<script lang="ts">
  import { Handle, Position, NodeToolbar } from "@xyflow/svelte";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { ResourceNodeData } from "../shared/types";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { goto } from "$app/navigation";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";

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

  const DEFAULT_COLOR = "#6B7280";
  const DEFAULT_ICON = resourceIconMapping[ResourceKind.Model];

  $: kind = data?.kind;
  $: icon =
    kind && resourceIconMapping[kind]
      ? resourceIconMapping[kind]
      : DEFAULT_ICON;
  $: color = kind ? undefined : DEFAULT_COLOR;
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
  $: artifact =
    resourceName && resourceKind
      ? fileArtifacts.findFileArtifact(resourceKind, resourceName)
      : undefined;

  function openFile(e?: MouseEvent) {
    e?.stopPropagation();
    if (!artifact?.path) return;

    // Set code view preference for this file
    try {
      const key = artifact.path;
      const prefs = JSON.parse(localStorage.getItem(key) || "{}");
      localStorage.setItem(key, JSON.stringify({ ...prefs, view: "code" }));
    } catch (error) {
      console.warn(`Failed to save file view preference:`, error);
    }

    goto(`/files${artifact.path}`);
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  class="node"
  class:selected
  class:route-highlighted={routeHighlighted}
  class:error={hasError}
  class:root={data?.isRoot}
  style={`--node-accent:${color}`}
  style:width={width ? `${width}px` : undefined}
  data-kind={kind}
  on:click={handleClick}
  role="button"
  tabindex="0"
  on:keydown={(e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleClick();
    }
  }}
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
    <svelte:component this={icon} size="20px" {color} />
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
      <p
        class="status"
        class:error={hasError}
        title={hasError ? data?.resource?.meta?.reconcileError : undefined}
      >
        {effectiveStatusLabel}
      </p>
    {/if}
  </div>

  <NodeToolbar
    isVisible={selected && !hasError && !!artifact?.path}
    position={Position.Top}
    align="center"
    offset={4}
  >
    <button
      class="toolbar-open-btn"
      aria-label="Open in code"
      title={`Open ${artifact?.path}`}
      on:click|stopPropagation={openFile}
    >
      <ExternalLink size="12px" />
      <span>Open</span>
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
              title={`Open ${artifact.path}`}>Open YAML</a
            >
          {/if}
          <button
            class="error-close"
            aria-label="Close error"
            on:click|stopPropagation={() => (showError = false)}
          >
            ✕
          </button>
        </div>
      </div>
      <pre
        class="error-message"
        title={data?.resource?.meta?.reconcileError}>{data?.resource?.meta
          ?.reconcileError}</pre>
    </div>
  {/if}
</div>

<style lang="postcss">
  .node {
    @apply relative flex items-center gap-x-3 rounded-lg border border-accent bg-surface-container px-3 py-2 cursor-pointer shadow-sm;
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
      var(--surface, #ffffff)
    );
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
    @apply h-6 w-6 rounded border border-red-300 bg-white text-xs text-red-600;
    line-height: 1rem;
  }

  .error-close:hover {
    @apply bg-red-50 text-red-700;
  }

  .error-actions {
    @apply flex items-center gap-2;
  }

  .error-open {
    @apply text-xs text-red-700 underline;
  }

  .error-open:hover {
    @apply text-red-800;
  }

  .error-message {
    @apply m-3 max-h-[40vh] overflow-auto whitespace-pre-wrap text-xs text-red-700;
  }

  .icon-wrapper {
    @apply flex h-10 w-10 items-center justify-center rounded-md;
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

  .toolbar-open-btn {
    @apply h-7 px-3 rounded-[2px] border flex items-center justify-center gap-x-1.5 shadow-sm transition-colors;
    @apply text-xs font-medium;
    @apply bg-primary-600 text-white border-primary-600;
  }

  .toolbar-open-btn:focus {
    @apply outline-none;
  }

  .toolbar-open-btn:focus-visible {
    @apply ring-2 ring-primary-400/60 ring-offset-1;
  }

  .toolbar-open-btn:hover {
    @apply bg-primary-700 border-primary-700;
  }

  .toolbar-open-btn:active {
    @apply bg-primary-800 border-primary-800;
  }

  .toolbar-open-btn :global(svg) {
    width: 12px;
    height: 12px;
    display: block;
    flex-shrink: 0;
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
