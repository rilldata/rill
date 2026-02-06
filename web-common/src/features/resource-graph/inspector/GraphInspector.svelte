<script lang="ts">
  import { selectedGraphNode, isGraphExpanded } from "./graph-inspector-store";
  import { displayResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    resourceIconMapping,
    resourceShorthandMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping";
  import {
    V1ReconcileStatus,
    createRuntimeServiceCreateTrigger,
  } from "@rilldata/web-common/runtime-client";
  import { goto } from "$app/navigation";
  import {
    Database,
    RefreshCw,
    Clock,
    RotateCcw,
    Palette,
    Bell,
    Plug,
    Layers,
    FileCode,
    AlertCircle,
    ExternalLink,
    GitFork,
    Building2,
    FolderGit2,
    Github,
    HardDrive,
    Sparkles,
    CheckCircle2,
    Circle,
    Table2,
    BarChart3,
    LayoutGrid,
    Component,
    FileText,
    Settings,
    Zap,
  } from "lucide-svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGitStatus,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import PartitionsTable from "@rilldata/web-common/features/models/partitions/PartitionsTable.svelte";
  import PartitionsFilter from "@rilldata/web-common/features/models/partitions/PartitionsFilter.svelte";
  import type { Selected } from "bits-ui";

  const DEFAULT_COLOR = "#6B7280";
  const DEFAULT_ICON = resourceIconMapping[ResourceKind.Model];

  // Partitions modal state
  let selectedPartitionFilter = "all";

  function onPartitionFilterChange(newSelection: Selected<string>) {
    selectedPartitionFilter = newSelection.value;
  }

  // Project info queries
  $: ({ instanceId } = $runtime);
  $: projectTitleQuery = useProjectTitle(instanceId);
  $: projectTitle = $projectTitleQuery.data ?? "Untitled Project";
  $: currentProjectQuery = createLocalServiceGetCurrentProject();
  $: currentProject = $currentProjectQuery.data;
  $: gitStatusQuery = createLocalServiceGitStatus();
  $: gitStatus = $gitStatusQuery.data;

  $: data = $selectedGraphNode;
  $: kind = data?.kind;
  $: icon =
    kind && resourceIconMapping[kind]
      ? resourceIconMapping[kind]
      : DEFAULT_ICON;
  $: color =
    kind && resourceShorthandMapping[kind]
      ? `var(--${resourceShorthandMapping[kind]})`
      : DEFAULT_COLOR;
  $: metadata = data?.metadata;
  $: resource = data?.resource;
  $: reconcileStatus = resource?.meta?.reconcileStatus;
  $: hasError = !!resource?.meta?.reconcileError;
  $: isIdle = reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: statusLabel =
    reconcileStatus && !isIdle
      ? reconcileStatus
          ?.replace("RECONCILE_STATUS_", "")
          ?.toLowerCase()
          ?.replaceAll("_", " ")
      : undefined;

  // Get resource info for navigation and actions
  $: resourceName = resource?.meta?.name?.name ?? "";
  // Use filePaths directly from resource meta (more reliable than artifact lookup)
  $: filePath = resource?.meta?.filePaths?.[0];

  // Get connector-specific icon
  $: connectorIcon =
    metadata?.connector &&
    connectorIconMapping[
      metadata.connector.toLowerCase() as keyof typeof connectorIconMapping
    ]
      ? connectorIconMapping[
          metadata.connector.toLowerCase() as keyof typeof connectorIconMapping
        ]
      : null;


  // Model refresh mutation
  const triggerMutation = createRuntimeServiceCreateTrigger();

  function refreshModel() {
    if (!resourceName) return;
    $triggerMutation.mutate({
      instanceId,
      data: {
        models: [{ model: resourceName, full: false }],
      },
    });
  }

  function openFile() {
    if (!filePath) return;
    try {
      const prefs = JSON.parse(localStorage.getItem(filePath) || "{}");
      localStorage.setItem(filePath, JSON.stringify({ ...prefs, view: "code" }));
    } catch (error) {
      console.warn(`Failed to save file view preference:`, error);
    }
    goto(`/files${filePath}`);
  }

  function handleViewLineage() {
    if (!resource?.meta?.name) return;
    const resourceKindName = resource.meta.name.kind;
    const resourceNameValue = resource.meta.name.name;

    // Use the resource ID as both the seed (via resource param) and expanded target
    // This ensures the lineage group containing this resource is visible and expanded
    const resourceId = encodeURIComponent(
      `${resourceKindName}:${resourceNameValue}`,
    );
    goto(`/graph?resource=${resourceId}&expanded=${resourceId}`);
  }
</script>

<div class="inspector-panel">
  {#if data}
    <!-- Node Info -->
    <div class="inspector-header">
      <div class="header-icon" style={`background:${color}20`}>
        <svelte:component this={icon} size="20px" {color} />
      </div>
      <div class="header-info">
        <h3 class="header-title" title={data.label}>{data.label}</h3>
        <p class="header-type">
          {#if kind}
            {displayResourceKind(kind)}
          {:else}
            Unknown
          {/if}
        </p>
      </div>
    </div>

    <!-- Connection Info -->
    {#if metadata?.connector || metadata?.sourcePath}
      <div class="section">
        <h4 class="section-title">Connector</h4>
        {#if metadata?.connector}
          <div class="metadata-row">
            <span class="metadata-icon">
              {#if connectorIcon}
                <svelte:component this={connectorIcon} size="16px" />
              {:else}
                <Database size={16} />
              {/if}
            </span>
            <span class="metadata-value">{metadata.connector}</span>
          </div>
        {/if}
        {#if metadata?.sourcePath}
          <div class="metadata-row">
            <span class="metadata-icon file">
              <FileText size={14} />
            </span>
            <span
              class="metadata-value source-path"
              title={metadata.sourcePath}
            >
              {metadata.sourcePath}
            </span>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Status section -->
    {#if hasError || statusLabel}
      <div class="section">
        <h4 class="section-title">Status</h4>
        {#if hasError}
          <div class="status-error">
            <AlertCircle size={14} />
            <span>Error</span>
          </div>
          <pre class="error-message">{resource?.meta?.reconcileError}</pre>
        {:else if statusLabel}
          <div class="status-item">
            <span class="status-label">{statusLabel}</span>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Model/Source Information -->
    {#if kind === ResourceKind.Model || kind === ResourceKind.Source}
      <div class="section">
        <div class="section-header-inline">
          <h4 class="section-title">{displayResourceKind(kind)} Information</h4>
          <span class="definition-type">{metadata?.isSqlModel ? "SQL" : "YAML"}</span>
        </div>

        <!-- Feature buttons row -->
        {#if metadata?.partitioned || metadata?.incremental}
          <div class="feature-buttons">
            {#if metadata?.partitioned && resource}
              <Dialog.Root>
                <Dialog.Trigger asChild let:builder>
                  <button
                    use:builder.action
                    {...builder}
                    class="feature-btn"
                    type="button"
                  >
                    <Layers size={14} />
                    <span>Partitioned</span>
                    {#if metadata?.partitionsHaveErrors}
                      <span class="error-dot"></span>
                    {/if}
                  </button>
                </Dialog.Trigger>
                <Dialog.Content class="max-w-screen-xl">
                  <Dialog.Header>
                    <Dialog.Title>Model partitions</Dialog.Title>
                  </Dialog.Header>
                  <div class="flex justify-end">
                    <PartitionsFilter
                      selectedFilter={selectedPartitionFilter}
                      onChange={onPartitionFilterChange}
                    />
                  </div>
                  <PartitionsTable
                    {resource}
                    whereErrored={selectedPartitionFilter === "errors"}
                    wherePending={selectedPartitionFilter === "pending"}
                  />
                </Dialog.Content>
              </Dialog.Root>
            {/if}
            {#if metadata?.incremental}
              <button class="feature-btn" type="button">
                <RefreshCw size={14} />
                <span>Incremental</span>
              </button>
            {/if}
          </div>
        {/if}
        <!-- Tests YAML -->
        {#if metadata?.testsYaml}
          <div class="yaml-section">
            <div class="yaml-header">
              <Zap size={12} />
              <span>Tests ({metadata.testCount})</span>
            </div>
            <pre class="yaml-code">{metadata.testsYaml}</pre>
          </div>
        {/if}

        <!-- Schedule -->
        {#if metadata?.hasSchedule}
          <div class="yaml-section">
            <div class="yaml-header">
              <Clock size={12} />
              <span>Schedule </span>
            </div>
            {#if metadata?.scheduleYaml}
              <pre class="yaml-code">{metadata.scheduleYaml}</pre>
            {/if}
          </div>
        {/if}

        <!-- Retry YAML -->
        {#if metadata?.retryYaml}
          <div class="yaml-section">
            <div class="yaml-header">
              <RotateCcw size={12} />
              <span>Retry</span>
            </div>
            <pre class="yaml-code">{metadata.retryYaml}</pre>
          </div>
        {/if}
      </div>
    {/if}

    <!-- MetricsView Information -->
    {#if kind === ResourceKind.MetricsView}
      <div class="section">
        <h4 class="section-title">MetricsView Information</h4>

        <!-- Data Source -->
        {#if metadata?.metricsConnector || metadata?.metricsTable || metadata?.metricsModel}
          <div class="subsection">
            <span class="subsection-title">Data Source</span>
            {#if metadata?.metricsConnector}
              <div class="metadata-row">
                <span class="metadata-icon">
                  <Database size={14} />
                </span>
                <span class="metadata-value">{metadata.metricsConnector}</span>
              </div>
            {/if}
            {#if metadata?.metricsTable}
              <div class="metadata-row">
                <span class="metadata-icon table">
                  <Table2 size={14} />
                </span>
                <span class="metadata-value">{metadata.metricsTable}</span>
              </div>
            {/if}
            {#if metadata?.metricsModel}
              <div class="metadata-row">
                <span class="metadata-icon model">
                  <LayoutGrid size={14} />
                </span>
                <span class="metadata-value">{metadata.metricsModel}</span>
              </div>
            {/if}
            {#if metadata?.timeDimension}
              <div class="metadata-row">
                <span class="metadata-icon time">
                  <Clock size={14} />
                </span>
                <span class="metadata-value">{metadata.timeDimension}</span>
              </div>
            {/if}
          </div>
        {/if}

        <!-- Dimensions -->
        {#if metadata?.dimensions && metadata.dimensions.length > 0}
          <div class="subsection">
            <span class="subsection-title"
              >Dimensions ({metadata.dimensions.length})</span
            >
            <div class="field-list">
              {#each metadata.dimensions as dim}
                <div class="field-item">
                  <span class="field-name">{dim.displayName || dim.name}</span>
                  {#if dim.type && dim.type !== "UNSPECIFIED"}
                    <span class="field-type">{dim.type.toLowerCase()}</span>
                  {/if}
                  {#if dim.description}
                    <span class="field-desc">{dim.description}</span>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Measures -->
        {#if metadata?.measures && metadata.measures.length > 0}
          <div class="subsection">
            <span class="subsection-title"
              >Measures ({metadata.measures.length})</span
            >
            <div class="field-list">
              {#each metadata.measures as measure}
                <div class="field-item">
                  <span class="field-name"
                    >{measure.displayName || measure.name}</span
                  >
                  {#if measure.expression}
                    <code class="field-expr">{measure.expression}</code>
                  {/if}
                  {#if measure.description}
                    <span class="field-desc">{measure.description}</span>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Explore Information -->
    {#if kind === ResourceKind.Explore}
      <div class="section">
        <h4 class="section-title">Explore Information</h4>
        {#if metadata?.metricsViewName}
          <div class="metadata-row">
            <span class="metadata-icon metrics">
              <BarChart3 size={14} />
            </span>
            <span class="metadata-value">{metadata.metricsViewName}</span>
          </div>
        {/if}
        {#if metadata?.theme}
          <div class="metadata-row">
            <span class="metadata-icon theme">
              <Palette size={14} />
            </span>
            <span class="metadata-value">{metadata.theme}</span>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Canvas Information -->
    {#if kind === ResourceKind.Canvas}
      <div class="section">
        <h4 class="section-title">Canvas Information</h4>
        {#if metadata?.componentCount}
          <div class="metadata-row">
            <span class="metadata-icon component">
              <Component size={14} />
            </span>
            <span class="metadata-value">
              {metadata.componentCount} component{metadata.componentCount > 1
                ? "s"
                : ""}
            </span>
          </div>
        {/if}
        {#if metadata?.rowCount}
          <div class="metadata-row">
            <span class="metadata-icon layout">
              <LayoutGrid size={14} />
            </span>
            <span class="metadata-value">
              {metadata.rowCount} row{metadata.rowCount > 1 ? "s" : ""}
            </span>
          </div>
        {/if}
        {#if metadata?.theme}
          <div class="metadata-row">
            <span class="metadata-icon theme">
              <Palette size={14} />
            </span>
            <span class="metadata-value">{metadata.theme}</span>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Consumers section -->
    {#if metadata?.alertCount || metadata?.apiCount}
      <div class="section">
        <h4 class="section-title">Consumers</h4>
        {#if metadata?.alertCount}
          <div class="metadata-row">
            <span class="metadata-icon alert">
              <Bell size={14} />
            </span>
            <span class="metadata-value">
              {metadata.alertCount} alert{metadata.alertCount > 1 ? "s" : ""}
            </span>
          </div>
        {/if}
        {#if metadata?.apiCount}
          <div class="metadata-row">
            <span class="metadata-icon api">
              <Plug size={14} />
            </span>
            <span class="metadata-value">
              {metadata.apiCount} API{metadata.apiCount > 1 ? "s" : ""}
            </span>
          </div>
        {/if}
      </div>
    {/if}

    <!-- SQL Query section (full height) -->
    {#if metadata?.sqlQuery}
      <div class="section sql-section">
        <h4 class="section-title">SQL Query</h4>
        <pre class="sql-code-full">{metadata.sqlQuery}</pre>
      </div>
    {/if}

    <!-- Actions -->
    <div class="actions">
      {#if (kind === ResourceKind.Model || kind === ResourceKind.Source) && resourceName}
        <Button type="secondary" onClick={refreshModel}>
          <RefreshCw size={14} />
          Refresh
        </Button>
      {/if}
      {#if filePath}
        <Button type="secondary" onClick={openFile}>
          <ExternalLink size={14} />
          Edit YAML
        </Button>
      {/if}
      {#if !$isGraphExpanded}
        <Button type="secondary" onClick={handleViewLineage}>
          <GitFork size={14} />
          View lineage
        </Button>
      {/if}
    </div>
  {:else}
    <!-- Project Info (when no node selected) -->
    <div class="project-info">
      <div class="inspector-header">
        <div class="header-icon project">
          <FolderGit2 size="20px" />
        </div>
        <div class="header-info">
          <h3 class="header-title" title={projectTitle}>{projectTitle}</h3>
          <p class="header-type">Project</p>
        </div>
      </div>

      <!-- Organization -->
      {#if currentProject?.project?.orgName}
        <div class="section">
          <h4 class="section-title">Organization</h4>
          <div class="metadata-row">
            <span class="metadata-icon org">
              <Building2 size={14} />
            </span>
            <span class="metadata-value">{currentProject.project.orgName}</span>
          </div>
        </div>
      {/if}

      <!-- GitHub Status -->
      <div class="section">
        <h4 class="section-title">GitHub</h4>
        <div class="metadata-row">
          <span class="metadata-icon github">
            <Github size={14} />
          </span>
          <span class="metadata-value">
            {#if gitStatus?.managedGit}
              Rill-Managed
            {:else if gitStatus?.githubUrl}
              Connected
            {:else}
              Not connected
            {/if}
          </span>
        </div>
        {#if gitStatus?.githubUrl && !gitStatus?.managedGit}
          <div class="metadata-row sub">
            <span
              class="metadata-value muted truncate"
              title={gitStatus.githubUrl}
            >
              {gitStatus.githubUrl.replace("https://github.com/", "")}
            </span>
          </div>
        {/if}
      </div>

      <!-- Status -->
      <div class="section">
        <h4 class="section-title">Status</h4>
        <div class="metadata-row">
          <span class="metadata-icon status-ok">
            <CheckCircle2 size={14} />
          </span>
          <span class="metadata-value">Running</span>
        </div>
      </div>

      <!-- Storage -->
      <div class="section">
        <h4 class="section-title">Storage</h4>
        <div class="metadata-row">
          <span class="metadata-icon storage">
            <HardDrive size={14} />
          </span>
          <span class="metadata-value">
            {#if currentProject?.project}
              Cloud Storage
            {:else}
              Local Storage
            {/if}
          </span>
        </div>
      </div>

      <!-- AI -->
      <div class="section">
        <h4 class="section-title">AI Connector</h4>
        <div class="metadata-row">
          <span class="metadata-icon ai">
            <Sparkles size={14} />
          </span>
          <span class="metadata-value">
            {#if currentProject?.project}
              Enabled
            {:else}
              Available
            {/if}
          </span>
        </div>
      </div>

      <div class="hint">
        <Circle size={6} class="text-fg-muted" />
        <span>Click a node to view details</span>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .inspector-panel {
    @apply h-full flex flex-col overflow-y-auto py-2;
  }

  .inspector-header {
    @apply flex items-center gap-3 p-4 border-b;
  }

  .header-icon {
    @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg;
  }

  .header-icon.project {
    @apply bg-primary-100 text-primary-600;
  }

  .header-info {
    @apply flex flex-col min-w-0 flex-1;
  }

  .inspector-header :global(button) {
    @apply flex-shrink-0 text-fg-muted;
  }

  .header-title {
    @apply font-semibold text-sm text-fg-primary;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .header-type {
    @apply text-xs text-fg-secondary capitalize;
  }

  .section {
    @apply px-4 py-3 border-b;
  }

  .section-title {
    @apply text-xs font-semibold text-fg-secondary uppercase tracking-wide mb-2;
  }

  .section-header-inline {
    @apply flex items-center justify-between mb-2;
  }

  .section-header-inline .section-title {
    @apply mb-0;
  }

  .definition-type {
    @apply text-xs font-medium px-2 py-0.5 rounded;
    @apply bg-gray-100 text-gray-600;
  }

  .metadata-row {
    @apply flex items-center gap-2 py-1;
  }

  .metadata-row.sub {
    @apply pl-7;
  }

  .metadata-icon {
    @apply flex items-center justify-center w-5 h-5 rounded;
  }

  .metadata-icon.sql {
    @apply text-emerald-600;
  }

  .metadata-icon.incremental {
    @apply text-cyan-600;
  }

  .metadata-icon.partitioned {
    @apply text-purple-600;
  }

  .metadata-icon.scheduled {
    @apply text-amber-600;
  }

  .metadata-icon.retry {
    @apply text-orange-600;
  }

  .metadata-icon.theme {
    @apply text-pink-600;
  }

  .metadata-icon.alert {
    @apply text-red-600;
  }

  .metadata-icon.api {
    @apply text-blue-600;
  }

  .metadata-icon.org {
    @apply text-indigo-600;
  }

  .metadata-icon.github {
    @apply text-fg-primary;
  }

  .metadata-icon.status-ok {
    @apply text-green-600;
  }

  .metadata-icon.storage {
    @apply text-slate-600;
  }

  .metadata-icon.ai {
    @apply text-violet-600;
  }

  .metadata-icon.file {
    @apply text-gray-600;
  }

  .metadata-icon.output {
    @apply text-teal-600;
  }

  .metadata-icon.stage {
    @apply text-slate-500;
  }

  .metadata-icon.settings {
    @apply text-gray-600;
  }

  .metadata-icon.test {
    @apply text-yellow-600;
  }

  .metadata-value {
    @apply text-sm text-fg-primary;
  }

  .metadata-value.muted {
    @apply text-fg-muted text-xs;
  }

  .metadata-value.truncate {
    @apply max-w-[180px];
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .metadata-value.source-path {
    @apply text-xs font-mono max-w-[180px];
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .metadata-value.error-text {
    @apply text-red-600 text-xs;
  }

  .metadata-sub {
    @apply text-fg-muted text-xs ml-1;
  }

  .status-error {
    @apply flex items-center gap-2 text-red-600 text-sm font-medium mb-2;
  }

  .error-message {
    @apply text-xs text-fg-secondary bg-surface-subtle p-2 rounded overflow-auto max-h-32 whitespace-pre-wrap;
  }

  .status-item {
    @apply text-sm text-fg-primary capitalize;
  }

  .actions {
    @apply flex flex-col gap-2 p-4 mt-auto;
  }

  .actions :global(button) {
    @apply w-full justify-center gap-2;
  }

  .project-info {
    @apply flex flex-col h-full;
  }

  .hint {
    @apply flex items-center gap-2 px-4 py-3 mt-auto text-xs text-fg-muted;
  }

  /* Icon indicators row */
  .icon-indicators {
    @apply flex flex-wrap gap-2 mb-3;
  }

  .indicator-badge {
    @apply flex items-center gap-1.5 px-2 py-1 rounded-md text-xs font-medium;
    @apply border-none bg-gray-100 text-gray-700;
  }

  .indicator-badge.sql {
    @apply bg-emerald-50 text-emerald-700;
  }

  .indicator-badge.incremental {
    @apply bg-cyan-50 text-cyan-700;
  }

  .indicator-badge.partitioned {
    @apply bg-purple-50 text-purple-700;
  }

  .indicator-badge.mode {
    @apply bg-slate-100 text-slate-700;
  }

  .indicator-badge.clickable {
    @apply cursor-pointer transition-colors;
  }

  .indicator-badge.clickable:hover {
    @apply bg-purple-100;
  }

  /* Property list (checkbox-style) */
  .property-list {
    @apply flex flex-col gap-1;
  }

  .property-item {
    @apply flex items-center gap-2 text-sm text-fg-primary py-1;
  }

  .property-item.clickable {
    @apply cursor-pointer rounded px-2 -mx-2 transition-colors;
    @apply border-none bg-transparent;
  }

  .property-item.clickable:hover {
    @apply bg-surface-subtle;
  }

  :global(.property-check) {
    @apply text-green-600 flex-shrink-0;
  }

  /* Feature buttons row */
  .feature-buttons {
    @apply flex flex-wrap gap-2 mb-3;
  }

  .feature-btn {
    @apply inline-flex items-center gap-1.5 px-3 py-1.5;
    @apply text-xs font-medium text-fg-secondary;
    @apply bg-surface-subtle rounded-md border border-gray-200;
    @apply cursor-default transition-colors;
  }

  :global(button.feature-btn) {
    @apply cursor-pointer;
  }

  :global(button.feature-btn:hover) {
    @apply bg-gray-100 border-gray-300;
  }

  /* Source Model label */
  .source-model-label {
    @apply text-xs text-fg-muted mb-2;
  }

  /* YAML sections */
  .yaml-section {
    @apply mb-3;
  }

  .yaml-header {
    @apply flex items-center gap-1.5 text-xs font-medium text-fg-secondary mb-1;
  }

  .yaml-code {
    @apply text-xs font-mono bg-gray-50 p-2 rounded border border-gray-200;
    @apply overflow-x-auto whitespace-pre;
    max-height: 120px;
  }

  /* Subsections within type info */
  .subsection {
    @apply mb-3;
  }

  .subsection:last-child {
    @apply mb-0;
  }

  .subsection-title {
    @apply text-xs font-medium text-fg-muted mb-1.5 block;
  }

  /* SQL code block - full height version */
  .sql-section {
    @apply flex-1 flex flex-col min-h-0;
  }

  .sql-code-full {
    @apply text-xs font-mono bg-surface-subtle p-3 rounded overflow-auto whitespace-pre-wrap flex-1;
    @apply border border-gray-200;
    min-height: 200px;
  }

  /* Field list for dimensions/measures */
  .field-list {
    @apply flex flex-col gap-1 mt-1;
  }

  .field-item {
    @apply flex flex-col gap-0.5 px-2 py-1.5 rounded bg-surface-subtle;
  }

  .field-name {
    @apply text-sm font-medium text-fg-primary;
  }

  .field-type {
    @apply text-xs text-fg-muted uppercase;
  }

  .field-expr {
    @apply text-xs font-mono text-fg-secondary bg-gray-100 px-1 py-0.5 rounded;
  }

  .field-desc {
    @apply text-xs text-fg-muted;
  }

  /* Additional metadata icon variants */
  .metadata-icon.table {
    @apply text-blue-600;
  }

  .metadata-icon.time {
    @apply text-amber-600;
  }

  .metadata-icon.model {
    @apply text-teal-600;
  }

  .metadata-icon.metrics {
    @apply text-purple-600;
  }

  .metadata-icon.component {
    @apply text-indigo-600;
  }

  .metadata-icon.layout {
    @apply text-slate-600;
  }

  /* Partitions dialog styles */
  .error-dot {
    @apply w-2 h-2 rounded-full bg-red-500;
    margin-left: 2px;
  }

  :global(.partitions-model-name) {
    @apply font-medium text-fg-primary;
  }

  :global(.partitions-status) {
    @apply inline-flex items-center gap-1 text-xs ml-2;
  }

  :global(.partitions-status.error) {
    @apply text-red-600;
  }

  :global(.partitions-status.success) {
    @apply text-green-600;
  }

  .partitions-toolbar {
    @apply flex items-center justify-between gap-4 mb-4;
  }

  .partitions-meta {
    @apply flex items-center gap-4 text-xs text-fg-muted;
  }

  .partitions-meta .meta-item {
    @apply inline-flex items-center gap-1;
  }
</style>
