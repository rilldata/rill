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
  import {
    clearAllFilters,
    toggleDimensionValue,
    toggleFilterMode,
  } from "../actions";
  import FilterButton from "./FilterButton.svelte";

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    actions: {
      dimensionsFilter: { toggleDimensionNameSelection },
    },
  } = StateManagers;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";

  const metaQuery = useMetaQuery(StateManagers);
  $: dimensions = $metaQuery.data?.dimensions ?? [];

  function isFiltered(filters: V1MetricsViewFilter): boolean {
    if (!filters || !filters.include || !filters.exclude) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }

  let searchText = "";
  let allValues: string[] | null = null;
  let activeDimensionName: string;

  $: activeColumn =
    dimensions.find((d) => d.name === activeDimensionName)?.column ??
    activeDimensionName;

  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;
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

  /** prune the values and prepare for templating */
  let currentDimensionFilters: {
    name: string;
    label: string;
    selectedValues: any[];
    filterType: string;
  }[] = [];

  $: {
    const dimensionIdMap = getMapFromArray(
      dimensions,
      (dimension) => dimension.name
    );

    const currentDimensionIncludeFilters =
      $dashboardStore?.filters?.include
        ?.filter((dimensionValues) => dimensionValues.name !== undefined)
        .map((dimensionValues) => {
          const name = dimensionValues.name as string;
          return {
            name,
            label: getDisplayName(
              dimensionIdMap.get(name) as MetricsViewSpecDimensionV2
            ),
            selectedValues: dimensionValues.in as any[],
            filterType: "include",
          };
        }) ?? [];

    const currentDimensionExcludeFilters =
      $dashboardStore?.filters?.exclude
        ?.filter((dimensionValues) => dimensionValues.name !== undefined)
        .map((dimensionValues) => {
          const name = dimensionValues.name as string;
          return {
            name,
            label: getDisplayName(
              dimensionIdMap.get(name) as MetricsViewSpecDimensionV2
            ),
            selectedValues: dimensionValues.in as any[],
            filterType: "exclude",
          };
        }) ?? [];

    currentDimensionFilters = [
      ...currentDimensionIncludeFilters,
      ...currentDimensionExcludeFilters,
    ];
    // sort based on name to make sure toggling include/exclude is not jarring
    currentDimensionFilters.sort((a, b) => (a.name > b.name ? 1 : -1));
  }

  function setActiveDimension(name, value = "") {
    activeDimensionName = name;
    searchText = value;
  }

  function getColorForChip(isInclude) {
    return isInclude ? defaultChipColors : excludeChipColors;
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
            on:remove={() => toggleDimensionNameSelection(name)}
            on:apply={(event) => {
              toggleDimensionValue(StateManagers, name, event.detail);
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
