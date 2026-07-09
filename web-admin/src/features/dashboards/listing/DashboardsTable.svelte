<script lang="ts">
  import { page } from "$app/state";
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import { useDashboards, useIsInitialBuild } from "./selectors";
  import { Search } from "@rilldata/web-common/components/search";
  import { UrlParamsState } from "web-common/src/lib/store-utils/url-params-state.svelte.ts";
  import { getAllTagsForResources } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import ResizableSidebar from "@rilldata/web-common/layout/ResizableSidebar.svelte";
  import DashboardsTagSidebar from "@rilldata/web-admin/features/dashboards/listing/DashboardsTagSidebar.svelte";
  import { filterResources } from "@rilldata/web-common/features/resources/resource-filter-utils.ts";
  import { Throttler } from "@rilldata/web-common/lib/throttler.ts";
  import { getDashboardFavouritesStore } from "@rilldata/web-admin/features/dashboards/listing/dashboard-favourites.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { escapeHtml } from "@rilldata/web-common/lib/i18n";

  let {
    isEmbedded = false,
    isPreview = false,
    previewLimit = undefined,
  }: {
    isEmbedded?: boolean;
    isPreview?: boolean;
    previewLimit?: number;
  } = $props();

  const selectedTagsState = UrlParamsState.createStringArrayParam("tags");

  const searchTextState = UrlParamsState.createStringParam("search");
  const throttler = new Throttler(500, 500);
  const throttledSearchSetter = (newValue: string) => {
    throttler.throttle(() => searchTextState.setter(newValue));
  };

  const runtimeClient = useRuntimeClient();
  let { organization, project } = $derived(page.params);

  const dashboards = useDashboards(runtimeClient);
  let {
    data: dashboardsData,
    isLoading,
    isError,
    isSuccess,
    error,
  } = $derived($dashboards);

  let initialBuild = useIsInitialBuild(runtimeClient);
  let isBuilding = $derived($initialBuild.data === true);

  let allDashboards = $derived(dashboardsData ?? []);
  let availableTags = $derived(getAllTagsForResources(allDashboards));
  let hasSomeTag = $derived(availableTags.length > 0);

  let filteredDashboards = $derived(
    filterResources(
      allDashboards,
      [],
      searchTextState.value,
      [],
      selectedTagsState.value,
    ),
  );

  let dashboardFavourites = $derived(
    getDashboardFavouritesStore(organization, project),
  );

  let validDashboardFavourites = $derived(
    dashboardFavourites.value.filter((f) =>
      filteredDashboards.find((r) => r.meta?.name?.name === f),
    ),
  );

  let hasMoreDashboards = $derived(
    isPreview && filteredDashboards.length > previewLimit,
  );

  const columns = [
    {
      id: "composite",
      cell: ({ row }) => {
        const resource = row.original as V1Resource;
        const name = resource.meta.name.name;
        const isMetricsExplorer = !!resource?.explore;
        const title = isMetricsExplorer
          ? resource.explore.spec.displayName
          : resource.canvas.spec.displayName;
        const description = isMetricsExplorer
          ? resource.explore.spec.description
          : "";
        const refreshedOn = isMetricsExplorer
          ? resource.explore?.state?.dataRefreshedOn
          : resource.canvas?.state?.dataRefreshedOn;
        const tags = resource.meta?.tags ?? [];

        return renderComponent(DashboardsTableCompositeCell, {
          name,
          title,
          lastRefreshed: refreshedOn,
          description,
          error: resource.meta.reconcileError,
          isMetricsExplorer,
          isEmbedded,
          organization,
          project,
          tags,
          dashboardFavourites,
        });
      },
    },
    {
      id: "title",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row?.explore;
        return isMetricsExplorer
          ? row.explore.spec.displayName
          : row.canvas.spec.displayName;
      },
    },
    {
      id: "name",
      accessorFn: (row: V1Resource) => row.meta.name.name,
    },
    {
      id: "lastRefreshed",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row?.explore;
        return isMetricsExplorer
          ? row.explore?.state?.dataRefreshedOn
          : row.canvas?.state?.dataRefreshedOn;
      },
    },
    {
      id: "description",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row?.explore;
        return isMetricsExplorer ? row.explore.spec.description : "";
      },
    },
  ];

  const columnVisibility = {
    title: false,
    name: false,
    lastRefreshed: false,
    description: false,
  };

  const initialSorting = [{ id: "name", desc: false }];
</script>

{#if isLoading || isBuilding}
  <div class="m-auto mt-20">
    <DelayedSpinner isLoading={true} size="24px" />
  </div>
{:else if isError}
  <ResourceError kind="dashboard" {error} />
{:else if isSuccess}
  <div class="flex flex-col w-full gap-y-3">
    {#if !isPreview}
      <div class="flex flex-row items-center gap-x-2">
        <Search
          placeholder={m.common_search()}
          autofocus={false}
          bind:value={searchTextState.getter, throttledSearchSetter}
          rounded="lg"
          retainValueOnMount
        />
      </div>
    {/if}

    <div class="flex flex-row flex-1 w-full gap-x-2 overflow-hidden">
      {#if hasSomeTag && !isPreview}
        <ResizableSidebar
          id="dashboards-tag-sidebar"
          minWidth={200}
          maxWidth={500}
          defaultWidth={200}
          additionalClass="overflow-hidden border rounded-lg"
          side="right"
        >
          <DashboardsTagSidebar
            resources={allDashboards}
            searchText={searchTextState.value}
          />
        </ResizableSidebar>
      {/if}

      <div class="flex flex-col flex-grow">
        <ResourceList
          kind="dashboard"
          data={filteredDashboards}
          {columns}
          {columnVisibility}
          {initialSorting}
          toolbar={false}
          pinnedRows={validDashboardFavourites}
          maxRows={previewLimit}
        >
          <ResourceListEmptyState
            slot="empty"
            icon={ExploreIcon}
            message={m.dashboard_list_empty()}
          >
            <span slot="action">
              {@html m.dashboard_list_create_to_start({
                link: `<a href="https://docs.rilldata.com/developers/build/dashboards" target="_blank" rel="noopener noreferrer">${escapeHtml(m.dashboard_list_create())}</a>`,
              })}
            </span>
          </ResourceListEmptyState>
        </ResourceList>

        {#if hasMoreDashboards}
          <div class="pl-4 py-1">
            <a
              href={`/${organization}/${project}/-/dashboards`}
              class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors inline-block"
            >
              {m.dashboard_list_see_all()} →
            </a>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
