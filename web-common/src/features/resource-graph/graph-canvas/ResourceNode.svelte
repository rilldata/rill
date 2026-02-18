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
  import { getRelativeTime } from "@rilldata/web-common/lib/time/relative-time";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { Unlock, AlertTriangle } from "lucide-svelte";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import type { ComponentType, SvelteComponent } from "svelte";

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
  $: metadata = data?.metadata;
  $: icon =
    kind && resourceIconMapping[kind]
      ? resourceIconMapping[kind]
      : DEFAULT_ICON;
  $: color =
    kind && resourceShorthandMapping[kind]
      ? `var(--${resourceShorthandMapping[kind]})`
      : DEFAULT_COLOR;
  $: reconcileStatus = data?.resource?.meta?.reconcileStatus;
  $: reconcileError = data?.resource?.meta?.reconcileError ?? "";
  // Test failures propagate as "tests failed:..." on the model itself and
  // "Error in dependency <name>: tests failed:..." on downstream resources.
  // Treat both as warnings (shown via check indicator), not node errors.
  $: isTestOnlyError =
    !!reconcileError && reconcileError.includes("tests failed:");
  $: hasError = !!reconcileError && !isTestOnlyError;
  $: isIdle = reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: isPending =
    reconcileStatus &&
    reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: routeHighlighted = (data as any)?.routeHighlighted === true;

  // Derived metadata for display
  $: lastRefreshed = metadata?.lastRefreshedOn
    ? getRelativeTime(metadata.lastRefreshedOn)
    : null;

  $: isSourceOrModel =
    kind === ResourceKind.Source || kind === ResourceKind.Model;
  $: isMetricsView = kind === ResourceKind.MetricsView;
  $: isExplore = kind === ResourceKind.Explore;
  $: isCanvas = kind === ResourceKind.Canvas;

  $: measuresCount = metadata?.measures?.length ?? 0;
  $: dimensionsCount = metadata?.dimensions?.length ?? 0;
  $: testCount = metadata?.testCount ?? 0;
  $: schedule = metadata?.scheduleDescription ?? null;
  $: isIncremental = metadata?.incremental === true;
  $: testHasErrors = (metadata?.testErrors?.length ?? 0) > 0;
  $: componentCount = metadata?.componentCount ?? 0;
  $: hasSecurityRules = metadata?.hasSecurityRules === true;
  $: connector = metadata?.connector ?? null;
  $: connectorIcon = (
    connector &&
    connectorIconMapping[connector as keyof typeof connectorIconMapping]
      ? connectorIconMapping[connector as keyof typeof connectorIconMapping]
      : null
  ) as ComponentType<SvelteComponent<{ size?: string }>> | null;
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
    class:warned={isTestOnlyError}
    class:pending={isPending}
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

    <!-- Title row: icon + name + actions -->
    <div class="title-row">
      <span class="inline-icon" style={`color:${color}`}>
        <svelte:component this={icon} size="12px" color="currentColor" />
      </span>
      <p class="title" title={data?.label}>{data?.label}</p>
      {#if showActions}
        <div class="actions-trigger">
          <ResourceNodeActions {data} />
        </div>
      {/if}
    </div>

    <!-- Content rows -->
    <div class="content">
      {#if isSourceOrModel}
        <!-- Source/Model Row 1: Kind (left) · Last refreshed (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-kind">
            {#if kind}{displayResourceKind(kind)}{:else}Unknown{/if}
          </span>
          {#if lastRefreshed}
            <span class="meta-detail">{lastRefreshed}</span>
          {/if}
        </div>
        <!-- Source/Model Row 2: Connector + Badges (left) · Check (right) -->
        <div class="meta-row meta-row-spread">
          <span class="badge-group">
            {#if connectorIcon}
              <svelte:component this={connectorIcon} size="10px" />
            {/if}
            {#if metadata?.isMaterialized}
              <span class="badge">Materialized</span>
            {/if}
            <span class="badge">{metadata?.isSqlModel ? "SQL" : "YAML"}</span>
          </span>
          <span
            class="check-indicator"
            class:checks-none={testCount === 0}
            class:checks-pass={testCount > 0 && !testHasErrors}
            class:checks-fail={testCount > 0 && testHasErrors}
          >
            {#if testHasErrors}
              <AlertTriangle size="10px" />
            {:else}
              <CheckCircle size="10px" color="currentColor" />
            {/if}
            {testCount}
          </span>
        </div>
      {:else if isMetricsView}
        <!-- MetricsView Row 1: Kind (left) · Time (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-kind">{displayResourceKind(kind)}</span>
          {#if lastRefreshed}
            <span class="meta-detail">{lastRefreshed}</span>
          {/if}
        </div>
        <!-- MetricsView Row 2: Measures/Dims (left) · Lock (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {measuresCount} meas, {dimensionsCount} dims
          </span>
          <span class="lock-indicator" class:secured={hasSecurityRules}>
            {#if hasSecurityRules}
              <Lock size="10px" color="currentColor" />
            {:else}
              <Unlock size="10px" />
            {/if}
          </span>
        </div>
      {:else if isExplore}
        <!-- Explore Row 1: Kind (left) · Time (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-kind">{displayResourceKind(kind)}</span>
          {#if lastRefreshed}
            <span class="meta-detail">{lastRefreshed}</span>
          {/if}
        </div>
        <!-- Explore Row 2: Measures/Dims (left) · Lock (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {metadata?.exploreMeasuresAll
              ? "all"
              : (metadata?.exploreMeasuresCount ?? 0)} meas,
            {metadata?.exploreDimensionsAll
              ? "all"
              : (metadata?.exploreDimensionsCount ?? 0)} dims
          </span>
          <span class="lock-indicator" class:secured={hasSecurityRules}>
            {#if hasSecurityRules}
              <Lock size="10px" color="currentColor" />
            {:else}
              <Unlock size="10px" />
            {/if}
          </span>
        </div>
      {:else if isCanvas}
        <!-- Canvas Row 1: Kind (left) · Time (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-kind">{displayResourceKind(kind)}</span>
          {#if lastRefreshed}
            <span class="meta-detail">{lastRefreshed}</span>
          {/if}
        </div>
        <!-- Canvas Row 2: Components (left) · Lock (right) -->
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {componentCount} component{componentCount !== 1 ? "s" : ""}
          </span>
          <span class="lock-indicator" class:secured={hasSecurityRules}>
            {#if hasSecurityRules}
              <Lock size="10px" color="currentColor" />
            {:else}
              <Unlock size="10px" />
            {/if}
          </span>
        </div>
      {:else}
        <!-- Fallback -->
        <div class="meta-row">
          <span class="meta-kind">Unknown</span>
        </div>
      {/if}
    </div>
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
    @apply relative border flex flex-col rounded-lg bg-surface-subtle px-2.5 py-2 cursor-pointer shadow-sm;
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
    @apply border-red-400;
    box-shadow:
      0 0 0 2px rgba(239, 68, 68, 0.25),
      0 4px 12px rgba(239, 68, 68, 0.15);
    background-color: color-mix(
      in srgb,
      #ef4444 5%,
      var(--surface-background, #ffffff)
    );
  }

  .node.warned {
    @apply border-amber-400;
    box-shadow:
      0 0 0 2px rgba(245, 158, 11, 0.2),
      0 4px 12px rgba(245, 158, 11, 0.1);
    background-color: color-mix(
      in srgb,
      #f59e0b 5%,
      var(--surface-background, #ffffff)
    );
  }

  .node.pending {
    border-color: color-mix(in srgb, #eab308 60%, transparent);
    border-style: dashed;
  }

  /* Title row */
  .title-row {
    @apply flex items-center gap-x-1.5 min-w-0;
  }

  .inline-icon {
    @apply flex-shrink-0 flex items-center;
  }

  .title {
    @apply font-normal text-xs leading-snug truncate flex-1 min-w-0;
  }

  .actions-trigger {
    @apply flex-shrink-0 ml-auto;
  }

  /* Content section below title */
  .content {
    @apply flex flex-col gap-y-0.5 mt-1;
  }

  .meta-row {
    @apply flex items-center gap-x-1.5 text-[11px] text-fg-secondary leading-tight;
  }

  .meta-row-spread {
    @apply justify-between;
  }

  .meta-kind {
    @apply capitalize inline-flex items-center gap-x-1;
  }

  .meta-detail {
    @apply text-fg-muted truncate;
  }

  .badge-group {
    @apply inline-flex items-center gap-x-1.5;
  }

  .badge {
    @apply inline-flex items-center px-1 py-px rounded text-[10px] font-medium bg-surface-subtle text-fg-secondary;
    border: 1px solid color-mix(in srgb, var(--node-accent) 25%, transparent);
  }

  /* Check indicator with icon */
  .check-indicator {
    @apply inline-flex items-center gap-x-0.5 text-[10px] font-medium;
  }

  .check-indicator.checks-none {
    @apply text-fg-muted opacity-40;
  }

  .check-indicator.checks-pass {
    @apply text-green-600;
  }

  .check-indicator.checks-fail {
    @apply text-amber-600;
  }

  /* Lock/unlock indicator */
  .lock-indicator {
    @apply flex items-center text-fg-muted opacity-40;
  }

  .lock-indicator.secured {
    @apply text-amber-600 opacity-100;
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
