<script lang="ts">
  import { goto } from "$app/navigation";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    resourceIconMapping,
    resourceColorMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
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

  let allResources: Resource[] = [];
  let metricsResources: Resource[] = [];
  let exploreResources: Resource[] = [];
  let dashboardResources: Resource[] = [];
  let loading = false;
  let error: string | null = null;
  let hoveredResource: string | null = null;
  let contextMenuResource: string | null = null;
  let contextMenuPos: { x: number; y: number } | null = null;

  async function loadResources() {
    try {
      loading = true;
      error = null;

      if (!$runtime?.instanceId || !$runtime?.host) {
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
        for (const resource of Object.values(data.resources || {})) {
          const resourceName = resource.meta?.name?.name || "Unknown";
          const filePath = resource.meta?.file_paths?.[0] || "";

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
      metricsResources = allResources.filter((r) => r.kind === ResourceKind.MetricsView);
      exploreResources = allResources.filter((r) => r.kind === ResourceKind.Model);
      dashboardResources = allResources.filter((r) => r.kind === ResourceKind.Canvas || r.kind === ResourceKind.Explore);
    } catch (err) {
      error =
        err instanceof Error
          ? err.message
          : "Failed to load resources";
      console.error("Error loading resources:", err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadResources();
  });

  // Retry when runtime becomes available
  $: if ($runtime?.instanceId && $runtime?.host && error?.includes("Waiting")) {
    loadResources();
  }

  function navigateToEditor(resource: Resource) {
    if (!resource.path) return;
    goto(`/edit${resource.path}`);
  }

  function getStatusIcon(resource: Resource): string {
    const state = (resource.state || "").toUpperCase();
    if (resource.error) return "⚠";
    if (state === "RECONCILING" || state === "COMPILING") return "⟳";
    return "✓";
  }

  function getStatusColor(resource: Resource): string {
    const state = (resource.state || "").toUpperCase();
    if (resource.error) return "text-red-600 dark:text-red-400";
    if (state === "RECONCILING" || state === "COMPILING") return "text-yellow-600 dark:text-yellow-400";
    return "text-green-600 dark:text-green-400";
  }

  function getStatusText(resource: Resource): string {
    const state = (resource.state || "").toUpperCase();
    if (resource.error) return "Error";
    if (state === "RECONCILING" || state === "COMPILING") return "Compiling";
    if (state === "OK") return "Ready";
    return state.charAt(0) + state.slice(1).toLowerCase();
  }

  function handleContextMenu(e: MouseEvent, resourceName: string) {
    e.preventDefault();
    contextMenuResource = resourceName;
    contextMenuPos = { x: e.clientX, y: e.clientY };
  }

  function closeContextMenu() {
    contextMenuResource = null;
    contextMenuPos = null;
  }

  function handleDeleteResource(resourceName: string) {
    console.log("Delete resource:", resourceName);
    closeContextMenu();
  }

  function handleRefreshResource(resourceName: string) {
    console.log("Refresh resource:", resourceName);
    closeContextMenu();
  }

  function handleRenameResource(resourceName: string) {
    console.log("Rename resource:", resourceName);
    closeContextMenu();
  }
</script>

<svelte:window on:click={closeContextMenu} />

<div class="h-full w-full flex">
  <!-- Sidebar -->
  <div class="w-80 max-w-sm border-r border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 flex flex-col">
    <!-- Header -->
    <div class="p-4 border-b border-gray-200 dark:border-gray-800">
      <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Resources</h2>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto">
      {#if error}
        <div class="p-4 bg-red-50 dark:bg-red-900 border-b border-red-200 dark:border-red-800">
          <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      {/if}

      {#if loading && allResources.length === 0}
        <div class="p-4 text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">Loading resources...</p>
        </div>
      {:else if metricsResources.length > 0 || dashboardResources.length > 0}
        <!-- Metrics Section -->
        {#if metricsResources.length > 0}
          <div class="py-2">
            <button
              on:click={() => resourceSectionState.toggle("metrics")}
              class="w-full flex items-center gap-1 px-3 py-1.5 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              <div class="flex-shrink-0 w-4 h-4 flex items-center justify-center">
                <CaretDownIcon
                  size="12px"
                  class={`transition-transform ${!$resourceSectionState.metrics ? "-rotate-90" : ""}`}
                />
              </div>
              <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-400 uppercase tracking-wide">
                Metric Views ({metricsResources.length})
              </h3>
            </button>

            {#if $resourceSectionState.metrics}
              <ul class="list-none p-0 m-0 w-full">
              {#each metricsResources as resource, idx (resource.name)}
                <li
                  class={`block w-full border border-gray-200 dark:border-gray-700 transition-colors hover:bg-slate-50 dark:hover:bg-slate-800 relative ${
                    idx === 0 ? "rounded-t-lg" : "border-t-0"
                  } ${
                    idx === metricsResources.length - 1 ? "rounded-b-lg" : ""
                  } ${
                    resource.error ? "border-l-4 border-l-red-600" : ""
                  }`}
                  on:mouseenter={() => (hoveredResource = resource.name)}
                  on:mouseleave={() => (hoveredResource = null)}
                >
                  <button
                    on:click={() => navigateToEditor(resource)}
                    on:contextmenu={(e) => handleContextMenu(e, resource.name)}
                    class="flex items-center gap-x-3 group px-4 py-3 w-full"
                  >
                    <!-- Icon Container -->
                    <div class="flex-shrink-0 h-10 w-10 rounded-md flex items-center justify-center" style={`background-color: ${resourceColorMapping[resource.kind]}20`}>
                      <svelte:component
                        this={resourceIconMapping[resource.kind]}
                        size="20px"
                        color={resourceColorMapping[resource.kind]}
                      />
                    </div>

                    <!-- Content -->
                    <div class="flex-1 min-w-0">
                      <div class={`text-sm font-semibold truncate ${
                        resource.error
                          ? "text-red-600 dark:text-red-400"
                          : "text-gray-700 dark:text-gray-300 group-hover:text-primary-600 dark:group-hover:text-primary-400"
                      }`}>
                        {resource.name}
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
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
                            : (resource.state?.toUpperCase() === "RECONCILING" || resource.state?.toUpperCase() === "COMPILING"
                              ? "#F59E0B"
                              : "#10B981")
                        }`}
                        title={getStatusText(resource)}
                      />
                    </div>
                  </button>

                  <!-- Error Tooltip -->
                  {#if hoveredResource === resource.name && resource.error}
                    <div class="absolute left-0 top-full mt-1 z-50 bg-red-600 dark:bg-red-700 text-white text-xs rounded px-2 py-1 max-w-xs break-words">
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
              class="w-full flex items-center gap-1 px-3 py-1.5 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              <div class="flex-shrink-0 w-4 h-4 flex items-center justify-center">
                <CaretDownIcon
                  size="12px"
                  class={`transition-transform ${!$resourceSectionState.dashboards ? "-rotate-90" : ""}`}
                />
              </div>
              <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-400 uppercase tracking-wide">
                Dashboards ({dashboardResources.length})
              </h3>
            </button>

            {#if $resourceSectionState.dashboards}
              <ul class="list-none p-0 m-0 w-full">
              {#each dashboardResources as resource, idx (resource.name)}
                <li
                  class={`block w-full border border-gray-200 dark:border-gray-700 transition-colors hover:bg-slate-50 dark:hover:bg-slate-800 relative ${
                    idx === 0 ? "rounded-t-lg" : "border-t-0"
                  } ${
                    idx === dashboardResources.length - 1 ? "rounded-b-lg" : ""
                  } ${
                    resource.error ? "border-l-4 border-l-red-600" : ""
                  }`}
                  on:mouseenter={() => (hoveredResource = resource.name)}
                  on:mouseleave={() => (hoveredResource = null)}
                >
                  <button
                    on:click={() => navigateToEditor(resource)}
                    on:contextmenu={(e) => handleContextMenu(e, resource.name)}
                    class="flex items-center gap-x-3 group px-4 py-3 w-full"
                  >
                    <!-- Icon Container -->
                    <div class="flex-shrink-0 h-10 w-10 rounded-md flex items-center justify-center" style={`background-color: ${resourceColorMapping[resource.kind]}20`}>
                      <svelte:component
                        this={resourceIconMapping[resource.kind]}
                        size="20px"
                        color={resourceColorMapping[resource.kind]}
                      />
                    </div>

                    <!-- Content -->
                    <div class="flex-1 min-w-0">
                      <div class={`text-sm font-semibold truncate ${
                        resource.error
                          ? "text-red-600 dark:text-red-400"
                          : "text-gray-700 dark:text-gray-300 group-hover:text-primary-600 dark:group-hover:text-primary-400"
                      }`}>
                        {resource.name}
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
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
                            : (resource.state?.toUpperCase() === "RECONCILING" || resource.state?.toUpperCase() === "COMPILING"
                              ? "#F59E0B"
                              : "#10B981")
                        }`}
                        title={getStatusText(resource)}
                      />
                    </div>
                  </button>

                  <!-- Error Tooltip -->
                  {#if hoveredResource === resource.name && resource.error}
                    <div class="absolute left-0 top-full mt-1 z-50 bg-red-600 dark:bg-red-700 text-white text-xs rounded px-2 py-1 max-w-xs break-words">
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
          <p class="text-sm text-gray-500 dark:text-gray-400">No resources yet</p>
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="border-t border-gray-200 dark:border-gray-800 p-3 space-y-2">
      <button
        on:click={loadResources}
        disabled={loading}
        class="w-full px-3 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors disabled:opacity-50"
      >
        {loading ? "Refreshing..." : "Refresh"}
      </button>
    </div>
  </div>

  <!-- Main Content -->
  <div class="flex-1 overflow-hidden flex flex-col items-center justify-center text-gray-500 dark:text-gray-400">
    <div class="text-center">
      <p class="text-lg font-medium text-gray-900 dark:text-white mb-2">
        Edit Resources
      </p>
      <p class="text-sm text-gray-600 dark:text-gray-400">
        Click a resource from the sidebar to edit it
      </p>
    </div>
  </div>

  <!-- Context Menu -->
  {#if contextMenuResource && contextMenuPos}
    <div
      class="fixed z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-lg"
      style={`left: ${contextMenuPos.x}px; top: ${contextMenuPos.y}px;`}
    >
      <button
        on:click={() => handleRefreshResource(contextMenuResource)}
        class="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 border-b border-gray-100 dark:border-gray-700"
      >
        Refresh
      </button>
      <button
        on:click={() => handleRenameResource(contextMenuResource)}
        class="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 border-b border-gray-100 dark:border-gray-700"
      >
        Rename
      </button>
      <button
        on:click={() => handleDeleteResource(contextMenuResource)}
        class="w-full text-left px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
      >
        Delete
      </button>
    </div>
  {/if}
</div>
