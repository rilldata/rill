<script lang="ts">
  import { page } from "$app/stores";
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import {
    UNTAGGED_KEY,
    getResourceTags,
    useDashboards,
    useIsInitialBuild,
  } from "./selectors";
  import { Search } from "@rilldata/web-common/components/search";
  import DashboardsTagFilter from "@rilldata/web-admin/features/dashboards/listing/DashboardsTagFilter.svelte";

  export let isEmbedded = false;
  export let isPreview = false;
  export let previewLimit = 5;

  $: selectedTags = ($page.url.searchParams.get("tags") ?? "")
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean);

  let searchText = "";

  const runtimeClient = useRuntimeClient();
  $: ({
    params: { organization, project },
  } = $page);

  $: dashboards = useDashboards(runtimeClient);
  $: ({
    data: dashboardsData,
    isLoading,
    isError,
    isSuccess,
    error,
  } = $dashboards);

  $: initialBuild = useIsInitialBuild(runtimeClient);
  $: isBuilding = $initialBuild.data === true;

  function matchesSearch(resource: V1Resource, query: string): boolean {
    if (!query) return true;
    const q = query.toLowerCase();
    const name = resource.meta?.name?.name ?? "";
    const title = resource.explore
      ? (resource.explore.spec?.displayName ?? "")
      : (resource.canvas?.spec?.displayName ?? "");
    const desc = resource.explore?.spec?.description ?? "";
    return (
      name.toLowerCase().includes(q) ||
      title.toLowerCase().includes(q) ||
      desc.toLowerCase().includes(q)
    );
  }

  $: allDashboards = dashboardsData ?? [];

  $: tagFilteredDashboards =
    selectedTags.length === 0
      ? allDashboards
      : allDashboards.filter((resource) => {
          const resourceTags = getResourceTags(resource);
          return selectedTags.some((t) =>
            t === UNTAGGED_KEY
              ? resourceTags.length === 0
              : resourceTags.includes(t),
          );
        });

  $: searchFilteredDashboards = tagFilteredDashboards.filter((r) =>
    matchesSearch(r, searchText),
  );

  $: displayData = isPreview
    ? searchFilteredDashboards.slice(0, previewLimit)
    : searchFilteredDashboards;

  $: hasMoreDashboards =
    isPreview && searchFilteredDashboards.length > previewLimit;

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
        <DashboardsTagFilter />

        <div class="flex-1 min-w-0">
          <Search
            placeholder="Search"
            autofocus={false}
            bind:value={searchText}
            rounded="lg"
          />
        </div>
      </div>
    {/if}

    <!-- Flat mode: standard list -->
    <ResourceList
      kind="dashboard"
      data={displayData}
      {columns}
      {columnVisibility}
      {initialSorting}
      toolbar={false}
    >
      <ResourceListEmptyState
        slot="empty"
        icon={ExploreIcon}
        message="You don't have any dashboards yet"
      >
        <span slot="action">
          <a
            href="https://docs.rilldata.com/developers/build/dashboards"
            target="_blank"
            rel="noopener noreferrer"
          >
            Create a dashboard</a
          > to get started
        </span>
      </ResourceListEmptyState>
    </ResourceList>

    {#if hasMoreDashboards}
      <div class="pl-4 py-1">
        <a
          href={`/${organization}/${project}/-/dashboards`}
          class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors inline-block"
        >
          See all dashboards →
        </a>
      </div>
    {/if}
  </div>
{/if}
