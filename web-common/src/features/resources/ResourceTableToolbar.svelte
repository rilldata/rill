<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
    ViewMode,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import {
    createUrlFilterSync,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import type { Table } from "tanstack-table-8-svelte-5";
  import { getContext, onMount, untrack } from "svelte";
  import { get, type Readable } from "svelte/store";

  const table = getContext<Readable<Table<unknown>>>("table");

  let {
    searchDisabled = false,
    filterGroups = [],
    onFilterChange,
    onClearAllFilters,
    sortColumnId = "",
    showSort = true,
    showViewToggle = false,
    viewMode = $bindable("list"),
    onViewModeChange,
  }: {
    searchDisabled?: boolean;
    filterGroups?: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
    onClearAllFilters?: () => void;
    sortColumnId?: string;
    showSort?: boolean;
    showViewToggle?: boolean;
    viewMode?: ViewMode;
    onViewModeChange?: (mode: ViewMode) => void;
  } = $props();

  // URL sync for search
  const filterSync = createUrlFilterSync([{ key: "q", type: "string" }]);
  filterSync.init(get(page).url);

  let searchText = $state(
    parseStringParam(get(page).url.searchParams.get("q")),
  );
  let sortDirection = $state<SortDirection>("newest");
  let mounted = $state(false);

  onMount(() => {
    mounted = true;
  });

  // Sync search to TanStack global filter; untrack $table to avoid
  // re-triggering when the store updates from setGlobalFilter
  $effect(() => {
    const text = searchText;
    untrack(() => {
      get(table).setGlobalFilter(text);
    });
  });

  // Sync search to URL; untrack to avoid circular $page dependency
  $effect(() => {
    const text = searchText;
    untrack(() => {
      if (mounted) {
        filterSync.syncToUrl({ q: text });
      }
    });
  });

  // Sync URL back to state on external navigation (back/forward);
  // subscribe to $page reactively, but write searchText in untrack
  // to avoid re-triggering the URL sync effect
  $effect(() => {
    const url = $page.url;
    if (mounted && filterSync.hasExternalNavigation(url)) {
      filterSync.markSynced(url);
      untrack(() => {
        searchText = parseStringParam(url.searchParams.get("q"));
      });
    }
  });

  function handleSortToggle() {
    sortDirection = sortDirection === "newest" ? "oldest" : "newest";
    if (sortColumnId) {
      get(table).setSorting([
        { id: sortColumnId, desc: sortDirection === "newest" },
      ]);
    }
  }

  beforeNavigate(() => (searchText = ""));
</script>

<TableToolbar
  bind:searchText
  {searchDisabled}
  {filterGroups}
  {onFilterChange}
  {onClearAllFilters}
  {sortDirection}
  onSortToggle={handleSortToggle}
  {showSort}
  {showViewToggle}
  bind:viewMode
  {onViewModeChange}
/>
