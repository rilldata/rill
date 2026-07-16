<script lang="ts">
  import TableToolbarAppliedFilters from "./TableToolbarAppliedFilters.svelte";
  import TableToolbarFilterDropdown from "./TableToolbarFilterDropdown.svelte";
  import TableToolbarSearch from "./TableToolbarSearch.svelte";
  import TableToolbarSort from "./TableToolbarSort.svelte";
  import TableToolbarViewToggle from "./TableToolbarViewToggle.svelte";
  import type { FilterGroup, SortOption, ViewMode } from "./types";
  import type { Snippet } from "svelte";
  import {
    ArrayRuneStore,
    type RuneStore,
  } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import type { ColumnSort } from "tanstack-table-8-svelte-5";

  let {
    searchTextStore,
    filterGroups = [],
    sortStore,
    sortOptions,
    viewModeStore,
    onClearAllFilters,
    children,
  }: {
    searchTextStore?: RuneStore<string>;
    filterGroups?: FilterGroup[];
    sortStore?: ArrayRuneStore<ColumnSort>;
    sortOptions?: SortOption[];
    viewModeStore?: RuneStore<ViewMode>;
    onClearAllFilters?: () => void;
    children?: Snippet;
  } = $props();
</script>

<section class="flex flex-col w-full gap-y-2">
  <div class="flex flex-row items-center gap-x-2.5">
    <div class="flex flex-row items-center gap-x-2.5 shrink-0">
      <TableToolbarFilterDropdown {filterGroups} />

      {#if sortStore && sortOptions}
        <TableToolbarSort {sortStore} {sortOptions} />
      {/if}
    </div>

    <TableToolbarSearch {searchTextStore} />

    <div class="flex flex-row items-center gap-x-2.5 shrink-0">
      {#if viewModeStore}
        <TableToolbarViewToggle {viewModeStore} />
      {/if}

      {@render children?.()}
    </div>
  </div>

  <TableToolbarAppliedFilters {filterGroups} {onClearAllFilters} />
</section>
