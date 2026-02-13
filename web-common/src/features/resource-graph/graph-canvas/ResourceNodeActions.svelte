<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import {
    ResourceKind,
    displayResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceIconMapping,
    resourceShorthandMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    RefreshCw,
    RotateCcw,
    ExternalLink,
    GitFork,
    Info,
    Database,
    Clock,
    FileText,
    Table2,
    LayoutGrid,
    BarChart3,
    Palette,
    Bell,
    Plug,
    Zap,
    Layers,
    Component,
    AlertCircle,
  } from "lucide-svelte";
  import { createRuntimeServiceCreateTrigger } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { goto } from "$app/navigation";
  import type { ResourceNodeData } from "../shared/types";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import {
    detectConnectorFromPath,
    detectConnectorFromContent,
  } from "@rilldata/web-common/features/connectors/connector-type-detector";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let data: ResourceNodeData;

  let isOpen = false;
  let fullRefreshConfirmOpen = false;
  let describeOpen = false;

  $: ({ instanceId } = $runtime);
  $: resource = data?.resource;
  $: kind = data?.kind;
  $: resourceName = resource?.meta?.name?.name ?? "";
  $: filePath = resource?.meta?.filePaths?.[0];
  $: canRefresh =
    (kind === ResourceKind.Model || kind === ResourceKind.Source) &&
    !!resourceName;
  $: metadata = data?.metadata;

  // Derive connector info for Describe modal
  $: derivedConnector = (() => {
    // For Connector resources, use the driver directly
    if (metadata?.connectorDriver) return metadata.connectorDriver;
    const partitionsProps = resource?.model?.spec
      ?.partitionsResolverProperties as Record<string, unknown> | undefined;
    if (partitionsProps) {
      for (const value of Object.values(partitionsProps)) {
        if (typeof value === "string") {
          const detected = detectConnectorFromPath(value);
          if (detected) return detected;
        }
      }
    }
    const fromSourcePath = detectConnectorFromPath(metadata?.sourcePath);
    if (fromSourcePath) return fromSourcePath;
    const fromSqlQuery = detectConnectorFromContent(metadata?.sqlQuery);
    if (fromSqlQuery) return fromSqlQuery;
    if (metadata?.connector) return metadata.connector;
    return null;
  })();

  $: connectorIcon = (
    derivedConnector &&
    connectorIconMapping[
      derivedConnector.toLowerCase() as keyof typeof connectorIconMapping
    ]
      ? connectorIconMapping[
          derivedConnector.toLowerCase() as keyof typeof connectorIconMapping
        ]
      : null
  ) as ComponentType<SvelteComponent<{ size?: string }>> | null;

  $: icon =
    kind && resourceIconMapping[kind]
      ? resourceIconMapping[kind]
      : resourceIconMapping[ResourceKind.Model];
  $: color =
    kind && resourceShorthandMapping[kind]
      ? `var(--${resourceShorthandMapping[kind]})`
      : "#6B7280";

  $: templatedKeys = new Set(metadata?.connectorTemplatedProperties ?? []);
  $: hasError = !!resource?.meta?.reconcileError;
  $: reconcileStatus = resource?.meta?.reconcileStatus;
  $: isIdle = reconcileStatus === "RECONCILE_STATUS_IDLE";
  $: statusLabel =
    reconcileStatus && !isIdle
      ? reconcileStatus
          ?.replace("RECONCILE_STATUS_", "")
          ?.toLowerCase()
          ?.replaceAll("_", " ")
      : undefined;

  const triggerMutation = createRuntimeServiceCreateTrigger();

  function refreshModel(full: boolean) {
    if (!resourceName) return;
    $triggerMutation.mutate({
      instanceId,
      data: {
        models: [{ model: resourceName, full }],
      },
    });
  }

  function handleIncrementalRefresh() {
    isOpen = false;
    refreshModel(false);
  }

  function handleFullRefreshClick() {
    isOpen = false;
    fullRefreshConfirmOpen = true;
  }

  function confirmFullRefresh() {
    fullRefreshConfirmOpen = false;
    refreshModel(true);
  }

  function openFile() {
    if (!filePath) return;
    isOpen = false;
    goto(`/files${filePath}`);
    try {
      const prefs = JSON.parse(localStorage.getItem(filePath) || "{}");
      localStorage.setItem(
        filePath,
        JSON.stringify({ ...prefs, view: "code" }),
      );
    } catch (error) {
      console.warn(`Failed to save file view preference:`, error);
    }
  }

  function handleViewLineage() {
    if (!resource?.meta?.name) return;
    isOpen = false;
    const resourceKindName = resource.meta.name.kind;
    const resourceNameValue = resource.meta.name.name;
    const resourceId = encodeURIComponent(
      `${resourceKindName}:${resourceNameValue}`,
    );
    goto(`/graph?resource=${resourceId}&expanded=${resourceId}`);
  }

  function handleDescribe() {
    isOpen = false;
    describeOpen = true;
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="actions-root" on:click|stopPropagation on:mousedown|stopPropagation>
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isOpen} size={20}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleDescribe}
      >
        <div class="flex items-center gap-x-2">
          <Info size="12px" />
          <span>Describe</span>
        </div>
      </DropdownMenu.Item>
      {#if filePath}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={openFile}
        >
          <div class="flex items-center gap-x-2">
            <ExternalLink size="12px" />
            <span>Edit File</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleViewLineage}
      >
        <div class="flex items-center gap-x-2">
          <GitFork size="12px" />
          <span>View Lineage</span>
        </div>
      </DropdownMenu.Item>
      {#if canRefresh}
        <DropdownMenu.Separator />
        {#if kind === ResourceKind.Model}
          <DropdownMenu.Item
            class="font-normal flex items-center"
            on:click={handleIncrementalRefresh}
          >
            <div class="flex items-center gap-x-2">
              <RefreshCw size="12px" />
              <span>Incremental Refresh</span>
            </div>
          </DropdownMenu.Item>
        {/if}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleFullRefreshClick}
        >
          <div class="flex items-center gap-x-2">
            <RotateCcw size="12px" />
            <span>Full Refresh</span>
          </div>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>

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

<!-- Describe Modal -->
<Dialog.Root bind:open={describeOpen}>
  <Dialog.Content class="max-w-lg">
    <Dialog.Header>
      <Dialog.Title>
        <div class="describe-header">
          <div class="describe-icon" style={`background:${color}20`}>
            <svelte:component this={icon} size="20px" {color} />
          </div>
          <div class="describe-header-info">
            <span class="describe-name">{data?.label}</span>
            <span class="describe-kind">
              {#if kind}{displayResourceKind(kind)}{:else}Unknown{/if}
            </span>
          </div>
        </div>
      </Dialog.Title>
    </Dialog.Header>

    <div class="describe-body">
      <!-- Status -->
      {#if hasError || statusLabel}
        <div class="describe-section">
          <h4 class="describe-section-title">Status</h4>
          {#if hasError}
            <div class="describe-status-error">
              <AlertCircle size={14} />
              <span>Error</span>
            </div>
            <pre class="describe-error-msg">{resource?.meta?.reconcileError}</pre>
          {:else if statusLabel}
            <p class="describe-status">{statusLabel}</p>
          {/if}
        </div>
      {/if}

      <!-- Connector Resource Info -->
      {#if kind === ResourceKind.Connector && metadata?.connectorDriver}
        <div class="describe-section">
          <h4 class="describe-section-title">Connector Info</h4>
          <div class="describe-row">
            <span class="describe-row-icon">
              {#if connectorIcon}
                <svelte:component this={connectorIcon} size="16" />
              {:else}
                <Database size={16} />
              {/if}
            </span>
            <span>Driver: <strong>{metadata.connectorDriver}</strong></span>
          </div>
          {#if metadata?.connectorProvision}
            <div class="describe-row">
              <span class="describe-row-icon"><Layers size={14} /></span>
              <span>Rill Managed</span>
            </div>
          {/if}
          {#if metadata?.connectorProperties && Object.keys(metadata.connectorProperties).length > 0}
            <div class="describe-props">
              {#each Object.entries(metadata.connectorProperties) as [key, value]}
                <div class="describe-prop-row">
                  <span class="describe-prop-key">{key}</span>
                  {#if templatedKeys.has(key)}
                    <!-- svelte-ignore a11y-invalid-attribute -->
                    <a
                      href="#"
                      class="describe-env-anchor text-xs"
                      on:click|preventDefault={() => {
                        describeOpen = false;
                        goto("/files/.env");
                      }}
                    >edit</a>
                  {:else}
                    <span class="describe-prop-value" title={value}>{value}</span>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
          {#if metadata?.connectorTemplatedProperties?.length}
            <div class="describe-env-link">
              <ExternalLink size={14} />
              <span>Credentials stored in</span>
              <!-- svelte-ignore a11y-invalid-attribute -->
              <a
                href="#"
                class="describe-env-anchor"
                on:click|preventDefault={() => {
                  describeOpen = false;
                  goto("/files/.env");
                }}
              >.env</a>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Connector (for Model/Source) -->
      {#if kind !== ResourceKind.Connector && (derivedConnector || metadata?.sourcePath)}
        <div class="describe-section">
          <h4 class="describe-section-title">Connector</h4>
          <div class="describe-row">
            <span class="describe-row-icon">
              {#if connectorIcon}
                <svelte:component this={connectorIcon} size="16" />
              {:else}
                <Database size={16} />
              {/if}
            </span>
            <span>{derivedConnector}</span>
          </div>
          {#if metadata?.sourcePath}
            <div class="describe-row">
              <span class="describe-row-icon"><FileText size={14} /></span>
              <span class="describe-mono" title={metadata.sourcePath}>{metadata.sourcePath}</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Model/Source Info -->
      {#if kind === ResourceKind.Model || kind === ResourceKind.Source}
        <div class="describe-section">
          <div class="describe-section-row">
            <h4 class="describe-section-title">{displayResourceKind(kind)} Info</h4>
            <span class="describe-badge">{metadata?.isSqlModel ? "SQL" : "YAML"}</span>
          </div>
          {#if metadata?.incremental}
            <div class="describe-row">
              <span class="describe-row-icon"><RefreshCw size={14} /></span>
              <span>Incremental</span>
            </div>
          {/if}
          {#if metadata?.partitioned}
            <div class="describe-row">
              <span class="describe-row-icon"><Layers size={14} /></span>
              <span>Partitioned</span>
            </div>
          {/if}
          {#if metadata?.hasSchedule && metadata?.scheduleDescription}
            <div class="describe-row">
              <span class="describe-row-icon"><Clock size={14} /></span>
              <span>{metadata.scheduleDescription}</span>
            </div>
          {/if}
          {#if metadata?.testCount}
            <div class="describe-row">
              <span class="describe-row-icon"><Zap size={14} /></span>
              <span>{metadata.testCount} test{metadata.testCount > 1 ? "s" : ""}</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- MetricsView Info -->
      {#if kind === ResourceKind.MetricsView}
        <div class="describe-section">
          <h4 class="describe-section-title">MetricsView Info</h4>
          {#if metadata?.metricsModel}
            <div class="describe-row">
              <span class="describe-row-icon"><LayoutGrid size={14} /></span>
              <span>Model: {metadata.metricsModel}</span>
            </div>
          {/if}
          {#if metadata?.metricsTable}
            <div class="describe-row">
              <span class="describe-row-icon"><Table2 size={14} /></span>
              <span>Table: {metadata.metricsTable}</span>
            </div>
          {/if}
          {#if metadata?.timeDimension}
            <div class="describe-row">
              <span class="describe-row-icon"><Clock size={14} /></span>
              <span>Time: {metadata.timeDimension}</span>
            </div>
          {/if}
          {#if metadata?.dimensions?.length}
            <div class="describe-row">
              <span class="describe-row-icon"><BarChart3 size={14} /></span>
              <span>{metadata.dimensions.length} dimension{metadata.dimensions.length > 1 ? "s" : ""}</span>
            </div>
          {/if}
          {#if metadata?.measures?.length}
            <div class="describe-row">
              <span class="describe-row-icon"><BarChart3 size={14} /></span>
              <span>{metadata.measures.length} measure{metadata.measures.length > 1 ? "s" : ""}</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Explore Info -->
      {#if kind === ResourceKind.Explore}
        <div class="describe-section">
          <h4 class="describe-section-title">Explore Info</h4>
          {#if metadata?.metricsViewName}
            <div class="describe-row">
              <span class="describe-row-icon"><BarChart3 size={14} /></span>
              <span>{metadata.metricsViewName}</span>
            </div>
          {/if}
          {#if metadata?.theme}
            <div class="describe-row">
              <span class="describe-row-icon"><Palette size={14} /></span>
              <span>{metadata.theme}</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Canvas Info -->
      {#if kind === ResourceKind.Canvas}
        <div class="describe-section">
          <h4 class="describe-section-title">Canvas Info</h4>
          {#if metadata?.componentCount}
            <div class="describe-row">
              <span class="describe-row-icon"><Component size={14} /></span>
              <span>{metadata.componentCount} component{metadata.componentCount > 1 ? "s" : ""}</span>
            </div>
          {/if}
          {#if metadata?.rowCount}
            <div class="describe-row">
              <span class="describe-row-icon"><LayoutGrid size={14} /></span>
              <span>{metadata.rowCount} row{metadata.rowCount > 1 ? "s" : ""}</span>
            </div>
          {/if}
          {#if metadata?.theme}
            <div class="describe-row">
              <span class="describe-row-icon"><Palette size={14} /></span>
              <span>{metadata.theme}</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Consumers -->
      {#if metadata?.alertCount || metadata?.apiCount}
        <div class="describe-section">
          <h4 class="describe-section-title">Consumers</h4>
          {#if metadata?.alertCount}
            <div class="describe-row">
              <span class="describe-row-icon"><Bell size={14} /></span>
              <span>{metadata.alertCount} alert{metadata.alertCount > 1 ? "s" : ""}</span>
            </div>
          {/if}
          {#if metadata?.apiCount}
            <div class="describe-row">
              <span class="describe-row-icon"><Plug size={14} /></span>
              <span>{metadata.apiCount} API{metadata.apiCount > 1 ? "s" : ""}</span>
            </div>
          {/if}
        </div>
      {/if}

      <!-- SQL Query -->
      {#if metadata?.sqlQuery}
        <div class="describe-section">
          <h4 class="describe-section-title">SQL Query</h4>
          <pre class="describe-sql">{metadata.sqlQuery}</pre>
        </div>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .actions-root {
    @apply flex items-center;
  }

  /* Describe modal styles */
  .describe-header {
    @apply flex items-center gap-3;
  }

  .describe-icon {
    @apply flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg;
  }

  .describe-header-info {
    @apply flex flex-col;
  }

  .describe-name {
    @apply font-semibold text-sm text-fg-primary;
  }

  .describe-kind {
    @apply text-xs text-fg-secondary capitalize;
  }

  .describe-body {
    @apply flex flex-col max-h-[60vh] overflow-y-auto;
  }

  .describe-section {
    @apply py-3 border-b border-gray-200;
  }

  .describe-section:last-child {
    @apply border-b-0;
  }

  .describe-section-title {
    @apply text-xs font-semibold text-fg-secondary uppercase tracking-wide mb-2;
  }

  .describe-section-row {
    @apply flex items-center justify-between mb-2;
  }

  .describe-section-row .describe-section-title {
    @apply mb-0;
  }

  .describe-badge {
    @apply text-xs font-medium px-2 py-0.5 rounded bg-gray-100 text-gray-600;
  }

  .describe-row {
    @apply flex items-center gap-2 py-1 text-sm text-fg-primary;
  }

  .describe-row-icon {
    @apply flex items-center justify-center w-5 h-5 text-fg-muted;
  }

  .describe-mono {
    @apply font-mono text-xs truncate max-w-[300px];
  }

  .describe-status-error {
    @apply flex items-center gap-2 text-red-600 text-sm font-medium mb-2;
  }

  .describe-error-msg {
    @apply text-xs text-fg-secondary bg-surface-subtle p-2 rounded overflow-auto max-h-32 whitespace-pre-wrap;
  }

  .describe-status {
    @apply text-sm text-fg-primary capitalize;
  }

  .describe-sql {
    @apply text-xs font-mono bg-surface-subtle p-3 rounded overflow-auto whitespace-pre-wrap;
    @apply border border-gray-200;
    max-height: 200px;
  }

  .describe-props {
    @apply mt-2 bg-surface-subtle rounded border border-gray-200 overflow-hidden;
  }

  .describe-prop-row {
    @apply flex items-center justify-between px-3 py-1.5 text-xs;
    @apply border-b border-gray-100;
  }

  .describe-prop-row:last-child {
    @apply border-b-0;
  }

  .describe-prop-key {
    @apply font-mono text-fg-secondary;
  }

  .describe-prop-value {
    @apply font-mono text-fg-primary truncate max-w-[200px] text-right;
  }

  .describe-env-link {
    @apply flex items-center gap-2 mt-3 text-xs text-fg-muted;
  }

  .describe-env-anchor {
    @apply text-primary-600 font-medium underline;
  }
</style>
