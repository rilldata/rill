<script lang="ts">
  import { Handle, Position, useSvelteFlow } from "@xyflow/svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceShorthandMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import type { ResourceNodeData } from "../shared/types";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import ConditionalTooltip from "@rilldata/web-common/components/tooltip/ConditionalTooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceNodeActions from "./ResourceNodeActions.svelte";

  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { AlertTriangle, Zap, Layers, Clock } from "lucide-svelte";
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

  let actionsRef: { open: () => void } | undefined;

  function handleContextMenu(e: MouseEvent) {
    e.preventDefault();
    actionsRef?.open();
  }

  const DEFAULT_COLOR = "#6B7280";

  $: kind = data?.kind;
  $: metadata = data?.metadata;
  $: color =
    kind && resourceShorthandMapping[kind]
      ? `var(--${resourceShorthandMapping[kind]})`
      : DEFAULT_COLOR;
  $: reconcileStatus = data?.resource?.meta?.reconcileStatus;
  $: reconcileError = data?.resource?.meta?.reconcileError ?? "";
  // Test failures propagate as "tests failed:..." on the model itself and
  // "Error in dependency <name>: tests failed:..." on downstream resources.
  // Treat both as warnings (shown via check indicator), not node errors.
  // String matching is necessary because downstream nodes only receive the
  // propagated error string; the structured `testErrors` field is only on the
  // originating model. If the backend message format changes, update this constant.
  const TEST_FAILURE_MARKER = "tests failed:";
  $: isTestOnlyError =
    !!reconcileError && reconcileError.includes(TEST_FAILURE_MARKER);
  $: hasError = !!reconcileError && !isTestOnlyError;
  $: isPending =
    reconcileStatus &&
    reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: routeHighlighted = (data as any)?.routeHighlighted === true;

  $: isSourceOrModel =
    kind === ResourceKind.Source || kind === ResourceKind.Model;
  $: isMetricsView = kind === ResourceKind.MetricsView;
  $: isExplore = kind === ResourceKind.Explore;
  $: isCanvas = kind === ResourceKind.Canvas;
  $: isConnector = kind === ResourceKind.Connector;

  $: measuresCount = metadata?.measures?.length ?? 0;
  $: dimensionsCount = metadata?.dimensions?.length ?? 0;
  $: testCount = metadata?.testCount ?? 0;
  $: testHasErrors = (metadata?.testErrors?.length ?? 0) > 0;
  $: componentCount = metadata?.componentCount ?? 0;
  $: hasSecurityRules = metadata?.hasSecurityRules === true;
  $: isIncremental = metadata?.incremental === true;
  $: isPartitioned = metadata?.partitioned === true;
  $: hasSchedule = metadata?.hasSchedule === true;
  $: connector = metadata?.connector ?? null;
  $: connectorIcon = (
    connector &&
    connectorIconMapping[connector as keyof typeof connectorIconMapping]
      ? connectorIconMapping[connector as keyof typeof connectorIconMapping]
      : null
  ) as ComponentType<SvelteComponent<{ size?: string }>> | null;

  $: filePath =
    data?.resource?.meta?.filePaths?.[0]?.replace(/^\//, "") ?? null;
  $: checkTooltip =
    testCount === 0
      ? "No checks"
      : testHasErrors
        ? `${testCount} check${testCount > 1 ? "s" : ""} failed`
        : `${testCount} check${testCount > 1 ? "s" : ""} passed`;

  // Connector node metadata
  $: connectorDriver = metadata?.connectorDriver ?? null;
  $: driverIcon = (
    connectorDriver &&
    connectorIconMapping[connectorDriver as keyof typeof connectorIconMapping]
      ? connectorIconMapping[connectorDriver as keyof typeof connectorIconMapping]
      : null
  ) as ComponentType<SvelteComponent<{ size?: string }>> | null;

  const { fitView } = useSvelteFlow();

  function handleDoubleClick() {
    fitView({ nodes: [{ id }], duration: 300, padding: 0.5 });
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleDoubleClick();
    }
  }
