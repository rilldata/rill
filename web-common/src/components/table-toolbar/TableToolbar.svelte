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
    searchPlaceholder = "Search",
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
    searchPlaceholder?: string;
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

<section class="flex flex-col w-full gap-y-2">
  <div class="flex flex-row items-center gap-x-2.5">
    <div class="flex flex-row items-center gap-x-2.5 shrink-0">
      <TableToolbarFilterDropdown {filterGroups} {onFilterChange} />

      {#if showSort}
        <TableToolbarSort bind:sortDirection />
      {/if}
    </div>

    <TableToolbarSearch
      bind:searchText
      disabled={searchDisabled}
      placeholder={searchPlaceholder}
    />

    <div class="flex flex-row items-center gap-x-2.5 shrink-0">
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
