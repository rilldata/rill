<script lang="ts">
  import TableToolbarAppliedFilters from "./TableToolbarAppliedFilters.svelte";
  import TableToolbarFilterDropdown from "./TableToolbarFilterDropdown.svelte";
  import TableToolbarSearch from "./TableToolbarSearch.svelte";
  import TableToolbarSort from "./TableToolbarSort.svelte";
  import TableToolbarViewToggle from "./TableToolbarViewToggle.svelte";
  import type { FilterGroup, SortDirection, ViewMode } from "./types";
  import type { Snippet } from "svelte";

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
    viewMode = "list" as ViewMode,
    onViewModeChange,
    children,
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
    children?: Snippet;
  } = $props();
</script>

<section class="flex flex-col gap-y-2 w-full">
  <div class="flex flex-row items-center justify-between h-9 gap-x-4">
    <div class="flex flex-row items-center">
      <TableToolbarFilterDropdown {filterGroups} {onFilterChange} />
    </div>

    <div class="flex flex-row items-center gap-x-3">
      <TableToolbarSearch
        {searchText}
        {onSearchChange}
        disabled={searchDisabled}
      />

      {#if showSort}
        <TableToolbarSort {sortDirection} {onSortToggle} />
      {/if}

      {#if showViewToggle}
        <TableToolbarViewToggle {viewMode} {onViewModeChange} />
      {/if}

      {@render children?.()}
    </div>
  </div>

  <hr class="border-t" />

  <TableToolbarAppliedFilters
    {filterGroups}
    {onFilterChange}
    {onClearAllFilters}
  />
</section>
