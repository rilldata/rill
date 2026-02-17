<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { workspaces as workspaceStore } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import CanvasWorkspace from "@rilldata/web-common/features/workspaces/CanvasWorkspace.svelte";
  import ExploreWorkspace from "@rilldata/web-common/features/workspaces/ExploreWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import { resourceSectionState } from "./resource-section-store";

  interface Resource {
    name: string;
    kind: ResourceKind | string;
    state?: string;
    error?: string;
    path?: string;
  }

  const workspaceComponents = new Map([
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
  let loading = false;
  let error: string | null = null;
  let hoveredResource: string | null = null;

  let selectedResource: Resource | null = null;
  let fileArtifact: any = null;
  let workspace: any = null;

  async function loadResources() {
    try {
      loading = true;
      error = null;

      if (!$runtime?.instanceId || $runtime?.host === undefined) {
        error = "Waiting for runtime to initialize...";
        loading = false;
        return;
      }

      // Fetch the list of resources from the runtime
      const response = await fetch(
        `${$runtime.host}/v1/instances/${$runtime.instanceId}/resources`,
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch resources: ${response.statusText}`);
      }

      const data = await response.json();

      // Collect and organize all resources
      allResources = [];

      if (data?.resources) {
        for (const resource of Object.values(data.resources || {}) as any[]) {
          const resourceName = resource.meta?.name?.name || "Unknown";
          const filePath = resource.meta?.filePaths?.[0] || "";

          // Detect resource type from which property exists
          let kind = "Unknown";
          if (resource.metricsView) kind = ResourceKind.MetricsView;
          else if (resource.model) kind = ResourceKind.Model;
          else if (resource.explore) kind = ResourceKind.Explore;
          else if (resource.canvas) kind = ResourceKind.Canvas;
          else if (resource.source) kind = ResourceKind.Source;
          else if (resource.component) kind = ResourceKind.Component;
          else if (resource.connector) kind = ResourceKind.Connector;

          allResources.push({
            name: resourceName,
            kind: kind,
            state: resource.meta?.reconcileStatus || "UNKNOWN",
            error: resource.meta?.reconcileError,
            path: filePath,
          });
        }
      }

      // Sort alphabetically
      allResources.sort((a, b) => (a.name || "").localeCompare(b.name || ""));

      // Group by type
      metricsResources = allResources.filter(
        (r) => r.kind === ResourceKind.MetricsView,
      );
      dashboardResources = allResources.filter(
        (r) =>
          r.kind === ResourceKind.Canvas || r.kind === ResourceKind.Explore,
      );
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load resources";
      console.error("Error loading resources:", err);
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    await loadResources();

    // Restore selected resource from URL query parameter
    const resourceName = $page.url.searchParams.get("resource");
    if (resourceName) {
      const resource = allResources.find((r) => r.name === resourceName);
      if (resource) {
        selectedResource = resource;

        // Clean URL if there are other parameters
        if (
          $page.url.search !== `?resource=${encodeURIComponent(resourceName)}`
        ) {
          const url = new URL($page.url);
          url.search = `?resource=${encodeURIComponent(resourceName)}`;
          goto(url, { replaceState: true });
        }
      }
    }
  });

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

  // Retry when runtime becomes available
  $: if ($runtime?.instanceId && $runtime?.host && error?.includes("Waiting")) {
    loadResources();
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
    workspace = workspaceComponents.get(kind);
  }

  function getStatusText(resource: Resource): string {
    const state = (resource.state || "").toUpperCase();
    if (resource.error) return "Error";
    if (state === "RECONCILING" || state === "COMPILING") return "Compiling";
    if (state === "OK") return "Ready";
    return state.charAt(0) + state.slice(1).toLowerCase();
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
        <!-- Metrics Section -->
        {#if metricsResources.length > 0}
          <div class="py-2">
            <button
              on:click={() => resourceSectionState.toggle("metrics")}
              class="w-full flex items-center gap-1 px-3 py-1.5 transition-colors section-toggle"
            >
              <div
                class="flex-shrink-0 w-4 h-4 flex items-center justify-center"
              >
                <CaretDownIcon
                  size="12px"
                  className={`transition-transform ${!$resourceSectionState.metrics ? "-rotate-90" : ""}`}
                />
              </div>
              <h3
                class="text-xs font-semibold uppercase tracking-wide"
                style="color: var(--fg-muted)"
              >
                Metric Views ({metricsResources.length})
              </h3>
            </button>

            {#if $resourceSectionState.metrics}
              <ul class="list-none p-0 m-0 w-full">
                {#each metricsResources as resource, idx (resource.name)}
                  <li
                    class={`block w-full border transition-colors resource-row relative ${
                      idx === 0 ? "rounded-t-lg" : "border-t-0"
                    } ${
                      idx === metricsResources.length - 1 ? "rounded-b-lg" : ""
                    } ${resource.error ? "border-l-4 border-l-red-600" : ""}`}
                    style="border-color: var(--border)"
                    on:mouseenter={() => (hoveredResource = resource.name)}
                    on:mouseleave={() => (hoveredResource = null)}
                  >
                    <button
                      on:click={() => navigateToEditor(resource)}
                      class="flex items-center gap-x-3 group px-4 py-3 w-full"
                    >
                      <!-- Icon Container -->
                      <div
                        class="flex-shrink-0 h-10 w-10 rounded-md flex items-center justify-center"
                        style="background: var(--surface-subtle)"
                      >
                        <svelte:component
                          this={resourceIconMapping[resource.kind]}
                          size="20px"
                          color="var(--fg-secondary)"
                        />
                      </div>

                      <!-- Content -->
                      <div class="flex-1 min-w-0">
                        <div
                          class={`text-sm font-semibold truncate ${
                            resource.error
                              ? "text-red-600 dark:text-red-400"
                              : "resource-name"
                          }`}
                          style:color={!resource.error
                            ? "var(--fg-secondary)"
                            : undefined}
                        >
                          {resource.name}
                        </div>
                        <div
                          class="text-xs truncate"
                          style="color: var(--fg-muted)"
                        >
                          {resource.path || "No path"}
                        </div>
                      </div>

                      <!-- Status Circle -->
                      <div class="flex-shrink-0 flex items-center gap-x-2">
                        <div
                          class="h-2.5 w-2.5 rounded-full"
                          style={`background-color: ${
                            resource.error
                              ? "#DC2626"
                              : resource.state?.toUpperCase() ===
                                    "RECONCILING" ||
                                  resource.state?.toUpperCase() === "COMPILING"
                                ? "#F59E0B"
                                : "#10B981"
                          }`}
                          title={getStatusText(resource)}
                        />
                      </div>
                    </button>

                    <!-- Error Tooltip -->
                    {#if hoveredResource === resource.name && resource.error}
                      <div
                        class="absolute left-0 top-full mt-1 z-50 bg-red-600 dark:bg-red-700 text-white text-xs rounded px-2 py-1 max-w-xs break-words"
                      >
                        {resource.error}
                      </div>
                    {/if}
                  </li>
                {/each}
              </ul>
            {/if}
          </div>
        {/if}

        <!-- Dashboards Section -->
        {#if dashboardResources.length > 0}
          <div class="py-2">
            <button
              on:click={() => resourceSectionState.toggle("dashboards")}
              class="w-full flex items-center gap-1 px-3 py-1.5 transition-colors section-toggle"
            >
              <div
                class="flex-shrink-0 w-4 h-4 flex items-center justify-center"
              >
                <CaretDownIcon
                  size="12px"
                  className={`transition-transform ${!$resourceSectionState.dashboards ? "-rotate-90" : ""}`}
                />
              </div>
              <h3
                class="text-xs font-semibold uppercase tracking-wide"
                style="color: var(--fg-muted)"
              >
                Dashboards ({dashboardResources.length})
              </h3>
            </button>

            {#if $resourceSectionState.dashboards}
              <ul class="list-none p-0 m-0 w-full">
                {#each dashboardResources as resource, idx (resource.name)}
                  <li
                    class={`block w-full border transition-colors resource-row relative ${
                      idx === 0 ? "rounded-t-lg" : "border-t-0"
                    } ${
                      idx === dashboardResources.length - 1
                        ? "rounded-b-lg"
                        : ""
                    } ${resource.error ? "border-l-4 border-l-red-600" : ""}`}
                    style="border-color: var(--border)"
                    on:mouseenter={() => (hoveredResource = resource.name)}
                    on:mouseleave={() => (hoveredResource = null)}
                  >
                    <button
                      on:click={() => navigateToEditor(resource)}
                      class="flex items-center gap-x-3 group px-4 py-3 w-full"
                    >
                      <!-- Icon Container -->
                      <div
                        class="flex-shrink-0 h-10 w-10 rounded-md flex items-center justify-center"
                        style="background: var(--surface-subtle)"
                      >
                        <svelte:component
                          this={resourceIconMapping[resource.kind]}
                          size="20px"
                          color="var(--fg-secondary)"
                        />
                      </div>

                      <!-- Content -->
                      <div class="flex-1 min-w-0">
                        <div
                          class={`text-sm font-semibold truncate ${
                            resource.error
                              ? "text-red-600 dark:text-red-400"
                              : "resource-name"
                          }`}
                          style:color={!resource.error
                            ? "var(--fg-secondary)"
                            : undefined}
                        >
                          {resource.name}
                        </div>
                        <div
                          class="text-xs truncate"
                          style="color: var(--fg-muted)"
                        >
                          {resource.path || "No path"}
                        </div>
                      </div>

                      <!-- Status Circle -->
                      <div class="flex-shrink-0 flex items-center gap-x-2">
                        <div
                          class="h-2.5 w-2.5 rounded-full"
                          style={`background-color: ${
                            resource.error
                              ? "#DC2626"
                              : resource.state?.toUpperCase() ===
                                    "RECONCILING" ||
                                  resource.state?.toUpperCase() === "COMPILING"
                                ? "#F59E0B"
                                : "#10B981"
                          }`}
                          title={getStatusText(resource)}
                        />
                      </div>
                    </button>

                    <!-- Error Tooltip -->
                    {#if hoveredResource === resource.name && resource.error}
                      <div
                        class="absolute left-0 top-full mt-1 z-50 bg-red-600 dark:bg-red-700 text-white text-xs rounded px-2 py-1 max-w-xs break-words"
                      >
                        {resource.error}
                      </div>
                    {/if}
                  </li>
                {/each}
              </ul>
            {/if}
          </div>
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
          on:click={loadResources}
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
  .section-toggle:hover {
    background: var(--surface-subtle);
  }

  .resource-row:hover {
    background: var(--surface-subtle);
  }

  .resource-name {
    color: var(--fg-secondary);
  }

  :global(.group):hover .resource-name {
    color: var(--fg-primary);
  }

  .refresh-btn:hover {
    background: var(--surface-subtle);
  }
</style>
