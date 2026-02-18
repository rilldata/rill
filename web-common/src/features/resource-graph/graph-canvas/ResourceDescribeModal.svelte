<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import {
    ResourceKind,
    displayResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceIconMapping,
    resourceShorthandMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
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
    ChevronDown,
    ChevronRight,
    RotateCcw,
    HardDrive,
    ArrowRightLeft,
    Timer,
    RefreshCw,
  } from "lucide-svelte";
  import type { ResourceNodeData } from "../shared/types";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import {
    detectConnectorFromPath,
    detectConnectorFromContent,
  } from "@rilldata/web-common/features/connectors/connector-type-detector";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import { goto } from "$app/navigation";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let open = false;
  export let data: ResourceNodeData;

  $: resource = data?.resource;
  $: kind = data?.kind;
  $: metadata = data?.metadata;

  // Derive connector info
  $: derivedConnector = (() => {
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

  // Collapsible state for measures/dimensions
  let showMeasures = false;
  let showDimensions = false;

  // Navigate to model file to view partitions
  $: filePath = resource?.meta?.filePaths?.[0] ?? null;
  function openPartitions() {
    if (!filePath) return;
    open = false;
    goto(`/files${filePath}?partitions=open`);
  }

  // Security rules from resource spec
  $: securityRules = (() => {
    if (kind === ResourceKind.MetricsView)
      return resource?.metricsView?.spec?.securityRules ?? [];
    if (kind === ResourceKind.Explore)
      return resource?.explore?.spec?.securityRules ?? [];
    if (kind === ResourceKind.Canvas)
      return resource?.canvas?.spec?.securityRules ?? [];
    return [];
  })();

  // Explore measures/dimensions from spec
  $: exploreSpec = resource?.explore?.spec;
  $: exploreMeasures = exploreSpec?.measures ?? [];
  $: exploreDimensions = exploreSpec?.dimensions ?? [];
  $: exploreMeasuresAll = exploreSpec?.measuresSelector?.all === true;
  $: exploreDimensionsAll = exploreSpec?.dimensionsSelector?.all === true;

  // Description from type-specific specs
  $: description = (() => {
    if (kind === ResourceKind.MetricsView)
      return resource?.metricsView?.spec?.description;
    if (kind === ResourceKind.Explore)
      return resource?.explore?.spec?.description;
    if (kind === ResourceKind.Canvas)
      return resource?.canvas?.spec?.displayName;
    return undefined;
  })();

  // MetricsView-specific fields
  $: mvSpec = resource?.metricsView?.spec;
  $: firstDayOfWeek = mvSpec?.firstDayOfWeek;
  $: firstMonthOfYear = mvSpec?.firstMonthOfYear;
  $: aiInstructions = mvSpec?.aiInstructions;

  // Day names for firstDayOfWeek
  const dayNames = [
    "",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
    "Sunday",
  ];
  // Month names for firstMonthOfYear
  const monthNames = [
    "",
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];

  // Format execution duration
  function formatDuration(ms: string | undefined): string | null {
    if (!ms) return null;
    const num = parseInt(ms, 10);
    if (isNaN(num)) return null;
    if (num < 1000) return `${num}ms`;
    if (num < 60_000) return `${(num / 1000).toFixed(1)}s`;
    return `${Math.floor(num / 60_000)}m ${Math.round((num % 60_000) / 1000)}s`;
  }

  $: executionDuration = formatDuration(metadata?.executionDurationMs);

  // Rebuild security rules as YAML-like text (matches Rill YAML format)
  $: securityYaml = (() => {
    if (!securityRules.length) return "";
    const lines: string[] = ["security:"];
    for (const rule of securityRules) {
      if (rule.access) {
        const val = rule.access.conditionExpression
          ? `"${rule.access.conditionExpression}"`
          : rule.access.allow
            ? "true"
            : "false";
        lines.push(`  access: ${val}`);
      }
      if (rule.fieldAccess) {
        lines.push(`  field_access:`);
        if (rule.fieldAccess.conditionExpression) {
          lines.push(`    if: "${rule.fieldAccess.conditionExpression}"`);
        }
        if (rule.fieldAccess.allFields) {
          lines.push(`    all: true`);
        } else if (rule.fieldAccess.fields?.length) {
          lines.push(
            `    fields: [${rule.fieldAccess.fields.join(", ")}]`,
          );
        }
        lines.push(
          `    allow: ${rule.fieldAccess.allow ? "true" : "false"}`,
        );
      }
      if (rule.rowFilter) {
        const val = rule.rowFilter.sql
          ? `"${rule.rowFilter.sql}"`
          : "true";
        lines.push(`  row_filter: ${val}`);
        if (rule.rowFilter.conditionExpression) {
          lines.push(`    if: "${rule.rowFilter.conditionExpression}"`);
        }
      }
    }
    return lines.join("\n");
  })();
</script>

<Dialog.Root bind:open>
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
      <!-- Description -->
      {#if description}
        <div class="describe-section">
          <p class="describe-description">{description}</p>
        </div>
      {/if}

      <!-- Status -->
      {#if hasError || statusLabel}
        <div class="describe-section">
          <h4 class="describe-section-title">Status</h4>
          {#if hasError}
            <div class="describe-status-error">
              <AlertCircle size={14} />
              <span>Error</span>
            </div>
            <pre class="describe-error-msg">{resource?.meta
                ?.reconcileError}</pre>
          {:else if statusLabel}
            <p class="describe-status">{statusLabel}</p>
          {/if}
        </div>
      {/if}

      <!-- Connector (for Model/Source) -->
      {#if derivedConnector || metadata?.sourcePath}
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
              <span class="describe-mono" title={metadata.sourcePath}
                >{metadata.sourcePath}</span
              >
            </div>
          {/if}
        </div>
      {/if}

      <!-- Model/Source Info -->
      {#if kind === ResourceKind.Model || kind === ResourceKind.Source}
        <div class="describe-section">
          <div class="describe-section-row">
            <h4 class="describe-section-title">
              {displayResourceKind(kind)} Info
            </h4>
            <span class="describe-badge"
              >{metadata?.isSqlModel ? "SQL" : "YAML"}</span
            >
          </div>
          {#if metadata?.inputConnector}
            <div class="describe-row">
              <span class="describe-row-icon"><Database size={14} /></span>
              <span>Input: {metadata.inputConnector}</span>
            </div>
          {/if}
          {#if metadata?.outputConnector}
            <div class="describe-row">
              <span class="describe-row-icon"><HardDrive size={14} /></span>
              <span>Output: {metadata.outputConnector}{metadata.stageConnector ? ` (stage: ${metadata.stageConnector})` : ""}</span>
            </div>
          {/if}
          {#if metadata?.resultTable}
            <div class="describe-row">
              <span class="describe-row-icon"><Table2 size={14} /></span>
              <span>Table: {metadata.resultTable}</span>
            </div>
          {/if}
          {#if metadata?.materialize}
            <div class="describe-row">
              <span class="describe-row-icon"><HardDrive size={14} /></span>
              <span>Materialized</span>
            </div>
          {/if}
          {#if metadata?.partitioned}
            <button class="describe-row describe-row-link" on:click={openPartitions}>
              <span class="describe-row-icon"><Layers size={14} /></span>
              <span>Partitioned</span>
              <ChevronRight size={12} />
            </button>
          {/if}
          {#if metadata?.incremental}
            <div class="describe-row">
              <span class="describe-row-icon"><Zap size={14} /></span>
              <span>Incremental</span>
            </div>
          {/if}
          {#if metadata?.changeMode}
            <div class="describe-row">
              <span class="describe-row-icon"><ArrowRightLeft size={14} /></span>
              <span>Change mode: {metadata.changeMode.replace("MODEL_CHANGE_MODE_", "").toLowerCase()}</span>
            </div>
          {/if}
          {#if metadata?.testCount}
            <div class="describe-row">
              <span class="describe-row-icon"><CheckCircle size="14px" color="currentColor" /></span>
              <span
                >{metadata.testCount} test{metadata.testCount > 1
                  ? "s"
                  : ""}</span
              >
            </div>
          {/if}
        </div>

        <!-- Refresh -->
        <div class="describe-section">
          <h4 class="describe-section-title">Refresh</h4>
          {#if metadata?.lastRefreshedOn}
            <div class="describe-row">
              <span class="describe-row-icon"><Clock size={14} /></span>
              <span>Last: {new Date(metadata.lastRefreshedOn).toLocaleString()}</span>
            </div>
          {/if}
          {#if executionDuration}
            <div class="describe-row">
              <span class="describe-row-icon"><Timer size={14} /></span>
              <span>Duration: {executionDuration}</span>
            </div>
          {/if}
          {#if metadata?.hasSchedule && metadata?.scheduleDescription}
            <div class="describe-row">
              <span class="describe-row-icon"><RefreshCw size={14} /></span>
              <span>Cron: {metadata.scheduleDescription}</span>
            </div>
          {/if}
          {#if metadata?.refUpdate}
            <div class="describe-row">
              <span class="describe-row-icon"><RefreshCw size={14} /></span>
              <span>Refresh on upstream change</span>
            </div>
          {/if}
          {#if metadata?.timeoutSeconds}
            <div class="describe-row">
              <span class="describe-row-icon"><Timer size={14} /></span>
              <span>Timeout: {metadata.timeoutSeconds}s</span>
            </div>
          {/if}
          {#if metadata?.retryAttempts}
            <div class="describe-row">
              <span class="describe-row-icon"><RotateCcw size={14} /></span>
              <span>
                Retry: {metadata.retryAttempts}x{metadata.retryDelaySeconds ? `, ${metadata.retryDelaySeconds}s delay` : ""}{metadata.retryExponentialBackoff ? ", exponential backoff" : ""}
              </span>
            </div>
          {/if}
          {#if !metadata?.lastRefreshedOn && !metadata?.hasSchedule && !metadata?.refUpdate && !metadata?.retryAttempts}
            <div class="describe-row">
              <span class="describe-row-icon"><Clock size={14} /></span>
              <span class="text-fg-muted">No refresh data</span>
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
          {#if firstDayOfWeek && firstDayOfWeek !== 1}
            <div class="describe-row">
              <span class="describe-row-icon"><Clock size={14} /></span>
              <span
                >First day of week: {dayNames[firstDayOfWeek] ??
                  firstDayOfWeek}</span
              >
            </div>
          {/if}
          {#if firstMonthOfYear && firstMonthOfYear !== 1}
            <div class="describe-row">
              <span class="describe-row-icon"><Clock size={14} /></span>
              <span
                >First month of year: {monthNames[firstMonthOfYear] ??
                  firstMonthOfYear}</span
              >
            </div>
          {/if}
          {#if aiInstructions}
            <div class="describe-row">
              <span class="describe-row-icon"><Zap size={14} /></span>
              <span>AI Instructions</span>
            </div>
            <pre class="describe-ai-instructions">{aiInstructions}</pre>
          {/if}
        </div>

        <!-- Measures dropdown -->
        {#if metadata?.measures?.length}
          <div class="describe-section">
            <button
              class="describe-collapse-toggle"
              on:click={() => (showMeasures = !showMeasures)}
            >
              {#if showMeasures}
                <ChevronDown size={14} />
              {:else}
                <ChevronRight size={14} />
              {/if}
              <span class="describe-section-title" style="margin-bottom:0"
                >Measures ({metadata.measures.length})</span
              >
            </button>
            {#if showMeasures}
              <div class="describe-list">
                {#each metadata.measures as measure}
                  <div class="describe-list-item">
                    <span class="describe-list-name"
                      >{measure.displayName || measure.name}</span
                    >
                    {#if measure.expression}
                      <span class="describe-list-detail"
                        >{measure.expression}</span
                      >
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        <!-- Dimensions dropdown -->
        {#if metadata?.dimensions?.length}
          <div class="describe-section">
            <button
              class="describe-collapse-toggle"
              on:click={() => (showDimensions = !showDimensions)}
            >
              {#if showDimensions}
                <ChevronDown size={14} />
              {:else}
                <ChevronRight size={14} />
              {/if}
              <span class="describe-section-title" style="margin-bottom:0"
                >Dimensions ({metadata.dimensions.length})</span
              >
            </button>
            {#if showDimensions}
              <div class="describe-list">
                {#each metadata.dimensions as dim}
                  <div class="describe-list-item">
                    <span class="describe-list-name"
                      >{dim.displayName || dim.name}</span
                    >
                    {#if dim.column && dim.column !== dim.name}
                      <span class="describe-list-detail">{dim.column}</span>
                    {:else if dim.expression}
                      <span class="describe-list-detail">{dim.expression}</span>
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        <!-- MetricsView Security -->
        {#if securityYaml}
          <div class="describe-section">
            <h4 class="describe-section-title">Security Policy</h4>
            <pre class="describe-yaml">{securityYaml}</pre>
          </div>
        {/if}
      {/if}

      <!-- Explore Info -->
      {#if kind === ResourceKind.Explore}
        <div class="describe-section">
          <h4 class="describe-section-title">Explore Info</h4>
          {#if metadata?.metricsViewName}
            <div class="describe-row">
              <span class="describe-row-icon"><BarChart3 size={14} /></span>
              <span>MetricsView: {metadata.metricsViewName}</span>
            </div>
          {/if}
          {#if metadata?.theme}
            <div class="describe-row">
              <span class="describe-row-icon"><Palette size={14} /></span>
              <span>Theme: {metadata.theme}</span>
            </div>
          {/if}
          <div class="describe-row">
            <span class="describe-row-icon"><BarChart3 size={14} /></span>
            <span>
              Measures:
              {#if exploreMeasuresAll}
                all
              {:else if exploreMeasures.length}
                {exploreMeasures.join(", ")}
              {:else}
                {metadata?.exploreMeasuresCount ?? 0}
              {/if}
            </span>
          </div>
          <div class="describe-row">
            <span class="describe-row-icon"><BarChart3 size={14} /></span>
            <span>
              Dimensions:
              {#if exploreDimensionsAll}
                all
              {:else if exploreDimensions.length}
                {exploreDimensions.join(", ")}
              {:else}
                {metadata?.exploreDimensionsCount ?? 0}
              {/if}
            </span>
          </div>
        </div>

        <!-- Explore Security -->
        {#if securityYaml}
          <div class="describe-section">
            <h4 class="describe-section-title">Security Policy</h4>
            <pre class="describe-yaml">{securityYaml}</pre>
          </div>
        {/if}
      {/if}

      <!-- Canvas Info -->
      {#if kind === ResourceKind.Canvas}
        <div class="describe-section">
          <h4 class="describe-section-title">Canvas Info</h4>
          {#if metadata?.componentCount}
            <div class="describe-row">
              <span class="describe-row-icon"><Component size={14} /></span>
              <span
                >{metadata.componentCount} component{metadata.componentCount > 1
                  ? "s"
                  : ""}</span
              >
            </div>
          {/if}
          {#if metadata?.rowCount}
            <div class="describe-row">
              <span class="describe-row-icon"><LayoutGrid size={14} /></span>
              <span
                >{metadata.rowCount} row{metadata.rowCount > 1 ? "s" : ""}</span
              >
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
              <span
                >{metadata.alertCount} alert{metadata.alertCount > 1
                  ? "s"
                  : ""}</span
              >
            </div>
          {/if}
          {#if metadata?.apiCount}
            <div class="describe-row">
              <span class="describe-row-icon"><Plug size={14} /></span>
              <span
                >{metadata.apiCount} API{metadata.apiCount > 1 ? "s" : ""}</span
              >
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
    @apply py-3 border-b;
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
    @apply text-xs font-medium px-2 py-0.5 rounded bg-surface-subtle text-fg-secondary;
  }

  .describe-row {
    @apply flex items-center gap-2 py-1 text-sm text-fg-primary;
  }

  button.describe-row-link {
    @apply cursor-pointer rounded px-1 -mx-1 bg-transparent border-none;
  }

  button.describe-row-link:hover {
    @apply bg-surface-hover text-primary-500;
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
    @apply border;
    max-height: 200px;
  }

  .describe-collapse-toggle {
    @apply flex items-center gap-1.5 w-full text-left py-0.5 cursor-pointer text-fg-secondary;
    background: none;
    border: none;
  }

  .describe-collapse-toggle:hover {
    @apply text-fg-primary;
  }

  .describe-list {
    @apply flex flex-col mt-1.5 ml-5 max-h-48 overflow-y-auto;
  }

  .describe-list-item {
    @apply flex items-baseline justify-between gap-2 py-0.5 text-sm;
  }

  .describe-list-name {
    @apply text-fg-primary truncate;
  }

  .describe-list-detail {
    @apply text-xs text-fg-muted font-mono truncate max-w-[200px] text-right;
  }

  .describe-description {
    @apply text-sm text-fg-secondary leading-relaxed;
  }

  .describe-ai-instructions {
    @apply text-xs font-mono bg-surface-subtle p-2 rounded overflow-auto whitespace-pre-wrap mt-1;
    max-height: 120px;
  }

  .describe-yaml {
    @apply text-xs font-mono bg-surface-subtle p-3 rounded overflow-auto whitespace-pre-wrap;
    max-height: 200px;
  }
</style>
