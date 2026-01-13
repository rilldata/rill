<script lang="ts">
  import { goto } from "$app/navigation";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Duration } from "luxon";
  import { onMount } from "svelte";

  interface Dashboard {
    name: string;
    title?: string;
    kind: "MetricsView" | "Canvas";
    lastRefreshed?: string;
    description?: string;
    filePath?: string;
    fullPath?: string;
  }

  let dashboards: Dashboard[] = [];
  let filteredDashboards: Dashboard[] = [];
  let searchQuery = "";
  let loading = false;
  let error: string | null = null;

  function timeAgo(date: Date): string {
    const now = DateTime.now();
    const then = DateTime.fromJSDate(date);
    const diff = Duration.fromMillis(now.diff(then).milliseconds);

    if (diff.as("minutes") < 1) return "Just now";

    const minutes = Math.round(diff.as("minutes"));
    if (diff.as("hours") < 1)
      return `${minutes} ${minutes === 1 ? "minute" : "minutes"} ago`;

    const hours = Math.round(diff.as("hours"));
    if (diff.as("days") < 1)
      return `${hours} ${hours === 1 ? "hour" : "hours"} ago`;

    const days = Math.round(diff.as("days"));
    if (diff.as("weeks") < 1)
      return `${days} ${days === 1 ? "day" : "days"} ago`;

    const weeks = Math.round(diff.as("weeks"));
    if (diff.as("months") < 1)
      return `${weeks} ${weeks === 1 ? "week" : "weeks"} ago`;

    const months = Math.round(diff.as("months"));
    if (diff.as("years") < 1)
      return `${months} ${months === 1 ? "month" : "months"} ago`;

    const years = Math.round(diff.as("years"));
    return `${years} ${years === 1 ? "year" : "years"} ago`;
  }

  function getDisplayName(dashboard: Dashboard): string {
    // Prefer title (displayName from API)
    if (dashboard.title) return dashboard.title;

    // Fallback to formatted name from file path or resource name
    let name = dashboard.name;
    if (dashboard.fullPath) {
      // Extract filename without extension and format it
      const filename = dashboard.fullPath.split('/').pop()?.split('.')[0] || dashboard.name;
      name = filename;
    }

    // Convert snake_case or kebab-case to Title Case
    return name
      .split(/[-_]/)
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  }

  function filterDashboards() {
    if (!searchQuery.trim()) {
      filteredDashboards = dashboards;
    } else {
      const query = searchQuery.toLowerCase();
      filteredDashboards = dashboards.filter((d) =>
        getDisplayName(d).toLowerCase().includes(query) || d.name.toLowerCase().includes(query)
      );
    }
  }

  $: if (searchQuery !== undefined) {
    filterDashboards();
  }

  async function loadDashboards() {
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
      dashboards = [];

      if (data?.resources) {
        for (const resource of data.resources) {
          const kind = resource.meta?.name?.kind;
          const name = resource.meta?.name?.name;

          // Check for Explore or Canvas resources
          if (kind === "rill.runtime.v1.Explore" || kind === "rill.runtime.v1.Canvas") {
            if (name) {
              // Extract metadata - displayName is in spec
              let title = "";
              if (kind === "rill.runtime.v1.Explore") {
                title = resource.explore?.spec?.displayName || "";
              } else if (kind === "rill.runtime.v1.Canvas") {
                title = resource.canvas?.spec?.displayName || "";
              }

              // Extract file path directory (e.g., "dashboards" from "dashboards/sales.yaml")
              let filePath = "";
              let fullPath = "";
              if (resource.meta?.filePaths?.[0]) {
                fullPath = resource.meta.filePaths[0];
                const parts = fullPath.split("/");
                if (parts.length > 1) {
                  filePath = parts[0]; // Get the directory name
                }
              }

              dashboards.push({
                name,
                title,
                kind: kind === "rill.runtime.v1.Explore" ? "MetricsView" : "Canvas",
                lastRefreshed: resource.meta?.stateUpdatedOn,
                filePath,
                fullPath,
                apiKind: kind,
              });
            }
          }
        }
      }

      // Sort alphabetically
      dashboards.sort((a, b) => a.name.localeCompare(b.name));

      // Apply filter to display
      filterDashboards();
    } catch (err) {
      error =
        err instanceof Error
          ? err.message
          : "Failed to load dashboards";
      console.error("Error loading dashboards:", err);
    } finally {
      loading = false;
    }
  }

  function navigateToDashboard(dashboard: Dashboard) {
    // MetricsView uses the explore view, Canvas uses the canvas view
    const dashboardSlug = dashboard.kind === "MetricsView" ? "explore" : "canvas";
    goto(`/${dashboardSlug}/${dashboard.name}`);
  }

  onMount(() => {
    loadDashboards();
  });

  // Retry when runtime becomes available
  $: if ($runtime?.instanceId && $runtime?.host && error?.includes("Waiting")) {
    loadDashboards();
  }
