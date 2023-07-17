<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import {
    Chip,
    ChipContainer,
    RemovableListChip,
  } from "@rilldata/web-common/components/chip";
  import {
    defaultChipColors,
    excludeChipColors,
    includeHiddenChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterRemove from "@rilldata/web-common/components/icons/FilterRemove.svelte";
  import { useMetaQuery, getFilterSearchList } from "../selectors/index";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type { V1MetricsViewFilter } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getDisplayName } from "./getDisplayName";
  import { getStateManagers } from "../state-managers/state-managers";
  import {
    clearAllFilters,
    clearFilterForDimension,
    toggleDimensionValue,
    toggleFilterMode,
  } from "../actions";

  const StateManagers = getStateManagers();
  const { dashboardStore } = StateManagers;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";

  const metaQuery = useMetaQuery(StateManagers);
  $: dimensions = $metaQuery.data?.dimensions ?? [];

  function isFiltered(filters: V1MetricsViewFilter): boolean {
    if (!filters) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }

  let searchText = "";
  let searchedValues: string[] | null = null;
  let activeDimensionName;
  $: activeColumn =
    dimensions.find((d) => d.name === activeDimensionName)?.column ??
    activeDimensionName;

  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;
  $: if (activeDimensionName)
    topListQuery = getFilterSearchList(StateManagers, {
      dimension: activeDimensionName,
      searchText,
      addNull: "null".includes(searchText),
    });

  $: if (!$topListQuery?.isFetching && searchText != "") {
    const topListData = $topListQuery?.data?.data ?? [];
    searchedValues = topListData.map((datum) => datum[activeColumn]) ?? [];
  }

  $: hasFilters = isFiltered($dashboardStore.filters);

  /** prune the values and prepare for templating */
  let currentDimensionFilters: {
    name: string;
    label: string;
    selectedValues: any[];
    filterType: string;
    isHidden: boolean;
  }[] = [];

  $: {
    const dimensionIdMap = getMapFromArray(
      dimensions,
      (dimension) => dimension.name
    );
    const currentDimensionIncludeFilters = $dashboardStore.filters.include.map(
      (dimensionValues) => ({
        name: dimensionValues.name,
        label: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
        selectedValues: dimensionValues.in,
        filterType: "include",
        isHidden: !$dashboardStore?.visibleDimensionKeys.has(
          dimensionValues.name
        ),
      })
    );
    const currentDimensionExcludeFilters = $dashboardStore.filters.exclude.map(
      (dimensionValues) => ({
        name: dimensionValues.name,
        label: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
        selectedValues: dimensionValues.in,
        filterType: "exclude",
        isHidden: !$dashboardStore?.visibleDimensionKeys.has(
          dimensionValues.name
        ),
      })
    );
    currentDimensionFilters = [
      ...currentDimensionIncludeFilters,
      ...currentDimensionExcludeFilters,
    ];
    // sort based on name to make sure toggling include/exclude is not jarring
    currentDimensionFilters.sort((a, b) => (a.name > b.name ? 1 : -1));
  }

  function setActiveDimension(name, value) {
    activeDimensionName = name;
    searchText = value;
  }

  function getColorForChip(isHidden, isInclude) {
    if (isInclude) {
      return isHidden ? includeHiddenChipColors : defaultChipColors;
    }
    return excludeChipColors;
  }
</script>

<section
  class="pl-2 grid gap-x-2 items-start"
  style:grid-template-columns="max-content auto"
  style:min-height={MIN_CONTAINER_HEIGHT}
>
  <div
    style:height={ROW_HEIGHT}
    style:width={ROW_HEIGHT}
    class="grid items-center place-items-center"
    class:ui-copy-icon-inactive={!hasFilters}
    class:ui-copy-icon={hasFilters}
  >
    <Filter size="16px" />
  </div>
  {#if currentDimensionFilters.length > 0}
    <ChipContainer>
      {#each currentDimensionFilters as { name, label, selectedValues, filterType, isHidden } (name)}
        {@const isInclude = filterType === "include"}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => toggleFilterMode(StateManagers, name)}
            on:remove={() =>
              clearFilterForDimension(
                StateManagers,
                name,
                isInclude ? true : false
              )}
            on:apply={(event) =>
              toggleDimensionValue(StateManagers, name, event.detail)}
            on:search={(event) => {
              setActiveDimension(name, event.detail);
            }}
            typeLabel={isHidden ? "hidden dimension" : "dimension"}
            name={isInclude ? label : `Exclude ${label}`}
            excludeMode={isInclude ? false : true}
            colors={getColorForChip(isHidden, isInclude)}
            label="View filter"
            {selectedValues}
            {searchedValues}
            {isHidden}
          >
            <svelte:fragment slot="body-tooltip-content">
              {#if isHidden}
                To show, use the dimension selector below.
              {:else}
                Click to edit the the filters in this dimension
              {/if}
            </svelte:fragment>
          </RemovableListChip>
        </div>
      {/each}
      <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
      {#if hasFilters}
        <div class="ml-auto">
          <Chip
            bgBaseClass="surface"
            bgHoverClass="hover:bg-gray-100 hover:dark:bg-gray-700"
            textClass="ui-copy-disabled-faint hover:text-gray-500 dark:text-gray-500"
            bgActiveClass="bg-gray-200 dark:bg-gray-600"
            outlineClass="outline-gray-400"
            on:click={() => clearAllFilters(StateManagers)}
          >
            <span slot="icon" class="ui-copy-disabled-faint">
              <FilterRemove size="16px" />
            </span>
            <svelte:fragment slot="body">Clear filters</svelte:fragment>
          </Chip>
        </div>
      {/if}
    </ChipContainer>
  {:else if currentDimensionFilters.length === 0}
    <div
      in:fly|local={{ duration: 200, x: 8 }}
      class="ui-copy-disabled grid items-center"
      style:min-height={ROW_HEIGHT}
    >
      No filters selected
    </div>
  {:else}
    &nbsp;
  {/if}
</section>
