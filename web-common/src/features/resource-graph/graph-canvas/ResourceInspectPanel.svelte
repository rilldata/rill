<script lang="ts">
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import type { ComponentType, SvelteComponent } from "svelte";
  import {
    AlertTriangleIcon,
    Zap,
    Layers,
    Clock,
    Lock,
    ExternalLink,
    X,
  } from "lucide-svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { inspectedNode, closeInspect } from "./inspect-store";
  import { goto } from "$app/navigation";
  import { getGraphNavigation } from "../shared/graph-navigation-context";
  import { TEST_FAILURE_MARKER } from "../shared/resource-status";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";

  const graphNav = getGraphNavigation();

  $: state = $inspectedNode;
  $: data = state?.data ?? null;
  $: kind = data?.kind;
  $: resource = data?.resource;
  $: metadata = data?.metadata;
  $: resourceName = resource?.meta?.name?.name ?? "";

  // Position below the clicked node
  $: panelTop = state ? state.y + state.height + 8 : 0;
  $: panelLeft = state ? state.x : 0;
  $: filePath = resource?.meta?.filePaths?.[0] ?? null;
  $: canOpenFile = !!filePath && (!!graphNav?.openFile || !graphNav);

  // Status
  $: reconcileError = resource?.meta?.reconcileError ?? "";
  $: isTestOnlyError =
    !!reconcileError && reconcileError.includes(TEST_FAILURE_MARKER);
  $: hasError = !!reconcileError && !isTestOnlyError;
  $: isPending =
    resource?.meta?.reconcileStatus &&
    resource.meta.reconcileStatus !== V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  // Model/Source metadata
  $: isSourceOrModel =
    kind === ResourceKind.Source || kind === ResourceKind.Model;
  $: connector = metadata?.connector ?? null;
  $: connectorIcon = (
    connector &&
    connectorIconMapping[connector as keyof typeof connectorIconMapping]
      ? connectorIconMapping[connector as keyof typeof connectorIconMapping]
      : null
  ) as ComponentType<SvelteComponent<{ size?: string }>> | null;
  $: isMaterialized = metadata?.isMaterialized === true;
  $: isIncremental = metadata?.incremental === true;
  $: isPartitioned = metadata?.partitioned === true;
  $: hasSchedule = metadata?.hasSchedule === true;
  $: testCount = metadata?.testCount ?? 0;
  $: testHasErrors = (metadata?.testErrors?.length ?? 0) > 0;
  $: hasSecurityRules = metadata?.hasSecurityRules === true;

  // MetricsView
  $: measuresCount = metadata?.measures?.length ?? 0;
  $: dimensionsCount = metadata?.dimensions?.length ?? 0;

  // Explore
  $: exploreMeasures = metadata?.exploreMeasuresAll
    ? "all"
    : String(metadata?.exploreMeasuresCount ?? 0);
  $: exploreDimensions = metadata?.exploreDimensionsAll
    ? "all"
    : String(metadata?.exploreDimensionsCount ?? 0);

  // Canvas
  $: componentCount = metadata?.componentCount ?? 0;

  // Connector
  $: connectorDriver = metadata?.connectorDriver ?? null;
  $: driverIcon = (
    connectorDriver &&
    connectorIconMapping[connectorDriver as keyof typeof connectorIconMapping]
      ? connectorIconMapping[
          connectorDriver as keyof typeof connectorIconMapping
        ]
      : null
  ) as ComponentType<SvelteComponent<{ size?: string }>> | null;

  function formatDate(value: string | undefined): string {
    if (!value) return "-";
    return new Date(value).toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }

  function formatDuration(ms: string | undefined): string {
    if (!ms) return "-";
    const num = Number(ms);
    if (num < 1000) return `${num}ms`;
    if (num < 60000) return `${(num / 1000).toFixed(1)}s`;
    return `${(num / 60000).toFixed(1)}m`;
  }

  function navigateToFile() {
    if (!filePath) return;
    closeInspect();
    if (graphNav?.openFile) {
      graphNav.openFile(filePath);
      return;
    }
    try {
      const prefs = JSON.parse(localStorage.getItem(filePath) || "{}");
      localStorage.setItem(
        filePath,
        JSON.stringify({ ...prefs, view: "code" }),
      );
    } catch {
      // ignore
    }
    goto(`/files${filePath}`);
  }
