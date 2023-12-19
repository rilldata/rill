<!-- @component
The main feature-set component for dashboard filters
 -->
<script context="module" lang="ts">
  import { writable } from "svelte/store";
  import {
    Chip,
    ChipContainer,
    RemovableListChip,
  } from "@rilldata/web-common/components/chip";
  import {
    defaultChipColors,
    excludeChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterRemove from "@rilldata/web-common/components/icons/FilterRemove.svelte";
  import { useMetaQuery, getFilterSearchList } from "../selectors/index";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type {
    MetricsViewSpecDimensionV2,
    V1MetricsViewFilter,
  } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getDisplayName } from "./getDisplayName";
  import { getStateManagers } from "../state-managers/state-managers";
  import { clearAllFilters, toggleFilterMode } from "../actions";
  import FilterButton from "./FilterButton.svelte";
  import { formatFilters } from "./formatFilters";

  export const potentialFilterName = writable<string | null>(null);
</script>

<script lang="ts">
  const StateManagers = getStateManagers();

  const {
    dashboardStore,
    selectors: {
      dimensionFilters: { isFilterExcludeMode },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionNameSelection,
        toggleDimensionValueSelection,
      },
    },
  } = StateManagers;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";

  const metaQuery = useMetaQuery(StateManagers);
  $: dimensions = $metaQuery.data?.dimensions ?? [];

  let searchText = "";
  let allValues: string[] | null = null;
  let activeDimensionName: string;
  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;

  $: filters = $dashboardStore.filters;

  $: activeColumn =
    dimensions.find((d) => d.name === activeDimensionName)?.column ??
    activeDimensionName;

  $: if (activeDimensionName) {
    topListQuery = getFilterSearchList(StateManagers, {
      dimension: activeDimensionName,
      searchText,
      addNull: "null".includes(searchText),
    });
  }

  $: if (!$topListQuery?.isFetching) {
    const topListData = $topListQuery?.data?.data ?? [];
    allValues = topListData.map((datum) => datum[activeColumn]) ?? [];
  }

  $: hasFilters = isFiltered($dashboardStore.filters);

  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name
  );

  $: currentDimensionIncludeFilters = formatFilters(
    filters.include,
    false,
    dimensionIdMap
  );

  $: currentDimensionExcludeFilters = formatFilters(
    filters.exclude,
    true,
    dimensionIdMap
  );

  $: temporaryFilter = $potentialFilterName
    ? [
        {
          name: $potentialFilterName,
          label: getDisplayName(
            dimensionIdMap.get(
              $potentialFilterName
            ) as MetricsViewSpecDimensionV2
          ),
          selectedValues: [],
          filterType: $isFilterExcludeMode($potentialFilterName)
            ? "exclude"
            : "include",
        },
      ]
    : [];

  $: currentDimensionFilters = [
    ...currentDimensionExcludeFilters,
    ...currentDimensionIncludeFilters,
    ...temporaryFilter,
  ].sort((a, b) => (a.name > b.name ? 1 : -1));

  function setActiveDimension(name: string, value = "") {
    activeDimensionName = name;
    searchText = value;
  }

  function getColorForChip(isInclude: boolean) {
    return isInclude ? defaultChipColors : excludeChipColors;
  }

  function isFiltered(filters: V1MetricsViewFilter): boolean {
    if (!filters || !filters.include || !filters.exclude) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }
</script>

<section
  class="pl-2 grid gap-x-2 items-start"
  style:grid-template-columns="max-content auto"
  style:min-height={MIN_CONTAINER_HEIGHT}
>
  <div
    class="grid items-center place-items-center"
    class:ui-copy-icon={hasFilters}
    class:ui-copy-icon-inactive={!hasFilters}
    style:height={ROW_HEIGHT}
    style:width={ROW_HEIGHT}
  >
    <Filter size="16px" />
  </div>

  <ChipContainer>
    {#if !currentDimensionFilters.length}
      <div
        in:fly|local={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid items-center"
        style:min-height={ROW_HEIGHT}
      >
        No filters selected
      </div>
    {:else}
      {#each currentDimensionFilters as { name, label, selectedValues, filterType } (name)}
        {@const isInclude = filterType === "include"}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => toggleFilterMode(StateManagers, name)}
            on:remove={() => {
              if ($potentialFilterName === name) {
                $potentialFilterName = null;
              } else {
                toggleDimensionNameSelection(name);
              }
            }}
            on:apply={(event) => {
              if ($potentialFilterName === name) {
                $potentialFilterName = null;
              }
              toggleDimensionValueSelection(name, event.detail);
            }}
            on:search={(event) => {
              setActiveDimension(name, event.detail);
            }}
            on:click={() => {
              setActiveDimension(name, "");
            }}
            on:mount={() => {
              setActiveDimension(name);
            }}
            typeLabel="dimension"
            name={isInclude ? label : `Exclude ${label}`}
            excludeMode={isInclude ? false : true}
            colors={getColorForChip(isInclude)}
            label="View filter"
            {selectedValues}
            {allValues}
          >
            <svelte:fragment slot="body-tooltip-content">
              Click to edit the the filters in this dimension
            </svelte:fragment>
          </RemovableListChip>
        </div>
      {/each}
    {/if}
    <FilterButton />
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
</section>
