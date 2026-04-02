<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import TableToolbarFilterDropdown from "./TableToolbarFilterDropdown.svelte";
  import TableToolbarSortDropdown from "./TableToolbarSortDropdown.svelte";
  import TableToolbarViewToggle from "./TableToolbarViewToggle.svelte";
  import type { FilterGroup, SortDirection, ViewMode } from "./types";

  let {
    searchText = $bindable(""),
    searchPlaceholder = "Search",
    searchDisabled = false,
    filterGroups = [],
    onFilterChange,
    sortDirection = "newest",
    onSortChange,
    showSort = true,
    showViewToggle = false,
    viewMode = $bindable("list"),
    onViewModeChange,
  }: {
    searchText?: string;
    searchPlaceholder?: string;
    searchDisabled?: boolean;
    filterGroups?: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
    sortDirection?: SortDirection;
    onSortChange?: (direction: SortDirection) => void;
    showSort?: boolean;
    showViewToggle?: boolean;
    viewMode?: ViewMode;
    onViewModeChange?: (mode: ViewMode) => void;
  } = $props();
</script>

<div class="flex flex-row items-center justify-between gap-x-3 w-full">
  <div class="flex flex-row items-center gap-x-2">
    <TableToolbarFilterDropdown {filterGroups} {onFilterChange} />
  </div>

  <div class="flex flex-row items-center gap-x-2">
    <div class="w-48">
      <Search
        placeholder={searchPlaceholder}
        bind:value={searchText}
        autofocus={false}
        showBorderOnFocus={false}
        disabled={searchDisabled}
      />
    </div>

    {#if showSort}
      <TableToolbarSortDropdown {sortDirection} {onSortChange} />
    {/if}

    {#if showViewToggle}
      <TableToolbarViewToggle {viewMode} {onViewModeChange} />
    {/if}
  </div>
</div>