</script>

<div class="h-full w-full flex flex-col bg-white dark:bg-gray-950 overflow-hidden">
  <div class="flex-1 overflow-auto flex flex-col items-center">
    <div class="w-full max-w-6xl p-8">
      <!-- Header -->
      <div class="mb-6">
        <div class="mb-4">
          <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-1">
            Dashboards
          </h2>
          <p class="text-sm text-gray-600 dark:text-gray-400">
            {dashboards.length} dashboard{dashboards.length !== 1 ? "s" : ""} available
          </p>
        </div>

        <!-- Search Bar -->
        {#if dashboards.length > 0}
          <input
            type="text"
            placeholder="Search dashboards..."
            bind:value={searchQuery}
            class="w-full px-4 py-2 border border-gray-200 dark:border-gray-800 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        {/if}
      </div>

      <!-- Error Message -->
      {#if error}
        <div class="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded p-4 mb-6">
          <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      {/if}

      <!-- Loading State -->
      {#if loading && dashboards.length === 0}
        <div class="flex items-center justify-center py-12">
          <p class="text-gray-500 dark:text-gray-400">Loading dashboards...</p>
        </div>
      {:else if dashboards.length === 0}
        <!-- Empty State -->
        <div class="flex items-center justify-center py-12 border-2 border-dashed border-gray-200 dark:border-gray-800 rounded">
          <div class="text-center">
            <p class="text-lg font-medium text-gray-900 dark:text-white mb-2">
              No dashboards found
            </p>
            <p class="text-sm text-gray-600 dark:text-gray-400">
              Create a metrics view or dashboard to get started
            </p>
          </div>
        </div>
      {:else if filteredDashboards.length === 0}
        <!-- No Results from Search -->
        <div class="flex items-center justify-center py-12 border-2 border-dashed border-gray-200 dark:border-gray-800 rounded">
          <div class="text-center">
            <p class="text-lg font-medium text-gray-900 dark:text-white mb-2">
              No dashboards match "{searchQuery}"
            </p>
            <p class="text-sm text-gray-600 dark:text-gray-400">
              Try a different search term
            </p>
          </div>
        </div>
      {:else}
        <!-- Dashboard List -->
        <div class="space-y-0 w-full border border-gray-200 dark:border-gray-800 rounded divide-y divide-gray-200 dark:divide-gray-800 overflow-hidden">
          {#each filteredDashboards as dashboard (dashboard.name)}
            <button
              on:click={() => navigateToDashboard(dashboard)}
              class="flex flex-col gap-y-1 group px-4 py-2.5 w-full hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors text-left"
            >
              <!-- Top row: Icon, Display Name -->
              <div class="flex gap-x-2 items-center min-h-[20px]">
                <ResourceTypeBadge
                  kind={dashboard.kind === "MetricsView"
                    ? ResourceKind.Explore
                    : ResourceKind.Canvas}
                />
                <span class="text-gray-700 dark:text-gray-100 text-sm font-semibold group-hover:text-primary-600 dark:group-hover:text-primary-400 truncate">
                  {getDisplayName(dashboard)}
                </span>
              </div>

              <!-- Bottom row: File Path on left, Last Refreshed and Description on right -->
              <div class="flex gap-x-1 text-gray-500 dark:text-gray-400 text-xs font-normal min-h-[16px] overflow-hidden">
                <span class="shrink-0">
                  {dashboard.fullPath || dashboard.name}
                </span>
                {#if dashboard.lastRefreshed}
                  <span class="shrink-0">•</span>
                  <Tooltip distance={8}>
                    <span class="shrink-0 truncate">
                      Last refreshed {timeAgo(new Date(dashboard.lastRefreshed))}
                    </span>
                    <TooltipContent slot="tooltip-content">
                      {new Date(dashboard.lastRefreshed).toLocaleString()}
                    </TooltipContent>
                  </Tooltip>
                {/if}
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>
