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
    showSearch = true,
    filterGroups = [],
    onFilterChange,
    onClearAllFilters,
    sortDirection = $bindable("newest"),
    showSort = true,
    showViewToggle = false,
    viewMode = $bindable("list"),
    disabled = false,
    children,
  }: {
    searchText?: string;
    searchDisabled?: boolean;
    showSearch?: boolean;
    filterGroups?: FilterGroup[];
    onFilterChange?: (key: string, selected: string | string[]) => void;
    onClearAllFilters?: () => void;
    sortDirection?: SortDirection;
    showSort?: boolean;
    showViewToggle?: boolean;
    viewMode?: ViewMode;
    /** Disables search, filter, and sort. Useful when the underlying data is empty. */
    disabled?: boolean;
    children?: Snippet;
  } = $props();
</script>

<section class="flex flex-col w-full">
  <div class="flex flex-row items-center justify-between h-9 gap-x-4">
    <div class="flex flex-row items-center">
      <TableToolbarFilterDropdown {filterGroups} {onFilterChange} {disabled} />
    </div>

    <div class="flex flex-row items-center gap-x-3">
      {#if showSearch}
        <TableToolbarSearch
          bind:searchText
          disabled={searchDisabled || disabled}
        />
      {/if}

      {#if showSort}
        <TableToolbarSort bind:sortDirection {disabled} />
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
