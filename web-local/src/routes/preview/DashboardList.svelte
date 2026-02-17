<script lang="ts">
  import { goto } from "$app/navigation";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client/gen/runtime-service/runtime-service";

  export let limit: number | undefined = undefined;
  export let showSearch = false;
  export let showSeeAll = false;
  export let seeAllHref = "/preview";

  interface Dashboard {
    name: string;
    title?: string;
    kind: "MetricsView" | "Canvas";
    lastRefreshed?: string;
    fullPath?: string;
    hasError?: boolean;
    errorMessage?: string;
  }

  let searchQuery = "";

  function getDisplayName(dashboard: Dashboard): string {
    if (dashboard.title) return dashboard.title;

    let name = dashboard.name;
    if (dashboard.fullPath) {
      const filename =
        dashboard.fullPath.split("/").pop()?.split(".")[0] || dashboard.name;
      name = filename;
    }

    return name
      .split(/[-_]/)
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(" ");
  }

  $: resourcesQuery = createRuntimeServiceListResources($runtime.instanceId, {});

  $: dashboards = ($resourcesQuery.data?.resources ?? [])
    .filter((resource) => {
      const kind = resource.meta?.name?.kind;
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

      const fullPath = resource.meta?.filePaths?.[0] || "";

      return {
        name,
        title,
        kind: kind === "rill.runtime.v1.Explore" ? "MetricsView" : "Canvas",
        lastRefreshed: resource.meta?.stateUpdatedOn,
        fullPath,
        hasError: !!resource.meta?.reconcileError,
        errorMessage: resource.meta?.reconcileError,
      } as Dashboard;
    })
    .sort((a, b) => a.name.localeCompare(b.name));

  $: filteredDashboards = searchQuery.trim()
    ? dashboards.filter(
        (d) =>
          getDisplayName(d).toLowerCase().includes(searchQuery.toLowerCase()) ||
          d.name.toLowerCase().includes(searchQuery.toLowerCase()),
      )
    : dashboards;

  $: displayData = limit ? filteredDashboards.slice(0, limit) : filteredDashboards;
  $: hasMore = limit ? dashboards.length > limit : false;

  function navigateToDashboard(dashboard: Dashboard) {
    if (dashboard.hasError) return;
    const dashboardSlug = dashboard.kind === "MetricsView" ? "explore" : "canvas";
    goto(`/${dashboardSlug}/${dashboard.name}`);
  }
</script>

<div class="flex flex-col w-full gap-y-3">
  {#if showSearch && dashboards.length > 0}
    <input
      type="text"
      placeholder="Search dashboards..."
      bind:value={searchQuery}
      class="w-full px-4 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
      style="border-color: var(--border); background: var(--surface-subtle); color: var(--fg-primary)"
    />
  {/if}

  {#if $resourcesQuery.isLoading && dashboards.length === 0}
    <div class="flex items-center justify-center py-8">
      <p style="color: var(--fg-muted)">Loading dashboards...</p>
    </div>
  {:else if dashboards.length === 0}
    <div
      class="flex items-center justify-center py-8 border-2 border-dashed rounded"
      style="border-color: var(--border)"
    >
      <div class="text-center">
        <p class="text-base font-medium mb-1" style="color: var(--fg-primary)">
          No dashboards found
        </p>
        <p class="text-sm" style="color: var(--fg-secondary)">
          Create a metrics view or dashboard to get started
        </p>
      </div>
    </div>
  {:else if filteredDashboards.length === 0}
    <div
      class="flex items-center justify-center py-8 border-2 border-dashed rounded"
      style="border-color: var(--border)"
    >
      <div class="text-center">
        <p class="text-base font-medium mb-1" style="color: var(--fg-primary)">
          No dashboards match "{searchQuery}"
        </p>
        <p class="text-sm" style="color: var(--fg-secondary)">
          Try a different search term
        </p>
      </div>
    </div>
  {:else}
    <div
      class="space-y-0 w-full border rounded overflow-hidden"
      style="border-color: var(--border)"
    >
      {#each displayData as dashboard, i (dashboard.name)}
        <Tooltip distance={4} alignment="start" suppress={!dashboard.hasError}>
          <div
            class:border-t={i > 0}
            style:border-color={i > 0 ? "var(--border)" : undefined}
            on:click={() => navigateToDashboard(dashboard)}
            on:keydown={(e) => e.key === "Enter" && navigateToDashboard(dashboard)}
            role={dashboard.hasError ? undefined : "button"}
            tabindex={dashboard.hasError ? -1 : 0}
            class="flex flex-col gap-y-1 group px-4 py-2.5 w-full transition-colors text-left dashboard-row"
            class:hoverable={!dashboard.hasError}
            class:cursor-pointer={!dashboard.hasError}
            class:opacity-60={dashboard.hasError}
          >
            <div class="flex gap-x-2 items-center min-h-[20px]">
              <ResourceTypeBadge
                kind={dashboard.kind === "MetricsView"
                  ? ResourceKind.Explore
                  : ResourceKind.Canvas}
              />
              <span
                class="text-sm font-semibold truncate"
                class:text-red-600={dashboard.hasError}
                class:dark:text-red-400={dashboard.hasError}
                class:group-hover:text-primary-600={!dashboard.hasError}
                class:dark:group-hover:text-primary-400={!dashboard.hasError}
                style:color={!dashboard.hasError ? "var(--fg-primary)" : undefined}
              >
                {getDisplayName(dashboard)}
              </span>
              {#if dashboard.hasError}
                <span
                  class="text-xs px-1.5 py-0.5 rounded bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-400 font-medium"
                >
                  Error
                </span>
              {/if}
            </div>

            <div
              class="flex gap-x-1 text-xs font-normal min-h-[16px] overflow-hidden"
              class:text-red-500={dashboard.hasError}
              class:dark:text-red-400={dashboard.hasError}
              style:color={!dashboard.hasError ? "var(--fg-muted)" : undefined}
            >
              <span class="shrink-0">
                {dashboard.fullPath || dashboard.name}
              </span>
              {#if !dashboard.hasError && dashboard.lastRefreshed}
                <span class="shrink-0">&bull;</span>
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
          </div>
          <TooltipContent slot="tooltip-content">
            {dashboard.errorMessage || "Dashboard has errors"}
          </TooltipContent>
        </Tooltip>
      {/each}
    </div>
    {#if showSeeAll && hasMore}
      <div class="pl-4 py-1">
        <a
          href={seeAllHref}
          class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors inline-block"
        >
          See all dashboards &rarr;
        </a>
      </div>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .dashboard-row.hoverable:hover {
    background: var(--surface-subtle);
  }
</style>
