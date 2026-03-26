<script lang="ts">
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import type { ComponentType, SvelteComponent } from "svelte";
  import {
    AlertTriangleIcon,
    ExternalLink,
    X,
    RefreshCw,
    Info,
    GitBranch,
    Copy,
    Check,
  } from "lucide-svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Button } from "@rilldata/web-common/components/button";
  import { inspectedNode, closeInspect } from "./inspect-store";
  import { goto } from "$app/navigation";
  import { getGraphNavigation } from "../shared/graph-navigation-context";
  import { tokenForKind } from "../navigation/seed-parser";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { TEST_FAILURE_MARKER } from "../shared/resource-status";
  import {
    createRuntimeServiceCreateTriggerMutation,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

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

  // Panel width: based on resource name length + padding for header icons
  const CHAR_WIDTH = 7.5;
  const PANEL_PADDING = 256; // badge + status icon + action buttons + padding
  const MIN_PANEL_WIDTH = 256;
  $: panelWidth = Math.max(
    MIN_PANEL_WIDTH,
    Math.round(resourceName.length * CHAR_WIDTH + PANEL_PADDING),
  );
  $: filePath = resource?.meta?.filePaths?.[0] ?? null;
  $: canOpenFile = !!filePath && (!!graphNav?.openFile || !graphNav);
  $: showNodeActions = !!graphNav?.openFile || !graphNav;

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

  // Action state
  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  let fullRefreshConfirmOpen = false;
  let specDialogOpen = false;

  $: canRefresh =
    (kind === ResourceKind.Source || kind === ResourceKind.Model) &&
    !!resourceName;

  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  function refreshModel(full: boolean) {
    if (!resourceName) return;
    $triggerMutation.mutate(
      {
        models: [{ model: resourceName, full }],
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries({
            queryKey: getRuntimeServiceListResourcesQueryKey(
              runtimeClient.instanceId,
              undefined,
            ),
          });
        },
        onError: (err) => {
          console.error(`Failed to refresh ${resourceName}:`, err);
          eventBus.emit("notification", {
            message: `Failed to refresh ${resourceName}`,
            type: "error",
          });
        },
      },
    );
  }

  function handleIncrementalRefresh() {
    refreshModel(false);
  }

  function handleFullRefreshClick() {
    fullRefreshConfirmOpen = true;
  }

  function confirmFullRefresh() {
    fullRefreshConfirmOpen = false;
    refreshModel(true);
  }

  let copiedError = false;
  let copiedTimeout: ReturnType<typeof setTimeout>;

  function handleCopyError() {
    navigator.clipboard
      .writeText(reconcileError)
      .then(() => {
        copiedError = true;
        clearTimeout(copiedTimeout);
        copiedTimeout = setTimeout(() => {
          copiedError = false;
        }, 2000);
      })
      .catch((err) => {
        console.error("Failed to copy error:", err);
      });
  }

  function viewNodeTree() {
    closeInspect();
    const kindToken = tokenForKind(kind);
    if (graphNav?.viewLineage) {
      graphNav.viewLineage(kindToken, resourceName);
      return;
    }
    const params = new URLSearchParams();
    if (kindToken) params.set("kind", kindToken);
    if (resourceName) params.set("resource", resourceName);
    goto(`/graph?${params.toString()}`);
  }

  function handleViewSpec() {
    specDialogOpen = true;
  }

  function getResourceSpec(res: V1Resource | undefined): string {
    if (!res) return "";
    const kindKeys = [
      "source",
      "model",
      "metricsView",
      "explore",
      "theme",
      "component",
      "canvas",
      "api",
      "connector",
      "report",
      "alert",
    ] as const;
    for (const key of kindKeys) {
      if (res[key]) {
        return JSON.stringify(res[key], null, 2);
      }
    }
    const rest = Object.fromEntries(
      Object.entries(res).filter(([k]) => k !== "meta"),
    );
    return JSON.stringify(rest, null, 2);
  }

  $: specContent = getResourceSpec(resource);

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
  <div
    class="inspect-panel"
    style="top: {panelTop}px; left: {panelLeft}px; width: {panelWidth}px;"
  >
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
      <div class="flex items-center gap-x-1">
        {#if canRefresh}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="shrink-0 p-0.5 rounded text-fg-muted hover:text-fg-primary hover:bg-surface-hover"
              title="Refresh"
            >
              <RefreshCw size="14px" />
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Item
                class="font-normal flex items-center gap-x-2"
                onclick={handleFullRefreshClick}
              >
                <RefreshCw size="12px" />
                <span>Full Refresh</span>
              </DropdownMenu.Item>
              {#if isIncremental}
                <DropdownMenu.Item
                  class="font-normal flex items-center gap-x-2"
                  onclick={handleIncrementalRefresh}
                >
                  <RefreshCw size="12px" />
                  <span>Incremental Refresh</span>
                </DropdownMenu.Item>
              {/if}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}
        <button
          class="header-action-btn"
          onclick={viewNodeTree}
          title="View Lineage"
        >
          <GitBranch size="14px" />
        </button>
        <button
          class="header-action-btn"
          onclick={handleViewSpec}
          title="Describe"
        >
          <Info size="14px" />
        </button>
        <button class="close-btn" onclick={closeInspect} aria-label="Close">
          <X size="14px" />
        </button>
      </div>
    </div>

    <!-- Body -->
    <div class="panel-body">
      {#if hasError}
        <div class="error-banner">
          <CancelCircle size="14px" className="text-destructive flex-none" />
          <pre class="error-message">{reconcileError}</pre>
          <button
            class="copy-error-btn"
            onclick={handleCopyError}
            title="Copy error"
          >
            {#if copiedError}
              <Check size="12px" class="text-green-500" />
            {:else}
              <Copy size="12px" />
            {/if}
          </button>
        </div>
      {:else if isTestOnlyError}
        <div class="warning-banner">
          <CancelCircle size="14px" className="text-destructive flex-none" />
          <pre class="error-message">{reconcileError}</pre>
          <button
            class="copy-error-btn"
            onclick={handleCopyError}
            title="Copy error"
          >
            {#if copiedError}
              <Check size="12px" class="text-green-500" />
            {:else}
              <Copy size="12px" />
            {/if}
          </button>
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
            <span class="detail-value">Yes</span>
          {/if}
          {#if isPartitioned}
            <span class="detail-label">Partitioned</span>
            <span class="detail-value">Yes</span>
          {/if}
          {#if hasSchedule}
            <span class="detail-label">Schedule</span>
            <span class="detail-value"
              >{metadata?.scheduleDescription ?? "Enabled"}</span
            >
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
              class="detail-value"
              class:text-green-600={!testHasErrors}
              class:text-amber-600={testHasErrors}
            >
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
            <span class="detail-value text-amber-600">Policy defined</span>
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
            <span class="detail-value text-amber-600">Policy defined</span>
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
            <span class="detail-value text-amber-600">Policy defined</span>
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
    </div>

    <!-- File link (dev only) -->
    {#if showNodeActions && canOpenFile && filePath}
      <div class="panel-actions">
        <button class="file-link" onclick={navigateToFile}>
          <ExternalLink size="12px" />
          <span>{filePath}</span>
        </button>
      </div>
    {/if}
  </div>
{/if}

<AlertDialog.Root bind:open={fullRefreshConfirmOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Full Refresh {resourceName}?</AlertDialog.Title>
      <AlertDialog.Description>
        <div class="mt-1">
          A full refresh will re-ingest ALL data from scratch. This operation
          can take a significant amount of time and will update all dependent
          resources. Only proceed if you're certain this is necessary.
        </div>
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button
        type="secondary"
        onClick={() => {
          fullRefreshConfirmOpen = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={confirmFullRefresh}>Yes, refresh</Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<Dialog.Root bind:open={specDialogOpen}>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>
        {resourceName}
        <span class="text-fg-tertiary font-normal text-sm ml-2"
          >{kind ? displayResourceKind(kind) : ""}</span
        >
      </Dialog.Title>
    </Dialog.Header>
    <div class="spec-container">
      {#if !resource}
        <p class="text-sm text-fg-secondary">No resource data available</p>
      {:else}
        <pre class="spec-content">{specContent}</pre>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .inspect-panel {
    @apply absolute z-30 rounded-lg border bg-surface-base shadow-lg;
    max-height: 320px;
    display: flex;
    flex-direction: column;
  }

  .panel-header {
    @apply flex items-center justify-between gap-x-2 px-3 py-2 border-b;
  }

  .header-action-btn {
    @apply shrink-0 p-0.5 rounded text-fg-muted;
  }

  .header-action-btn:hover {
    @apply text-fg-primary bg-surface-hover;
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
    @apply text-xs font-mono whitespace-pre-wrap max-h-[80px] overflow-auto flex-1 min-w-0;
  }

  .copy-error-btn {
    @apply shrink-0 p-0.5 rounded text-fg-muted self-start;
  }

  .copy-error-btn:hover {
    @apply text-fg-primary bg-surface-hover;
  }

  .file-link {
    @apply flex items-center gap-x-1.5 text-xs text-fg-secondary underline;
  }

  .file-link:hover {
    @apply text-fg-primary;
  }

  .panel-actions {
    @apply flex flex-wrap gap-1 px-3 py-2 border-t;
  }

  .spec-container {
    @apply overflow-auto flex-1 min-h-0;
  }

  .spec-content {
    @apply text-xs font-mono whitespace-pre-wrap bg-surface-subtle rounded-md p-4;
  }
</style>
