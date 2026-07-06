<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { BreadcrumbItemDropdownProps } from "@rilldata/web-common/components/navigation/breadcrumbs/types.ts";
  import BreadcrumbDropdownItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbDropdownItem.svelte";
  import {
    getAllTagsForResources,
    getResourceTags,
    UNTAGGED_KEY,
  } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import DashboardsTagFilter from "@rilldata/web-admin/features/dashboards/listing/DashboardsTagFilter.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import {
    ArrayRuneStore,
    InMemoryRuneStore,
  } from "@rilldata/web-common/lib/store-utils/types.ts";

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

  // `id` in `options` is the resource name (`V1Resource.meta.name.name`), so
  // build a lookup to resolve each option back to its resource for tag filtering.
  let dashboardsByName = $derived(
    new Map(allDashboards.map((r) => [r.meta?.name?.name?.toLowerCase(), r])),
  );

  function matchesTags(id: string): boolean {
    if (selectedTagsStore.value.length === 0) return true;
    const resource = dashboardsByName.get(id);
    if (!resource) return false;
    const resourceTags = getResourceTags(resource);
    return selectedTagsStore.value.some((t) =>
      t === UNTAGGED_KEY ? resourceTags.length === 0 : resourceTags.includes(t),
    );
  }

  function matchesSearch(id: string, label: string): boolean {
    if (!searchText) return true;
    const q = searchText.toLowerCase();
    return id.toLowerCase().includes(q) || label.toLowerCase().includes(q);
  }

  let filteredOptions = $derived(
    [...options].filter(
      ([id, option]) => matchesTags(id) && matchesSearch(id, option.label),
    ),
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

  {#each filteredOptions as [id, option] (id)}
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
