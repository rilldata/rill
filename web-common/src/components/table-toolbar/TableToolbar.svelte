<script lang="ts">
  import TableToolbarAppliedFilters from "./TableToolbarAppliedFilters.svelte";
  import TableToolbarFilterDropdown from "./TableToolbarFilterDropdown.svelte";
  import TableToolbarSearch from "./TableToolbarSearch.svelte";
  import TableToolbarSort from "./TableToolbarSort.svelte";
  import TableToolbarViewToggle from "./TableToolbarViewToggle.svelte";
  import type { FilterGroup, SortDirection, ViewMode } from "./types";

  let {
    searchText = "",
    onSearchChange,
    searchDisabled = false,
    filterGroups = [],
    onFilterChange,
    onClearAllFilters,
    sortDirection = "newest",
    onSortToggle,
    showSort = true,
    showViewToggle = false,
    viewMode = $bindable("list"),
    onViewModeChange,
  }: {
    searchText?: string;
    onSearchChange?: (text: string) => void;
    searchDisabled?: boolean;
    filterGroups?: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
    onClearAllFilters?: () => void;
    sortDirection?: SortDirection;
    onSortToggle?: () => void;
    showSort?: boolean;
    showViewToggle?: boolean;
    viewMode?: ViewMode;
    onViewModeChange?: (mode: ViewMode) => void;
  } = $props();

  function handleSearchChange(text: string) {
    onSearchChange?.(text);
  }
</script>

<div class="flex flex-col gap-y-0 w-full">
  <div class="flex flex-row items-center justify-between h-9">
    <div class="flex flex-row items-center">
      <TableToolbarFilterDropdown {filterGroups} {onFilterChange} />
    </div>

    <div class="flex flex-row items-center gap-x-3">
      {#if showSort}
        <TableToolbarSort {sortDirection} {onSortToggle} />
      {/if}

      <TableToolbarSearch
        {searchText}
        onSearchChange={handleSearchChange}
        disabled={searchDisabled}
      />

      {#if showViewToggle}
        <TableToolbarViewToggle {viewMode} {onViewModeChange} />
      {/if}
    </div>
  </div>

  <TableToolbarAppliedFilters
    {filterGroups}
    {onFilterChange}
    {onClearAllFilters}
  />
</div>
