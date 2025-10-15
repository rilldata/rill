<script lang="ts">
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useDashboards } from "./selectors";
  import { timeAgo } from "./utils";

  export let organization: string;
  export let project: string;
  export let limit = 5;

  $: ({ instanceId } = $runtime);

  $: dashboardsQuery = useDashboards(instanceId);
  $: ({ data: dashboards, isLoading: dashboardsLoading } = $dashboardsQuery);
  $: shortlistDashboards = dashboards?.slice(0, limit) ?? [];
</script>

<div class="flex flex-col gap-y-6">
  <h2 class="text-xl font-semibold text-gray-900">Dashboards</h2>

  {#if dashboardsLoading}
    <div class="flex justify-center py-12">
      <div class="text-gray-500">Loading dashboards...</div>
    </div>
  {:else if shortlistDashboards.length === 0}
    <div class="text-center py-12 text-gray-500">
      <p class="text-base">No dashboards yet.</p>
      <p class="text-sm mt-3">
        Learn how to create a dashboard in our
        <a
          href="https://docs.rilldata.com/"
          target="_blank"
          class="text-primary-600 hover:text-primary-700"
        >
          docs
        </a>
      </p>
    </div>
  {:else}
    <div class="flex flex-col">
      {#each shortlistDashboards as resource, i}
        {@const name = resource.meta?.name?.name ?? ""}
        {@const isMetricsExplorer = !!resource?.explore}
        {@const title = isMetricsExplorer
          ? (resource.explore?.spec?.displayName ?? "")
          : (resource.canvas?.spec?.displayName ?? "")}
        {@const description = isMetricsExplorer
          ? (resource.explore?.spec?.description ?? "")
          : ""}
        {@const refreshedOn = isMetricsExplorer
          ? resource.explore?.state?.dataRefreshedOn
          : resource.canvas?.state?.dataRefreshedOn}
        {@const lastRefreshedDate = refreshedOn ? new Date(refreshedOn) : null}
        {@const dashboardSlug = isMetricsExplorer ? "explore" : "canvas"}
        {@const href = `/${organization}/${project}/${dashboardSlug}/${name}`}

        <a
          {href}
          class="flex items-center gap-x-3 px-4 py-3 hover:bg-gray-50 transition-colors group"
          class:border-t={i > 0}
          class:border-gray-200={i > 0}
        >
          <div class="flex-shrink-0">
            {#if isMetricsExplorer}
              <ExploreIcon size="16px" className="text-gray-500" />
            {:else}
              <CanvasIcon size="16px" className="text-gray-500" />
            {/if}
          </div>
          <div class="flex-1 min-w-0">
            <div
              class="font-medium text-gray-900 text-sm group-hover:text-primary-600 transition-colors"
            >
              {title !== "" ? title : name}
            </div>
            <div class="flex items-center gap-x-2 text-xs text-gray-500 mt-0.5">
              <span class="font-mono truncate">{name}</span>
              {#if lastRefreshedDate}
                <span>•</span>
                <span class="shrink-0"
                  >Updated {timeAgo(lastRefreshedDate)}</span
                >
              {/if}
              {#if description}
                <span>•</span>
                <span class="truncate">{description}</span>
              {/if}
            </div>
          </div>
          <div class="flex-shrink-0">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 20 20"
              fill="currentColor"
              class="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <path
                fill-rule="evenodd"
                d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
                clip-rule="evenodd"
              />
            </svg>
          </div>
        </a>
      {/each}
    </div>

    {#if dashboards && dashboards.length > limit}
      <div class="flex justify-center pt-2">
        <a
          href={`/${organization}/${project}/-/dashboards`}
          class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors"
        >
          See all dashboards →
        </a>
      </div>
    {/if}
  {/if}
</div>
