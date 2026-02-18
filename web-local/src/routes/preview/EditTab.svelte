<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { workspaces as workspaceStore } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import CanvasWorkspace from "@rilldata/web-common/features/workspaces/CanvasWorkspace.svelte";
  import ExploreWorkspace from "@rilldata/web-common/features/workspaces/ExploreWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client/gen/runtime-service/runtime-service";
  import { resourceSectionState } from "./resource-section-store";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type { ComponentType } from "svelte";
  import ResourceListSection from "./ResourceListSection.svelte";

  interface Resource {
    name: string;
    kind: ResourceKind | string;
    state?: string;
    error?: string;
    path?: string;
  }

  const workspaceComponents = new Map<
    ResourceKind | null | undefined,
    ComponentType | null
  >([
    [ResourceKind.Source, ModelWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Explore, ExploreWorkspace],
    [ResourceKind.Canvas, CanvasWorkspace],
    [null, null],
    [undefined, null],
  ]);

  let allResources: Resource[] = [];
  let metricsResources: Resource[] = [];
  let dashboardResources: Resource[] = [];
  let hoveredResource: string | null = null;

  let selectedResource: Resource | null = null;
  let fileArtifact: FileArtifact | null = null;
  let workspace: ComponentType | null = null;

  function getResourceKind(resource: V1Resource): ResourceKind | string {
    if (resource.metricsView) return ResourceKind.MetricsView;
    if (resource.model) return ResourceKind.Model;
    if (resource.explore) return ResourceKind.Explore;
    if (resource.canvas) return ResourceKind.Canvas;
    if (resource.source) return ResourceKind.Source;
    if (resource.component) return ResourceKind.Component;
    if (resource.connector) return ResourceKind.Connector;
    return "Unknown";
  }

  $: resourcesQuery = createRuntimeServiceListResources(
    $runtime.instanceId,
    {},
  );

  $: {
    const rawResources = $resourcesQuery.data?.resources ?? [];

    allResources = rawResources
      .map((resource) => ({
        name: resource.meta?.name?.name || "Unknown",
        kind: getResourceKind(resource),
        state: resource.meta?.reconcileStatus || "UNKNOWN",
        error: resource.meta?.reconcileError,
        path: resource.meta?.filePaths?.[0] || "",
      }))
      .sort((a, b) => a.name.localeCompare(b.name));

    metricsResources = allResources.filter(
      (r) => r.kind === ResourceKind.MetricsView,
    );
    dashboardResources = allResources.filter(
      (r) => r.kind === ResourceKind.Canvas || r.kind === ResourceKind.Explore,
    );
  }

  $: loading = $resourcesQuery.isLoading;
  $: error = $resourcesQuery.isError
    ? ($resourcesQuery.error?.message ?? "Failed to load resources")
    : null;

  // Watch for query parameter changes and update selected resource
  $: {
    const resourceName = $page.url.searchParams.get("resource");
    if (resourceName && allResources.length) {
      const resource = allResources.find((r) => r.name === resourceName);
      if (resource && selectedResource?.name !== resource.name) {
        selectedResource = resource;
      }
    }
  }

  async function navigateToEditor(resource: Resource) {
    if (!resource.path) return;
    selectedResource = resource;

    // Update URL with clean query parameter
    const url = new URL($page.url);
    url.search = `?resource=${encodeURIComponent(resource.name)}`;
    await goto(url);
  }

  $: if (selectedResource && selectedResource.path) {
    fileArtifact = fileArtifacts.getFileArtifact(selectedResource.path);
    fileArtifact.fetchContent();

    // Force viz mode
    const ws = workspaceStore.get(selectedResource.path);
    ws.view.set("viz");
  } else {
    fileArtifact = null;
    workspace = null;
  }

  $: if (fileArtifact) {
    const kind = selectedResource?.kind as ResourceKind | undefined;
    workspace = workspaceComponents.get(kind) ?? null;
  }
</script>

<div class="h-full w-full flex" style="background: var(--surface-base)">
  <!-- Sidebar -->
  <div
    class="w-80 max-w-sm border-r flex flex-col"
    style="border-color: var(--border); background: var(--surface-subtle)"
  >
    <!-- Header -->
    <div class="p-4 border-b" style="border-color: var(--border)">
      <h2 class="text-sm font-semibold" style="color: var(--fg-primary)">
        {selectedResource?.name || "Resources"}
      </h2>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto">
      {#if error}
        <div
          class="p-4 bg-red-50 dark:bg-red-900 border-b border-red-200 dark:border-red-800"
        >
          <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      {/if}

      {#if loading && allResources.length === 0}
        <div class="p-4 text-center">
          <p class="text-sm" style="color: var(--fg-muted)">
            Loading resources...
          </p>
        </div>
      {:else if metricsResources.length > 0 || dashboardResources.length > 0}
        {#if metricsResources.length > 0}
          <ResourceListSection
            title="Metric Views"
            resources={metricsResources}
            expanded={$resourceSectionState.metrics}
            onToggle={() => resourceSectionState.toggle("metrics")}
            onSelect={navigateToEditor}
            {hoveredResource}
            onHover={(name) => (hoveredResource = name)}
          />
        {/if}

        {#if dashboardResources.length > 0}
          <ResourceListSection
            title="Dashboards"
            resources={dashboardResources}
            expanded={$resourceSectionState.dashboards}
            onToggle={() => resourceSectionState.toggle("dashboards")}
            onSelect={navigateToEditor}
            {hoveredResource}
            onHover={(name) => (hoveredResource = name)}
          />
        {/if}
      {:else}
        <!-- Empty State -->
        <div class="p-4 text-center">
          <p class="text-sm" style="color: var(--fg-muted)">No resources yet</p>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="border-t p-3 space-y-2" style="border-color: var(--border)">
      <div class="flex items-center justify-between gap-2">
        {#if selectedResource?.path}
          <span class="text-xs truncate" style="color: var(--fg-muted)">
            {selectedResource.path}
          </span>
        {/if}
        <button
          on:click={() => $resourcesQuery.refetch()}
          disabled={loading}
          class="flex-shrink-0 px-3 py-2 text-sm rounded transition-colors disabled:opacity-50 refresh-btn"
          style="color: var(--fg-secondary)"
        >
          {loading ? "Refreshing..." : "Refresh"}
        </button>
      </div>
    </div>
  </div>

  <!-- Main Content -->
  <div class="flex-1 overflow-hidden flex flex-col">
    {#if workspace && fileArtifact}
      <svelte:component this={workspace} {fileArtifact} hideCodeToggle={true} />
    {:else}
      <div
        class="flex items-center justify-center h-full"
        style="color: var(--fg-muted)"
      >
        <div class="text-center">
          <p class="text-lg font-medium mb-2" style="color: var(--fg-primary)">
            Edit Resources
          </p>
          <p class="text-sm" style="color: var(--fg-muted)">
            Click a resource from the sidebar to edit it
          </p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .refresh-btn:hover {
    background: var(--surface-subtle);
  }
</style>
