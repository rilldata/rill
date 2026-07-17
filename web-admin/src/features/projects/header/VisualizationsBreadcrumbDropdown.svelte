<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { BreadcrumbItemDropdownProps } from "@rilldata/web-common/components/navigation/breadcrumbs/types.ts";
  import BreadcrumbDropdownItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbDropdownItem.svelte";
  import { getAllTagsForResources } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import DashboardsTagFilter from "@rilldata/web-admin/features/dashboards/listing/DashboardsTagFilter.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import {
    ArrayRuneStore,
    InMemoryRuneStore,
  } from "web-common/src/lib/store-utils/types.svelte.ts";
  import { filterResources } from "@rilldata/web-common/features/resources/resource-filter-utils.ts";
  import {
    getDashboardFavouritesStore,
    sortByFavourites,
  } from "../../dashboards/listing/dashboard-favourites.ts";
  import { page } from "$app/state";

  let {
    options,
    current,
    currentPath = [],
    depth = 0,
    onSelect = undefined,
    linkMaker,
  }: BreadcrumbItemDropdownProps = $props();

  const runtimeClient = useRuntimeClient();
  const dashboards = useDashboards(runtimeClient);
  let allDashboards = $derived($dashboards.data ?? []);
  let availableTags = $derived(getAllTagsForResources(allDashboards));
  let hasSomeTag = $derived(availableTags.length > 0);

  const selectedTagsStore = new ArrayRuneStore<string>(
    new InMemoryRuneStore([]),
  );
  let searchText = $state("");

  let filteredDashboards = $derived(
    filterResources(allDashboards, [], searchText, [], selectedTagsStore.value),
  );

  let filteredDashboardNames = $derived(
    new Set(
      filteredDashboards.map((r) => r.meta?.name?.name?.toLowerCase() ?? ""),
    ),
  );

  let filteredOptions = $derived(
    [...options].filter(([id]) => filteredDashboardNames.has(id)),
  );

  let { organization, project } = $derived(page.params);
  let dashboardFavourites = $derived(
    getDashboardFavouritesStore(organization, project),
  );

  let sortedOptions = $derived(
    sortByFavourites(filteredOptions, dashboardFavourites.value, ([id]) => id),
  );
</script>

<DropdownMenu.Content align="start" class="min-w-60 max-h-96 overflow-y-auto">
  {#if hasSomeTag}
    <div class="flex flex-row gap-x-2 mb-2 items-center">
      <DashboardsTagFilter size="xs" {selectedTagsStore} />

      <Search
        placeholder="Search"
        autofocus={false}
        bind:value={searchText}
        rounded="lg"
      />
    </div>
  {/if}

  {#each sortedOptions as [id, option] (id)}
    <BreadcrumbDropdownItem
      {id}
      {option}
      {current}
      {currentPath}
      {depth}
      {onSelect}
      {linkMaker}
    />
  {/each}
</DropdownMenu.Content>
