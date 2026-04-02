<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
    ViewMode,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import type { Table } from "tanstack-table-8-svelte-5";
  import { getContext } from "svelte";
  import type { Readable } from "svelte/store";

  const table = getContext<Readable<Table<unknown>>>("table");

  let {
    searchPlaceholder = "Search",
    searchDisabled = false,
    filterGroups = [],
    onFilterChange,
    sortColumnId = "",
    showSort = true,
    showViewToggle = false,
    viewMode = $bindable("list"),
    onViewModeChange,
  }: {
    searchPlaceholder?: string;
    searchDisabled?: boolean;
    filterGroups?: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
    sortColumnId?: string;
    showSort?: boolean;
    showViewToggle?: boolean;
    viewMode?: ViewMode;
    onViewModeChange?: (mode: ViewMode) => void;
  } = $props();

  let searchText = $state("");
  let sortDirection = $state<SortDirection>("newest");

  $effect(() => {
    $table.setGlobalFilter(searchText);
  });

  function handleSortChange(direction: SortDirection) {
    sortDirection = direction;
    if (sortColumnId) {
      $table.setSorting([{ id: sortColumnId, desc: direction === "newest" }]);
    }
  }

  beforeNavigate(() => (searchText = ""));
</script>

<TableToolbar
  bind:searchText
  {searchPlaceholder}
  {searchDisabled}
  {filterGroups}
  {onFilterChange}
  {sortDirection}
  onSortChange={handleSortChange}
  {showSort}
  {showViewToggle}
  bind:viewMode
  {onViewModeChange}
/>
