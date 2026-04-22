<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import DashboardsFilterToolbar from "./DashboardsFilterToolbar.svelte";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import DashboardsTagFolder from "./DashboardsTagFolder.svelte";
  import {
    UNTAGGED_KEY,
    getResourceTags,
    useDashboards,
    useIsInitialBuild,
  } from "./selectors";

  export let isEmbedded = false;
  export let isPreview = false;
  export let previewLimit = 5;

  const TAGS_PARAM = "tags";
  const { tagAsFolders } = featureFlags;

  $: selectedTags = ($page.url.searchParams.get(TAGS_PARAM) ?? "")
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean);

  function setSelectedTags(tags: string[]) {
    const url = new URL($page.url);
    if (tags.length === 0) {
      url.searchParams.delete(TAGS_PARAM);
    } else {
      url.searchParams.set(TAGS_PARAM, tags.join(","));
    }
    void goto(url, { replaceState: true, noScroll: true, keepFocus: true });
  }

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

  $: availableTags = Array.from(
    new Set(allDashboards.flatMap(getResourceTags)),
  ).sort();

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

  // Folder mode: group dashboards by tag. A dashboard with multiple tags
  // appears under each of its tags. Tags respect the selectedTags filter.
  $: tagGroups = (() => {
    const activeTags = selectedTags.length > 0 ? selectedTags : availableTags;
    const groups: { tag: string; resources: V1Resource[] }[] = [];
    const untagged: V1Resource[] = [];
    const untaggedVisible =
      selectedTags.length === 0 || selectedTags.includes(UNTAGGED_KEY);

    for (const tag of activeTags) {
      if (tag === UNTAGGED_KEY) continue;
      const members = searchFilteredDashboards.filter((r) =>
        getResourceTags(r).includes(tag),
      );
      if (members.length > 0) groups.push({ tag, resources: members });
    }

    if (untaggedVisible) {
      for (const r of searchFilteredDashboards) {
        if (getResourceTags(r).length === 0) untagged.push(r);
      }
      if (untagged.length > 0)
        groups.push({ tag: UNTAGGED_KEY, resources: untagged });
    }

    return groups;
  })();

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
        const tags = isMetricsExplorer
          ? (resource.explore?.spec?.tags ?? [])
          : (resource.canvas?.spec?.tags ?? []);

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
      <DashboardsFilterToolbar
        availableTags={$tagAsFolders ? [] : availableTags}
        {selectedTags}
        onTagsChange={setSelectedTags}
        bind:searchText
      />
    {/if}

    {#if $tagAsFolders && !isPreview}
      <!-- Folder mode: grouped by tag, all inside one bordered container -->
      {#if tagGroups.length === 0}
        <div class="text-center py-16 text-fg-secondary text-sm font-semibold">
          {searchText
            ? "No dashboards match your search"
            : "You don't have any dashboards yet"}
        </div>
      {:else}
        <div class="w-full border rounded-lg overflow-hidden divide-y">
          {#each tagGroups as { tag, resources } (tag)}
            <DashboardsTagFolder
              {tag}
              {resources}
              {organization}
              {project}
              {isEmbedded}
            />
          {/each}
        </div>
      {/if}
    {:else}
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
    {/if}

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
