<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
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
    hasError?: boolean;
    errorMessage?: string;
  }

  let searchQuery = "";

  $: resourcesQuery = createRuntimeServiceListResources(
    $runtime.instanceId,
    {},
  );

  $: dashboards = ($resourcesQuery.data?.resources ?? [])
    .filter((resource) => {
      const kind = resource.meta?.name?.kind;
      return (
        kind === "rill.runtime.v1.Explore" || kind === "rill.runtime.v1.Canvas"
      );
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

      const isExplore = kind === "rill.runtime.v1.Explore";
      const refreshedOn = isExplore
        ? resource.explore?.state?.dataRefreshedOn
        : resource.canvas?.state?.dataRefreshedOn;

      return {
        name,
        title,
        kind: isExplore ? "MetricsView" : "Canvas",
        lastRefreshed: refreshedOn,
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

  $: displayData = limit
    ? filteredDashboards.slice(0, limit)
    : filteredDashboards;
  $: hasMore = limit ? dashboards.length > limit : false;

  function getDisplayName(dashboard: Dashboard): string {
    return dashboard.title || dashboard.name;
  }

  function getDashboardHref(dashboard: Dashboard): string {
    const slug = dashboard.kind === "MetricsView" ? "explore" : "canvas";
    return `/${slug}/${dashboard.name}`;
  }
</script>

<div class="flex flex-col w-full gap-y-3">
  {#if showSearch && dashboards.length > 0}
    <Search
      placeholder="Search"
      autofocus={false}
      bind:value={searchQuery}
      rounded="lg"
    />
  {/if}

  {#if $resourcesQuery.isLoading && dashboards.length === 0}
    <div class="m-auto mt-20">
      <DelayedSpinner isLoading={true} size="24px" />
    </div>
  {:else if dashboards.length === 0}
    <div class="text-center py-16">
      <div class="flex flex-col gap-y-2 items-center text-sm">
        <div class="text-fg-secondary font-semibold">
          You don't have any dashboards yet
        </div>
      </div>
    </div>
  {:else if filteredDashboards.length === 0}
    <div class="text-center py-16">
      <div class="flex flex-col gap-y-2 items-center text-sm">
        <div class="text-fg-secondary font-semibold">
          No dashboards match your search
        </div>
        <div class="text-fg-secondary">Try adjusting your search terms</div>
      </div>
    </div>
  {:else}
    <ul class="resource-list">
      {#each displayData as dashboard (dashboard.name)}
        <li class="resource-list-item">
          <a
            class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
            href={getDashboardHref(dashboard)}
          >
            <div class="flex gap-x-2 items-center min-h-[20px]">
              <ResourceTypeBadge
                kind={dashboard.kind === "MetricsView"
                  ? ResourceKind.Explore
                  : ResourceKind.Canvas}
              />
              <span
                class="text-fg-secondary text-sm font-semibold group-hover:text-accent-primary-action truncate"
              >
                {getDisplayName(dashboard)}
              </span>
              {#if dashboard.hasError}
                <Tag color="red">Error</Tag>
              {/if}
            </div>
            <div
              class="flex gap-x-1 text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden"
            >
              <span class="shrink-0">{dashboard.name}</span>
              {#if dashboard.lastRefreshed}
                <span class="shrink-0">&bull;</span>
                <Tooltip distance={8}>
                  <span class="shrink-0"
                    >Last refreshed {timeAgo(
                      new Date(dashboard.lastRefreshed),
                    )}</span
                  >
                  <TooltipContent slot="tooltip-content">
                    {new Date(dashboard.lastRefreshed).toLocaleString()}
                  </TooltipContent>
                </Tooltip>
              {/if}
            </div>
          </a>
        </li>
      {/each}
    </ul>
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
  .resource-list {
    @apply list-none p-0 m-0 w-full;
  }

  .resource-list-item {
    @apply block w-full border bg-surface-background;
  }

  .resource-list-item + .resource-list-item {
    @apply border-t-0;
  }

  .resource-list-item:first-child {
    @apply rounded-t-lg;
  }

  .resource-list-item:last-child {
    @apply rounded-b-lg;
  }

  .resource-list-item:hover {
    @apply bg-surface-hover;
  }
</style>