</script>

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
  on:dblclick={handleDoubleClick}
  on:keydown={handleKeydown}
  on:contextmenu={handleContextMenu}
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

  <ConditionalTooltip
    showTooltip={hasError}
    location="top"
    distance={8}
    activeDelay={150}
  >
    <!-- Title row: kind badge + name + actions -->
    <div class="title-row">
      {#if kind}<ResourceTypeBadge {kind} />{/if}
      <p class="title" title={data?.label}>{data?.label}</p>
      {#if showActions}
        <div class="actions-trigger">
          <ResourceNodeActions bind:this={actionsRef} {data} />
        </div>
      {/if}
    </div>

    <!-- Content row -->
    <div class="content">
      {#if isSourceOrModel}
        {@const rightIndicators = [
          metadata?.isMaterialized ? { type: "materialized" } : null,
          isIncremental ? { type: "incremental" } : null,
          isPartitioned ? { type: "partitioned" } : null,
          hasSchedule ? { type: "schedule" } : null,
          testCount > 0 ? { type: "checks" } : null,
        ].filter(Boolean).slice(0, 3)}
        <div class="meta-row meta-row-spread">
          <span class="badge-group">
            {#if connectorIcon && kind === ResourceKind.Source}
              <span class="accent-icon" title={connector}>
                <svelte:component this={connectorIcon} size="10px" />
              </span>
            {/if}
          </span>
          <span class="badge-group">
            {#each rightIndicators as ind}
              {#if ind?.type === "materialized"}
                <span class="badge" title="Materialized">Materialized</span>
              {:else if ind?.type === "incremental"}
                <span class="icon-indicator" title="Incremental">
                  <Zap size="10px" />
                </span>
              {:else if ind?.type === "partitioned"}
                <span class="icon-indicator" title="Partitioned">
                  <Layers size="10px" />
                </span>
              {:else if ind?.type === "schedule"}
                <span class="icon-indicator" title="Scheduled">
                  <Clock size="10px" />
                </span>
              {:else if ind?.type === "checks"}
                <span
                  class="check-indicator"
                  class:checks-pass={!testHasErrors}
                  class:checks-fail={testHasErrors}
                  title={checkTooltip}
                >
                  {#if testHasErrors}
                    <AlertTriangle size="10px" />
                  {:else}
                    <CheckCircle size="10px" color="currentColor" />
                  {/if}
                  {testCount}
                </span>
              {/if}
            {/each}
          </span>
        </div>
      {:else if isMetricsView}
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {measuresCount} meas, {dimensionsCount} dims
          </span>
          {#if hasSecurityRules}
            <span class="lock-indicator secured" title="Security policy defined">
              <Lock size="10px" color="currentColor" />
            </span>
          {/if}
        </div>
      {:else if isExplore}
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {metadata?.exploreMeasuresAll
              ? "all"
              : (metadata?.exploreMeasuresCount ?? 0)} meas,
            {metadata?.exploreDimensionsAll
              ? "all"
              : (metadata?.exploreDimensionsCount ?? 0)} dims
          </span>
          {#if hasSecurityRules}
            <span class="lock-indicator secured" title="Security policy defined">
              <Lock size="10px" color="currentColor" />
            </span>
          {/if}
        </div>
      {:else if isCanvas}
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {componentCount} component{componentCount !== 1 ? "s" : ""}
          </span>
          {#if hasSecurityRules}
            <span class="lock-indicator secured" title="Security policy defined">
              <Lock size="10px" color="currentColor" />
            </span>
          {/if}
        </div>
      {:else if isConnector}
        {#if driverIcon}
          <div class="meta-row">
            <span title={connectorDriver}>
              <svelte:component this={driverIcon} size="12px" />
            </span>
          </div>
        {/if}
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
</div>

<style lang="postcss">
  .node {
    @apply relative border flex flex-col justify-between rounded-lg bg-surface-subtle pl-3.5 pr-2.5 py-2 cursor-pointer shadow-sm overflow-hidden;
    border-color: var(--border);
    transition:
      box-shadow 120ms ease,
      border-color 120ms ease,
      transform 120ms ease,
      background 120ms ease;
  }

  /* Left-edge accent stripe */
  .node::before {
    content: "";
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: var(--node-accent);
  }

  .node.root {
    border-color: var(--border);
    box-shadow: 0 8px 18px rgba(15, 23, 42, 0.12);
    background-color: var(--surface-subtle);
  }

  .node.selected {
    @apply shadow-md;
    background-color: color-mix(in srgb, var(--node-accent) 6%, var(--surface-background, #ffffff));
    border-color: color-mix(in srgb, var(--node-accent) 30%, var(--border));
  }

  .node.error {
    border-color: var(--color-red-400);
  }

  .node.error::before,
  .node.warned::before {
    display: none;
  }

  .node.warned {
    border-color: var(--color-amber-400);
  }

  .node.pending {
    border-color: color-mix(in srgb, var(--color-yellow-500) 60%, var(--surface-background, #ffffff));
    border-style: dashed;
  }

  /* Title row */
  .title-row {
    @apply flex items-center gap-x-1.5 min-w-0;
  }

  .title {
    @apply font-normal text-xs leading-snug truncate flex-1 min-w-0;
  }

  .actions-trigger {
    @apply flex-shrink-0 ml-auto;
  }

  /* Content section below title */
  .content {
    @apply flex flex-col gap-y-0.5;
  }

  .meta-row {
    @apply flex items-center gap-x-1.5 text-[11px] text-fg-secondary leading-tight;
  }

  .meta-row-spread {
    @apply justify-between;
  }

  .meta-detail {
    @apply text-fg-muted truncate;
  }

  .badge-group {
    @apply inline-flex items-center gap-x-1.5;
  }

  .badge {
    @apply inline-flex items-center px-1 py-px rounded text-[10px] font-medium bg-surface-subtle text-fg-secondary;
    border: 1px solid var(--border);
  }

  .icon-indicator {
    @apply inline-flex items-center text-fg-muted;
  }

  /* Check indicator with icon */
  .check-indicator {
    @apply inline-flex items-center gap-x-0.5 text-[10px] font-medium;
  }

  .check-indicator.checks-pass {
    @apply text-green-600;
  }

  .check-indicator.checks-fail {
    @apply text-amber-600;
  }

  /* Lock indicator (only rendered when security rules are present) */
  .lock-indicator.secured {
    @apply flex items-center text-amber-600;
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

  /* Hide handle dots — edges connect as plain lines with no anchors */
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
