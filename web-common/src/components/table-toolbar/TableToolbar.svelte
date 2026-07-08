<script lang="ts">
  import TableToolbarAppliedFilters from "./TableToolbarAppliedFilters.svelte";
  import TableToolbarFilterDropdown from "./TableToolbarFilterDropdown.svelte";
  import TableToolbarSearch from "./TableToolbarSearch.svelte";
  import TableToolbarSort from "./TableToolbarSort.svelte";
  import TableToolbarViewToggle from "./TableToolbarViewToggle.svelte";
  import type { FilterGroup, SortDirection, ViewMode } from "./types";
  import type { Snippet } from "svelte";
  import type { RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  let {
    searchTextStore,
    filterGroups = [],
    sortDirectionStore,
    viewModeStore,
    onClearAllFilters,
    children,
  }: {
    searchTextStore?: RuneStore<string>;
    filterGroups?: FilterGroup[];
    sortDirectionStore?: RuneStore<SortDirection>;
    viewModeStore?: RuneStore<ViewMode>;
    onClearAllFilters?: () => void;
    children?: Snippet;
  } = $props();
</script>

<section class="flex flex-col w-full">
  <div class="flex flex-row items-center justify-between h-9 gap-x-4">
    <div class="flex flex-row items-center">
      <TableToolbarFilterDropdown {filterGroups} />
    </div>

    <div class="flex flex-row items-center gap-x-3">
      <TableToolbarSearch {searchTextStore} />

      {#if sortDirectionStore}
        <TableToolbarSort {sortDirectionStore} />
      {/if}

      {#if viewModeStore}
        <TableToolbarViewToggle {viewModeStore} />
      {/if}

      {@render children?.()}
    </div>
  </div>

  <TableToolbarAppliedFilters {filterGroups} {onClearAllFilters} />
</section>
