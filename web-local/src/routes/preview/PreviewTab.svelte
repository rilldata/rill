<script lang="ts">
  import { goto } from "$app/navigation";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client/gen/runtime-service/runtime-service";
  import { DateTime, Duration } from "luxon";

  interface Dashboard {
    name: string;
    title?: string;
    kind: "MetricsView" | "Canvas";
    lastRefreshed?: string;
    description?: string;
    filePath?: string;
    fullPath?: string;
    hasError?: boolean;
    errorMessage?: string;
  }

  let searchQuery = "";

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

  // Reactive query - automatically refetches when resources change
  $: resourcesQuery = createRuntimeServiceListResources($runtime.instanceId, {});

  // Transform resources into dashboards (exclude MetricsView resources)
  $: dashboards = ($resourcesQuery.data?.resources ?? [])
    .filter((resource) => {
      const kind = resource.meta?.name?.kind;
      // Only show Explore and Canvas, not MetricsView
      return kind === "rill.runtime.v1.Explore" || kind === "rill.runtime.v1.Canvas";
    })
    .map((resource) => {
      const kind = resource.meta?.name?.kind;
      const name = resource.meta?.name?.name || "";

      let title = "";
      if (kind === "rill.runtime.v1.Explore") {
        title = resource.explore?.spec?.displayName || "";
      } else if (kind === "rill.runtime.v1.Canvas") {
        title = resource.canvas?.spec?.displayName || "";
      }

      let filePath = "";
      let fullPath = "";
      if (resource.meta?.filePaths?.[0]) {
        fullPath = resource.meta.filePaths[0];
        const parts = fullPath.split("/");
        if (parts.length > 1) {
          filePath = parts[0];
        }
      }

      return {
        name,
        title,
        kind: kind === "rill.runtime.v1.Explore" ? "MetricsView" : "Canvas",
        lastRefreshed: resource.meta?.stateUpdatedOn,
        filePath,
        fullPath,
        hasError: !!resource.meta?.reconcileError,
        errorMessage: resource.meta?.reconcileError,
      } as Dashboard;
    })
    .sort((a, b) => a.name.localeCompare(b.name));

  // Filter dashboards based on search
  $: filteredDashboards = searchQuery.trim()
    ? dashboards.filter((d) =>
        getDisplayName(d).toLowerCase().includes(searchQuery.toLowerCase()) ||
        d.name.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : dashboards;

  function navigateToDashboard(dashboard: Dashboard) {
    // Don't navigate if dashboard has errors
    if (dashboard.hasError) return;

    // MetricsView uses the explore view, Canvas uses the canvas view
    const dashboardSlug = dashboard.kind === "MetricsView" ? "explore" : "canvas";
    goto(`/${dashboardSlug}/${dashboard.name}`);
  }
</script>

<div class="h-full w-full flex flex-col overflow-hidden" style="background: var(--surface-base)">
  <div class="flex-1 overflow-auto flex flex-col items-center">
    <div class="w-full max-w-6xl p-8">
      <!-- Header -->
      <div class="mb-6">
        <div class="mb-4">
          <h2 class="text-2xl font-semibold mb-1" style="color: var(--fg-primary)">
            Dashboards
          </h2>
          <p class="text-sm" style="color: var(--fg-secondary)">
            {dashboards.length} dashboard{dashboards.length !== 1 ? "s" : ""} available
          </p>
        </div>

        <!-- Search Bar -->
        {#if dashboards.length > 0}
          <input
            type="text"
            placeholder="Search dashboards..."
            bind:value={searchQuery}
            class="w-full px-4 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            style="border-color: var(--border); background: var(--surface-subtle); color: var(--fg-primary)"
          />
        {/if}
      </div>

      <!-- Error Message -->
      {#if $resourcesQuery.error}
        <div class="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded p-4 mb-6">
          <p class="text-sm text-red-600 dark:text-red-400">{$resourcesQuery.error.message || "Failed to load dashboards"}</p>
        </div>
      {/if}

      <!-- Loading State -->
      {#if $resourcesQuery.isLoading && dashboards.length === 0}
        <div class="flex items-center justify-center py-12">
          <p style="color: var(--fg-muted)">Loading dashboards...</p>
        </div>
      {:else if dashboards.length === 0}
        <!-- Empty State -->
        <div class="flex items-center justify-center py-12 border-2 border-dashed rounded" style="border-color: var(--border)">
          <div class="text-center">
            <p class="text-lg font-medium mb-2" style="color: var(--fg-primary)">
              No dashboards found
            </p>
            <p class="text-sm" style="color: var(--fg-secondary)">
              Create a metrics view or dashboard to get started
            </p>
          </div>
        </div>
      {:else if filteredDashboards.length === 0}
        <!-- No Results from Search -->
        <div class="flex items-center justify-center py-12 border-2 border-dashed rounded" style="border-color: var(--border)">
          <div class="text-center">
            <p class="text-lg font-medium mb-2" style="color: var(--fg-primary)">
              No dashboards match "{searchQuery}"
            </p>
            <p class="text-sm" style="color: var(--fg-secondary)">
              Try a different search term
            </p>
          </div>
        </div>
      {:else}
        <!-- Dashboard List -->
        <div class="space-y-0 w-full border rounded overflow-hidden" style="border-color: var(--border)">
          {#each filteredDashboards as dashboard, i (dashboard.name)}
            <Tooltip distance={4} alignment="start" suppress={!dashboard.hasError}>
              <button
                type="button"
                class:border-t={i > 0}
                style:border-color={i > 0 ? 'var(--border)' : undefined}
                on:click={() => navigateToDashboard(dashboard)}
                disabled={dashboard.hasError}
                class="flex flex-col gap-y-1 group px-4 py-2.5 w-full transition-colors text-left dashboard-row"
                class:hoverable={!dashboard.hasError}
                class:cursor-pointer={!dashboard.hasError}
                class:opacity-60={dashboard.hasError}
              >
              <!-- Top row: Icon, Display Name -->
              <div class="flex gap-x-2 items-center min-h-[20px]">
                <ResourceTypeBadge
                  kind={dashboard.kind === "MetricsView"
                    ? ResourceKind.Explore
                    : ResourceKind.Canvas}
                />
                <span class="text-sm font-semibold truncate dashboard-name" class:text-red-600={dashboard.hasError} class:dark:text-red-400={dashboard.hasError} class:group-hover:text-primary-600={!dashboard.hasError} class:dark:group-hover:text-primary-400={!dashboard.hasError} style:color={!dashboard.hasError ? 'var(--fg-primary)' : undefined}>
                  {getDisplayName(dashboard)}
                </span>
                {#if dashboard.hasError}
                  <span class="text-xs px-1.5 py-0.5 rounded bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-400 font-medium">
                    Error
                  </span>
                {/if}
              </div>

              <!-- Bottom row: File Path on left, Last Refreshed and Description on right -->
              <div class="flex gap-x-1 text-xs font-normal min-h-[16px] overflow-hidden" class:text-red-500={dashboard.hasError} class:dark:text-red-400={dashboard.hasError} style:color={!dashboard.hasError ? 'var(--fg-muted)' : undefined}>
                <span class="shrink-0">
                  {dashboard.fullPath || dashboard.name}
                </span>
                {#if !dashboard.hasError && dashboard.lastRefreshed}
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
              <TooltipContent slot="tooltip-content">
                {dashboard.errorMessage || "Dashboard has errors"}
              </TooltipContent>
            </Tooltip>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .dashboard-row.hoverable:hover {
    background: var(--surface-subtle);
  }
</style>
