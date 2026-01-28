<script lang="ts">
  import { Handle, Position } from "@xyflow/svelte";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceIconMapping,
    resourceShorthandMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { ResourceNodeData } from "../shared/types";
  import {
    V1ReconcileStatus,
    createRuntimeServiceCreateTrigger,
  } from "@rilldata/web-common/runtime-client";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { goto } from "$app/navigation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { GitFork } from "lucide-svelte";
  import { builderActions, getAttrs } from "bits-ui";

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

  $: resourceName = data?.resource?.meta?.name?.name ?? "";
  $: resourceKind = kind;
  // Use original kind from resource meta for artifact lookup (not coerced kind)
  // because file artifacts are stored by the resource's actual kind
  $: originalKind = (data?.resource?.meta?.name?.kind ?? kind) as ResourceKind;
  $: artifact =
    resourceName && originalKind
      ? fileArtifacts.findFileArtifact(originalKind, resourceName)
      : undefined;

  // Dropdown menu state
  let menuOpen = false;

  // Refresh functionality
  const triggerMutation = createRuntimeServiceCreateTrigger();
  $: ({ instanceId } = $runtime);
  // Source Models and Models can both be refreshed
  $: isModelOrSource =
    kind === ResourceKind.Model || kind === ResourceKind.Source;
  $: isInOverlay = (data as any)?.isOverlay === true;
  $: isRefreshing = $triggerMutation.isPending;

  function openFile() {
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

  function handleRefresh() {
    if (!isModelOrSource || !data?.resource?.meta?.name?.name || isRefreshing)
      return;

    void $triggerMutation.mutateAsync({
      instanceId,
      data: {
        models: [{ model: data.resource.meta.name.name, full: true }],
      },
    });
  }

  function handleViewGraph() {
    if (!data?.resource?.meta?.name) return;

    const resourceKindName = data.resource.meta.name.kind;
    const resourceNameValue = data.resource.meta.name.name;

    // Determine kind token for URL
    let kindToken = "models";
    if (resourceKindName === "rill.runtime.v1.MetricsView") {
      kindToken = "metrics";
    } else if (
      resourceKindName === "rill.runtime.v1.Explore" ||
      resourceKindName === "rill.runtime.v1.Canvas"
    ) {
      kindToken = "dashboards";
    }

    // Build expanded ID (ResourceKind:Name format)
    const expandedId = encodeURIComponent(
      `${resourceKindName}:${resourceNameValue}`,
    );

    goto(`/graph?kind=${kindToken}&expanded=${expandedId}`);
  }
</script>

{#if hasError}
  <Tooltip location="top" distance={8} activeDelay={150}>
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

      <!-- Dropdown menu in top-right -->
      {#if !isInOverlay}
        <div class="node-menu">
          <DropdownMenu.Root bind:open={menuOpen}>
            <DropdownMenu.Trigger asChild let:builder>
              <button
                class="menu-trigger"
                class:visible={menuOpen}
                aria-label="Node actions"
                use:builderActions={{ builders: [builder] }}
                {...getAttrs([builder])}
                on:click|stopPropagation
              >
                <MoreHorizontal size="14px" />
              </button>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start" side="right" sideOffset={4}>
              {#if artifact?.path}
                <NavigationMenuItem on:click={openFile}>
                  <EditIcon slot="icon" />
                  Edit YAML
                </NavigationMenuItem>
              {/if}
              {#if isModelOrSource}
                <NavigationMenuItem
                  on:click={handleRefresh}
                  disabled={isRefreshing}
                >
                  <RefreshIcon slot="icon" size="14px" />
                  {isRefreshing ? "Refreshing..." : "Refresh"}
                </NavigationMenuItem>
              {/if}
              <NavigationMenuItem on:click={handleViewGraph}>
                <GitFork slot="icon" size="14px" />
                View lineage
              </NavigationMenuItem>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>
      {/if}

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
        <p class="status error">{effectiveStatusLabel}</p>
      </div>
    </div>
    <TooltipContent slot="tooltip-content" maxWidth="420px">
      <div class="error-tooltip-content">
        <span class="error-tooltip-title">Error</span>
        <pre class="error-tooltip-message">{data?.resource?.meta
            ?.reconcileError}</pre>
      </div>
    </TooltipContent>
  </Tooltip>
{:else}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="node"
    class:selected
    class:route-highlighted={routeHighlighted}
    class:root={data?.isRoot}
    style:--node-accent={color}
    style:width={width ? `${width}px` : undefined}
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

    <!-- Dropdown menu in top-right -->
    {#if !isInOverlay}
      <div class="node-menu">
        <DropdownMenu.Root bind:open={menuOpen}>
          <DropdownMenu.Trigger asChild let:builder>
            <button
              class="menu-trigger"
              class:visible={menuOpen}
              aria-label="Node actions"
              use:builderActions={{ builders: [builder] }}
              {...getAttrs([builder])}
              on:click|stopPropagation
            >
              <MoreHorizontal size="14px" />
            </button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="start" side="right" sideOffset={4}>
            {#if artifact?.path}
              <NavigationMenuItem on:click={openFile}>
                <EditIcon slot="icon" />
                Edit YAML
              </NavigationMenuItem>
            {/if}
            {#if isModelOrSource}
              <NavigationMenuItem
                on:click={handleRefresh}
                disabled={isRefreshing}
              >
                <RefreshIcon slot="icon" size="14px" />
                {isRefreshing ? "Refreshing..." : "Refresh"}
              </NavigationMenuItem>
            {/if}
            <NavigationMenuItem on:click={handleViewGraph}>
              <GitFork slot="icon" size="14px" />
              View lineage
            </NavigationMenuItem>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    {/if}

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
        <p class="status">{effectiveStatusLabel}</p>
      {/if}
    </div>
  </div>
{/if}

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

  /* Node menu (... dropdown) */
  .node-menu {
    @apply absolute top-1 right-1 z-10;
  }

  .menu-trigger {
    @apply h-6 w-6 rounded flex items-center justify-center;
    @apply text-fg-secondary bg-transparent;
    @apply opacity-0 transition-opacity duration-150;
  }

  /* Show on hover or when menu is open */
  .node:hover .menu-trigger,
  .menu-trigger.visible {
    @apply opacity-100;
  }

  .menu-trigger:hover {
    @apply bg-surface-muted text-fg-primary;
  }

  .menu-trigger:focus {
    @apply outline-none;
  }

  .menu-trigger:focus-visible {
    @apply ring-2 ring-primary-400/60;
  }

  /* Error tooltip styling */
  .error-tooltip-content {
    @apply flex flex-col gap-y-1;
  }

  .error-tooltip-title {
    @apply text-xs font-semibold text-red-400 uppercase tracking-wide;
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
