<script lang="ts">
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { formatConnectorName } from "@rilldata/web-common/features/resources/display-utils";
  import ErrorsOverviewSection from "@rilldata/web-common/features/resources/overview/ErrorsOverviewSection.svelte";
  import ResourcesOverviewSection from "@rilldata/web-common/features/resources/overview/ResourcesOverviewSection.svelte";
  import {
    countByKind,
    groupErrorsByKind,
  } from "@rilldata/web-common/features/resources/overview-utils";
  import {
    createRuntimeServiceGetInstance,
    createRuntimeServiceGetResource,
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createLocalServiceGetVersion } from "@rilldata/web-common/runtime-client/local-service";
  import { goto } from "$app/navigation";
  import TablesSection from "../../features/tables/TablesSection.svelte";

  const runtimeClient = useRuntimeClient();

  // Instance query for connector info
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;

  // Runtime version
  $: versionQuery = createLocalServiceGetVersion();
  $: version = $versionQuery.data?.current ?? "";

  // Resources query for overview sections
  $: resourcesQuery = createRuntimeServiceListResources(runtimeClient, {});
  $: allResources = ($resourcesQuery.data?.resources ?? []) as V1Resource[];
  $: resourceCounts = countByKind(allResources);

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    runtimeClient,
    {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  // Resource errors grouped by kind
  $: erroredResources = allResources.filter(
    (r) =>
      !!r.meta?.reconcileError && r.meta?.name?.kind !== ResourceKind.Component,
  );
  $: errorsByKind = groupErrorsByKind(erroredResources);
  $: totalErrors = parseErrors.length + erroredResources.length;

  function goToResources(
    statusFilter: string[] = [],
    typeFilter: string[] = [],
  ) {
    const params = new URLSearchParams();
    if (statusFilter.length > 0) params.set("status", statusFilter.join(","));
    if (typeFilter.length > 0) params.set("kind", typeFilter.join(","));
    const search = params.toString();
    void goto(`/status/resources${search ? `?${search}` : ""}`);
  }
</script>

<!-- Project Section -->
<div class="section">
  <div class="section-header">
    <h3 class="section-title">Project</h3>
  </div>
  <div class="info-grid">
    <div class="info-row">
      <span class="info-label">Status</span>
      <span class="info-value flex items-center gap-2">
        {#if $projectParserQuery.isLoading || $resourcesQuery.isLoading}
          <span class="status-dot bg-gray-400"></span>
          Loading
        {:else if totalErrors > 0}
          <span class="status-dot bg-red-500"></span>
          {totalErrors}
          {totalErrors === 1 ? "error" : "errors"}
        {:else}
          <span class="status-dot bg-green-500"></span>
          Running
        {/if}
      </span>
    </div>
    <div class="info-row">
      <span class="info-label">Environment</span>
      <span class="info-value">Development</span>
    </div>
    {#if version}
      <div class="info-row">
        <span class="info-label">Runtime</span>
        <span class="info-value font-mono text-xs">{version}</span>
      </div>
    {/if}
    <div class="info-row">
      <span class="info-label">OLAP Engine</span>
      <span class="info-value">
        {formatConnectorName(instance?.olapConnector || "duckdb")}
      </span>
    </div>
    <div class="info-row">
      <span class="info-label">AI Connector</span>
      <span class="info-value">
        {instance?.aiConnector && instance.aiConnector !== "admin"
          ? formatConnectorName(instance.aiConnector)
          : "Rill Managed"}
      </span>
    </div>
  </div>
</div>

<ResourcesOverviewSection
  {resourceCounts}
  onViewAll={() => goToResources()}
  onChipClick={(kind) => goToResources([], [kind])}
/>

<TablesSection />

<ErrorsOverviewSection
  parseErrorCount={parseErrors.length}
  {errorsByKind}
  {totalErrors}
  isLoading={$projectParserQuery.isLoading || $resourcesQuery.isLoading}
  isError={$projectParserQuery.isError || $resourcesQuery.isError}
  onSectionClick={() => goToResources(["error"])}
  onParseErrorChipClick={() => goToResources(["error"])}
  onKindChipClick={(kind) => goToResources(["error"], [kind])}
/>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5 text-left w-full;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .info-grid {
    @apply flex flex-col;
  }
  .info-row {
    @apply flex items-center py-2;
  }
  .info-label {
    @apply text-sm text-fg-secondary w-32 shrink-0;
  }
  .info-value {
    @apply text-sm text-fg-primary;
  }
  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }
</style>