</script>

{#if data}
  <div class="inspect-panel" style="top: {panelTop}px; left: {panelLeft}px;">
    <!-- Header -->
    <div class="panel-header">
      <div class="flex items-center gap-x-2">
        {#if kind}<ResourceTypeBadge {kind} />{/if}
        <span class="text-sm font-medium">{resourceName}</span>
        {#if isPending}
          <LoadingSpinner size="14px" />
        {:else if hasError}
          <CancelCircle size="14px" className="text-red-500 flex-none" />
        {:else if isTestOnlyError}
          <AlertTriangleIcon size="14px" class="text-yellow-500 flex-none" />
        {/if}
      </div>
      <button class="close-btn" onclick={closeInspect} aria-label="Close">
        <X size="14px" />
      </button>
    </div>

    <!-- Body -->
    <div class="panel-body">
      {#if hasError}
        <div class="error-banner">
          <CancelCircle size="14px" className="text-destructive flex-none" />
          <pre class="error-message">{reconcileError}</pre>
        </div>
      {:else if isTestOnlyError}
        <div class="warning-banner">
          <AlertTriangleIcon size="14px" class="text-yellow-500 flex-none" />
          <pre class="error-message">{reconcileError}</pre>
        </div>
      {/if}

      <!-- Source / Model details -->
      {#if isSourceOrModel}
        <div class="detail-grid">
          {#if connector}
            <span class="detail-label">Connector</span>
            <span class="detail-value flex items-center gap-x-1.5">
              {#if connectorIcon}
                <svelte:component this={connectorIcon} size="14px" />
              {/if}
              {connector}
            </span>
          {/if}
          <span class="detail-label">Type</span>
          <span class="detail-value">{isMaterialized ? "Table" : "View"}</span>
          {#if isIncremental}
            <span class="detail-label">Incremental</span>
            <span class="detail-value flex items-center gap-x-1">
              <Zap size="12px" class="text-fg-muted" /> Yes
            </span>
          {/if}
          {#if isPartitioned}
            <span class="detail-label">Partitioned</span>
            <span class="detail-value flex items-center gap-x-1">
              <Layers size="12px" class="text-fg-muted" /> Yes
            </span>
          {/if}
          {#if hasSchedule}
            <span class="detail-label">Schedule</span>
            <span class="detail-value flex items-center gap-x-1">
              <Clock size="12px" class="text-fg-muted" />
              {metadata?.scheduleDescription ?? "Enabled"}
            </span>
          {/if}
          {#if metadata?.lastRefreshedOn}
            <span class="detail-label">Last refreshed</span>
            <span class="detail-value"
              >{formatDate(metadata.lastRefreshedOn)}</span
            >
          {/if}
          {#if metadata?.executionDurationMs}
            <span class="detail-label">Duration</span>
            <span class="detail-value"
              >{formatDuration(metadata.executionDurationMs)}</span
            >
          {/if}
          {#if testCount > 0}
            <span class="detail-label">Checks</span>
            <span
              class="detail-value flex items-center gap-x-1"
              class:text-green-600={!testHasErrors}
              class:text-amber-600={testHasErrors}
            >
              {#if testHasErrors}
                <AlertTriangleIcon size="12px" />
              {:else}
                <CheckCircle size="12px" color="currentColor" />
              {/if}
              {testCount} check{testCount > 1 ? "s" : ""}
              {testHasErrors ? "failed" : "passed"}
            </span>
          {/if}
        </div>
      {/if}

      <!-- MetricsView details -->
      {#if kind === ResourceKind.MetricsView}
        <div class="detail-grid">
          {#if metadata?.metricsModel}
            <span class="detail-label">Model</span>
            <span class="detail-value">{metadata.metricsModel}</span>
          {/if}
          {#if metadata?.timeDimension}
            <span class="detail-label">Time dim</span>
            <span class="detail-value">{metadata.timeDimension}</span>
          {/if}
          <span class="detail-label">Measures</span>
          <span class="detail-value">{measuresCount}</span>
          <span class="detail-label">Dimensions</span>
          <span class="detail-value">{dimensionsCount}</span>
          {#if hasSecurityRules}
            <span class="detail-label">Security</span>
            <span class="detail-value flex items-center gap-x-1 text-amber-600">
              <Lock size="12px" /> Policy defined
            </span>
          {/if}
        </div>
      {/if}

      <!-- Explore details -->
      {#if kind === ResourceKind.Explore}
        <div class="detail-grid">
          {#if metadata?.metricsViewName}
            <span class="detail-label">Metrics view</span>
            <span class="detail-value">{metadata.metricsViewName}</span>
          {/if}
          <span class="detail-label">Measures</span>
          <span class="detail-value">{exploreMeasures}</span>
          <span class="detail-label">Dimensions</span>
          <span class="detail-value">{exploreDimensions}</span>
          {#if hasSecurityRules}
            <span class="detail-label">Security</span>
            <span class="detail-value flex items-center gap-x-1 text-amber-600">
              <Lock size="12px" /> Policy defined
            </span>
          {/if}
        </div>
      {/if}

      <!-- Canvas details -->
      {#if kind === ResourceKind.Canvas}
        <div class="detail-grid">
          <span class="detail-label">Components</span>
          <span class="detail-value">{componentCount}</span>
          {#if metadata?.rowCount}
            <span class="detail-label">Rows</span>
            <span class="detail-value">{metadata.rowCount}</span>
          {/if}
          {#if hasSecurityRules}
            <span class="detail-label">Security</span>
            <span class="detail-value flex items-center gap-x-1 text-amber-600">
              <Lock size="12px" /> Policy defined
            </span>
          {/if}
        </div>
      {/if}

      <!-- Connector details -->
      {#if kind === ResourceKind.Connector}
        <div class="detail-grid">
          {#if connectorDriver}
            <span class="detail-label">Driver</span>
            <span class="detail-value flex items-center gap-x-1.5">
              {#if driverIcon}
                <svelte:component this={driverIcon} size="14px" />
              {/if}
              {connectorDriver}
            </span>
          {/if}
        </div>
      {/if}

      <!-- File path link -->
      {#if canOpenFile && filePath}
        <button class="file-link" onclick={navigateToFile}>
          <ExternalLink size="12px" />
          <span>{filePath}</span>
        </button>
      {/if}
    </div>
  </div>
{/if}

<style lang="postcss">
  .inspect-panel {
    @apply absolute z-30 rounded-lg border bg-surface-base shadow-lg;
    min-width: 16rem;
    max-height: 320px;
    display: flex;
    flex-direction: column;
  }

  .panel-header {
    @apply flex items-center justify-between gap-x-2 px-3 py-2 border-b;
  }

  .close-btn {
    @apply shrink-0 p-0.5 rounded text-fg-muted;
  }

  .close-btn:hover {
    @apply text-fg-primary bg-surface-hover;
  }

  .panel-body {
    @apply flex flex-col gap-y-3 px-3 py-2.5 overflow-y-auto overflow-x-hidden text-sm;
  }

  .detail-grid {
    @apply grid grid-cols-[auto_1fr] gap-x-3 gap-y-1;
  }

  .detail-label {
    @apply text-xs text-fg-tertiary whitespace-nowrap;
  }

  .detail-value {
    @apply text-xs text-fg-primary;
  }

  .error-banner {
    @apply flex gap-x-2 items-start border border-destructive bg-destructive/15 text-fg-primary border-l-4 px-2 py-2 rounded;
  }

  .warning-banner {
    @apply flex gap-x-2 items-start border border-yellow-400 bg-yellow-500/15 text-fg-primary border-l-4 px-2 py-2 rounded;
  }

  .error-message {
    @apply text-xs font-mono whitespace-pre-wrap max-h-[80px] overflow-auto;
  }

  .file-link {
    @apply flex items-center gap-x-1.5 text-xs text-fg-secondary underline;
  }

  .file-link:hover {
    @apply text-fg-primary;
  }
</style>
