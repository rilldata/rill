<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import TabNav from "@rilldata/web-common/components/nav/TabNav.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    countByKind,
    groupErrorsByKind,
    pluralizeKind,
  } from "@rilldata/web-common/features/resources/overview-utils";
  import {
    createRuntimeServiceGetInstance,
    createRuntimeServiceGetResource,
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createLocalServiceGetVersion } from "@rilldata/web-common/runtime-client/local-service";
  import Rocket from "svelte-radix/Rocket.svelte";
  import ResourcesSection from "../status/ResourcesSection.svelte";
  import ParseErrorsSection from "../status/ParseErrorsSection.svelte";

  function formatConnectorName(name: string | undefined): string {
    if (!name) return "—";
    const lower = name.toLowerCase();
    if (lower === "duckdb") return "DuckDB";
    if (lower === "clickhouse") return "ClickHouse";
    if (lower === "mysql") return "MySQL";
    if (lower === "bigquery") return "BigQuery";
    if (lower === "openai") return "OpenAI";
    if (lower === "claude") return "Claude";
    if (lower === "gemini") return "Gemini";
    return name.charAt(0).toUpperCase() + name.slice(1);
  }

  let selectedPage = "overview";

  const navItems = [
    { label: "Overview", value: "overview" },
    { label: "Resources", value: "resources" },
    { label: "Logs", value: "logs" },
  ];

  $: ({ instanceId } = $runtime);

  // Instance query for connector info
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;

  // Runtime version
  $: versionQuery = createLocalServiceGetVersion();
  $: version = $versionQuery.data?.version ?? "";

  // Resources query for overview sections
  $: resourcesQuery = createRuntimeServiceListResources(instanceId, {});
  $: allResources = ($resourcesQuery.data?.resources ?? []) as V1Resource[];
  $: resourceCounts = countByKind(allResources);

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  // Resource errors grouped by kind
  $: erroredResources = allResources.filter((r) => !!r.meta?.reconcileError);
  $: errorsByKind = groupErrorsByKind(erroredResources);
  $: totalErrors = parseErrors.length + erroredResources.length;

  let resourceStatusFilter: string[] = [];
  let resourceTypeFilter: string[] = [];
  let lastSelectedPage = selectedPage;

  // Clear pre-set filters when user switches tabs via TabNav (not via goToResources)
  $: if (selectedPage !== lastSelectedPage) {
    if (selectedPage === "resources" && lastSelectedPage !== "overview") {
      // Only reset if not coming from overview (goToResources sets filters before switching)
      resourceStatusFilter = [];
      resourceTypeFilter = [];
    }
    lastSelectedPage = selectedPage;
  }

  function goToResources(statusFilter: string[] = [], typeFilter: string[] = []) {
    resourceStatusFilter = statusFilter;
    resourceTypeFilter = typeFilter;
    selectedPage = "resources";
  }
</script>

<ContentContainer title="Status" maxWidth={1100}>
  <div class="flex pt-6 gap-6 max-w-full overflow-hidden">
    <TabNav items={navItems} bind:selected={selectedPage} />

    <!-- Main Content -->
    <div class="flex flex-col gap-y-6 w-full overflow-hidden">
      {#if selectedPage === "overview"}
        <!-- Project Section -->
        <div class="section">
          <div class="section-header">
            <h3 class="section-title">Project</h3>
          </div>
          <div class="info-grid">
            <div class="info-row">
              <span class="info-label">Status</span>
              <span class="info-value flex items-center gap-2">
                <span class="status-dot bg-green-500"></span>
                Running
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
                {instance?.aiConnector
                  ? formatConnectorName(instance.aiConnector)
                  : "Rill Managed"}
              </span>
            </div>
          </div>
        </div>

        <!-- Resources Section -->
        {#if resourceCounts.length > 0}
          <div class="section">
            <div class="section-header">
              <h3 class="section-title">Resources</h3>
              <button class="view-all" on:click={() => goToResources()}>View all</button
              >
            </div>
            <div class="resource-chips">
              {#each resourceCounts as { kind, label, count } (kind)}
                <button class="resource-chip" on:click={() => goToResources([], [kind])}>
                  {#if resourceIconMapping[kind]}
                    <svelte:component
                      this={resourceIconMapping[kind]}
                      size="12px"
                    />
                  {/if}
                  <span class="font-medium">{count}</span>
                  <span class="text-fg-secondary"
                    >{pluralizeKind(label, count)}</span
                  >
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Errors Section -->
        {#if totalErrors > 0}
          <button
            class="section section-error section-clickable"
            on:click={() => goToResources(["error"])}
          >
            <div class="section-header">
              <h3 class="section-title flex items-center gap-2">
                Errors
                <span class="error-badge">{totalErrors}</span>
              </h3>
            </div>
            <div class="error-chips">
              {#if parseErrors.length > 0}
                <span class="error-chip">
                  <AlertCircleOutline size="12px" />
                  <span class="font-medium">{parseErrors.length}</span>
                  <span
                    >Parse error{parseErrors.length !== 1 ? "s" : ""}</span
                  >
                </span>
              {/if}
              {#each errorsByKind as { kind, label, count } (kind)}
                <span class="error-chip">
                  {#if resourceIconMapping[kind]}
                    <svelte:component
                      this={resourceIconMapping[kind]}
                      size="12px"
                    />
                  {/if}
                  <span class="font-medium">{count}</span>
                  <span>{pluralizeKind(label, count)}</span>
                </span>
              {/each}
            </div>
          </button>
        {:else}
          <div class="section">
            <div class="section-header">
              <h3 class="section-title flex items-center gap-2">Errors</h3>
            </div>
            {#if $projectParserQuery.isError || $resourcesQuery.isError}
              <p class="text-sm text-fg-secondary">
                Unable to check for errors.
              </p>
            {:else if $projectParserQuery.isLoading || $resourcesQuery.isLoading}
              <p class="text-sm text-fg-secondary">Checking for errors...</p>
            {:else}
              <p class="text-sm text-fg-secondary">No errors detected.</p>
            {/if}
          </div>
        {/if}
      {:else if selectedPage === "resources"}
        <ResourcesSection
          initialStatusFilter={resourceStatusFilter}
          initialTypeFilter={resourceTypeFilter}
        />
        <ParseErrorsSection />
      {:else if selectedPage === "logs"}
        <div class="section">
          <div class="section-header">
            <h3 class="section-title">Logs</h3>
          </div>
          <p class="text-sm text-fg-muted">
            Real-time logs are available after deploying to Rill Cloud.
          </p>
          <div class="mt-4">
            <Button type="primary" href="/deploy" compact>
              <Rocket size="14px" />
              Deploy to Rill Cloud
            </Button>
          </div>
        </div>
      {/if}
    </div>
  </div>
</ContentContainer>

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
  .view-all {
    @apply text-xs text-primary-500 bg-transparent border-none cursor-pointer p-0;
  }
  .view-all:hover {
    @apply text-primary-600;
  }
  .resource-chips {
    @apply flex flex-wrap gap-2;
  }
  .resource-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle cursor-pointer;
  }
  .resource-chip:hover {
    @apply border-primary-500 text-primary-600;
  }
  .section-clickable {
    @apply cursor-pointer;
  }
  .section-error {
    @apply border-red-500;
  }
  .section-clickable:hover {
    @apply border-red-600;
  }
  .error-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .error-chips {
    @apply flex flex-wrap gap-2;
  }
  .error-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md;
    @apply border border-red-300 bg-red-50 text-red-700;
  }
</style>
