<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import TabNav from "@rilldata/web-common/components/nav/TabNav.svelte";
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
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createLocalServiceGetVersion } from "@rilldata/web-common/runtime-client/local-service";
  import Rocket from "svelte-radix/Rocket.svelte";
  import ResourcesSection from "../status/ResourcesSection.svelte";
  import ParseErrorsSection from "../status/ParseErrorsSection.svelte";

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
  $: version = $versionQuery.data?.current ?? "";

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

  function goToResources(
    statusFilter: string[] = [],
    typeFilter: string[] = [],
  ) {
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

        <ErrorsOverviewSection
          parseErrorCount={parseErrors.length}
          {errorsByKind}
          {totalErrors}
          isLoading={$projectParserQuery.isLoading || $resourcesQuery.isLoading}
          isError={$projectParserQuery.isError || $resourcesQuery.isError}
          onSectionClick={() => goToResources(["error"])}
        />
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
            Real-time logs are available after deploying.
          </p>
          <div class="mt-4">
            <Button type="primary" href="/deploy" compact>
              <Rocket size="14px" />
              Deploy to unlock
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
</style>
