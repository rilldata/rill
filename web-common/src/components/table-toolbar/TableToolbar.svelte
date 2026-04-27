<script lang="ts">
  import TableToolbarAppliedFilters from "./TableToolbarAppliedFilters.svelte";
  import TableToolbarFilterDropdown from "./TableToolbarFilterDropdown.svelte";
  import TableToolbarSearch from "./TableToolbarSearch.svelte";
  import TableToolbarSort from "./TableToolbarSort.svelte";
  import TableToolbarViewToggle from "./TableToolbarViewToggle.svelte";
  import type { FilterGroup, SortDirection, ViewMode } from "./types";
  import type { Snippet } from "svelte";

  let {
    searchText = $bindable(""),
    searchDisabled = false,
    filterGroups = [],
    onFilterChange,
    onClearAllFilters,
    sortDirection = $bindable("newest"),
    showSort = true,
    showViewToggle = false,
    viewMode = $bindable("list"),
    children,
  }: {
    searchText?: string;
    searchDisabled?: boolean;
    filterGroups?: FilterGroup[];
    onFilterChange?: (key: string, selected: string | string[]) => void;
    onClearAllFilters?: () => void;
    sortDirection?: SortDirection;
    showSort?: boolean;
    showViewToggle?: boolean;
    viewMode?: ViewMode;
    children?: Snippet;
  } = $props();
</script>

<section class="flex flex-col w-full">
  <div class="flex flex-row items-center justify-between h-9 gap-x-4">
    <div class="flex flex-row items-center">
      <TableToolbarFilterDropdown {filterGroups} {onFilterChange} />
    </div>

    <div class="flex flex-row items-center gap-x-3">
      <TableToolbarSearch bind:searchText disabled={searchDisabled} />

      {#if showSort}
        <TableToolbarSort bind:sortDirection />
      {/if}

      {#if showViewToggle}
        <TableToolbarViewToggle bind:viewMode />
      {/if}

      {@render children?.()}
    </div>
  </div>

  <TableToolbarAppliedFilters
    {filterGroups}
    {onFilterChange}
    {onClearAllFilters}
  />
</section>
