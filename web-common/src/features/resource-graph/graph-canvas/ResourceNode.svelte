<script lang="ts">
  import { Handle, Position } from "@xyflow/svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceShorthandMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { goto } from "$app/navigation";
  import { tokenForKind } from "../navigation/seed-parser";
  import type { ResourceNodeData } from "../shared/types";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import ConditionalTooltip from "@rilldata/web-common/components/tooltip/ConditionalTooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceNodeActions from "./ResourceNodeActions.svelte";

  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { Unlock, AlertTriangle, Zap, Layers, Clock } from "lucide-svelte";
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

  function handleDoubleClick() {
    const name = data?.resource?.meta?.name?.name;
    const kindToken = tokenForKind(kind);
    const params = new URLSearchParams();
    if (kindToken) params.set("kind", kindToken);
    if (name) params.set("resource", name);
    goto(`/graph?${params.toString()}`);
  }
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
    on:dblclick={handleDoubleClick}
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

    <!-- Title row: kind badge + name + actions -->
    <div class="title-row">
      {#if kind}<ResourceTypeBadge {kind} />{/if}
      <p class="title" title={data?.label}>{data?.label}</p>
      {#if showActions}
        <div class="actions-trigger">
          <ResourceNodeActions {data} />
        </div>
      {/if}
    </div>

    <!-- Content row -->
    <div class="content">
      {#if isSourceOrModel}
        <div class="meta-row meta-row-spread">
          <span class="badge-group">
            {#if connectorIcon && kind === ResourceKind.Source}
              <span title={connector}>
                <svelte:component this={connectorIcon} size="10px" />
              </span>
            {/if}
            {#if metadata?.isMaterialized}
              <span class="badge" title="Materialized">Materialized</span>
            {/if}
            <span class="badge" title={filePath}
              >{metadata?.isSqlModel ? "SQL" : "YAML"}</span
            >
          </span>
          <span class="icon-group">
            {#if metadata?.incremental}
              <span class="icon-indicator" title="Incremental">
                <Zap size="10px" />
              </span>
            {/if}
            {#if metadata?.partitioned}
              <span class="icon-indicator" title="Partitioned">
                <Layers size="10px" />
              </span>
            {/if}
            {#if metadata?.hasSchedule}
              <span
                class="icon-indicator"
                title={metadata?.scheduleDescription ?? "Scheduled"}
              >
                <Clock size="10px" />
              </span>
            {/if}
            <span
              class="check-indicator"
              class:checks-none={testCount === 0}
              class:checks-pass={testCount > 0 && !testHasErrors}
              class:checks-fail={testCount > 0 && testHasErrors}
              title={checkTooltip}
            >
              {#if testHasErrors}
                <AlertTriangle size="10px" />
              {:else}
                <CheckCircle size="10px" color="currentColor" />
              {/if}
              {testCount}
            </span>
          </span>
        </div>
      {:else if isMetricsView}
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {measuresCount} meas, {dimensionsCount} dims
          </span>
          <span
            class="lock-indicator"
            class:secured={hasSecurityRules}
            title={hasSecurityRules
              ? "Security policy defined"
              : "No security policy"}
          >
            {#if hasSecurityRules}
              <Lock size="10px" color="currentColor" />
            {:else}
              <Unlock size="10px" />
            {/if}
          </span>
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
          <span
            class="lock-indicator"
            class:secured={hasSecurityRules}
            title={hasSecurityRules
              ? "Security policy defined"
              : "No security policy"}
          >
            {#if hasSecurityRules}
              <Lock size="10px" color="currentColor" />
            {:else}
              <Unlock size="10px" />
            {/if}
          </span>
        </div>
      {:else if isCanvas}
        <div class="meta-row meta-row-spread">
          <span class="meta-detail">
            {componentCount} component{componentCount !== 1 ? "s" : ""}
          </span>
          <span
            class="lock-indicator"
            class:secured={hasSecurityRules}
            title={hasSecurityRules
              ? "Security policy defined"
              : "No security policy"}
          >
            {#if hasSecurityRules}
              <Lock size="10px" color="currentColor" />
            {:else}
              <Unlock size="10px" />
            {/if}
          </span>
        </div>
      {:else if isConnector}
        {#if connectorDriver}
          <div class="meta-row">
            <span class="badge">{connectorDriver}</span>
          </div>
        {/if}
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
    @apply relative border flex flex-col justify-between rounded-lg bg-surface-subtle px-2.5 py-2 cursor-pointer shadow-sm;
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

  .icon-group {
    @apply inline-flex items-center gap-x-1.5;
  }

  .icon-indicator {
    @apply inline-flex items-center text-fg-muted;
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
