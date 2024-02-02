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
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getFilterSearchList, useMetricsView } from "../selectors/index";
  import { getStateManagers } from "../state-managers/state-managers";
  import FilterButton from "./FilterButton.svelte";

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        toggleDimensionFilterMode,
        removeDimensionFilter,
      },
      measuresFilter: { setMeasureFilter, removeMeasureFilter },
      filters: { clearAllFilters },
    },
    selectors: {
      dimensionFilters: { getDimensionFilterItems, getAllDimensionFilterItems },
      measureFilters: { getMeasureFilterItems, getAllMeasureFilterItems },
    },
  } = StateManagers;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";

  const metaQuery = useMetricsView(StateManagers);
  $: dimensions = $metaQuery.data?.dimensions ?? [];
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name as string,
  );
  $: measures = $metaQuery.data?.measures ?? [];
  $: measureIdMap = getMapFromArray(measures, (m) => m.name as string);

  let searchText = "";
  let allValues: Record<string, string[]> = {};
  let activeDimensionName: string;
  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;

  $: activeColumn =
    dimensions.find((d) => d.name === activeDimensionName)?.column ??
    activeDimensionName;

  $: if (activeDimensionName && dimensionIdMap.has(activeDimensionName)) {
    topListQuery = getFilterSearchList(StateManagers, {
      dimension: activeDimensionName,
      searchText,
      addNull: "null".includes(searchText),
    });
  }

  $: if (!$topListQuery?.isFetching) {
    const topListData = $topListQuery?.data?.data ?? [];
    allValues[activeDimensionName] =
      topListData.map((datum) => datum[activeColumn]) ?? [];
  }

  $: currentDimensionFilters = $getDimensionFilterItems(dimensionIdMap);
  $: allDimensionFilters = $getAllDimensionFilterItems(
    currentDimensionFilters,
    dimensionIdMap,
  );

  $: currentMeasureFilters = $getMeasureFilterItems(measureIdMap);
  $: allMeasureFilters = $getAllMeasureFilterItems(
    currentMeasureFilters,
    measureIdMap,
  );

  // hasFilter only checks for complete filters and excludes temporary ones
  $: hasFilters =
    currentDimensionFilters.length > 0 || currentMeasureFilters.length > 0;

  function setActiveDimension(name: string, value = "") {
    activeDimensionName = name;
    searchText = value;
  }

  function getColorForChip(isInclude: boolean) {
    return isInclude ? defaultChipColors : excludeChipColors;
  }

  function handleMeasureFilterApply(
    dimension: string,
    measureName: string,
    oldDimension: string,
    expr: V1Expression,
  ) {
    if (oldDimension && oldDimension !== dimension) {
      removeMeasureFilter(oldDimension, measureName);
    }
    setMeasureFilter(dimension, measureName, expr);
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
    {#if !allDimensionFilters.length && !allMeasureFilters.length}
      <div
        in:fly|local={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid items-center"
        style:min-height={ROW_HEIGHT}
      >
        No filters selected
      </div>
    {:else}
      {#each allDimensionFilters as { name, label, selectedValues } (name)}
        {@const isInclude =
          !$dashboardStore.dimensionFilterExcludeMode.get(name)}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => toggleDimensionFilterMode(name)}
            on:remove={() => {
              removeDimensionFilter(name);
            }}
            on:apply={(event) => {
              toggleDimensionValueSelection(name, event.detail, true);
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
            excludeMode={!isInclude}
            colors={getColorForChip(isInclude)}
            label="View filter"
            {selectedValues}
            allValues={allValues[activeDimensionName]}
          >
            <svelte:fragment slot="body-tooltip-content">
              Click to edit the the filters in this dimension
            </svelte:fragment>
          </RemovableListChip>
        </div>
      {/each}
      {#each allMeasureFilters as { name, label, dimensionName, expr } (name)}
        <div animate:flip={{ duration: 200 }}>
          <MeasureFilter
            {name}
            {label}
            {dimensionName}
            {expr}
            on:remove={() => removeMeasureFilter(dimensionName, name)}
            on:apply={({ detail: { dimension, oldDimension, expr } }) =>
              handleMeasureFilterApply(dimension, name, oldDimension, expr)}
          />
        </div>
      {/each}
    {/if}
    <FilterButton
      on:focus={({ detail: { name } }) => {
        setActiveDimension(name);
      }}
      on:hover={({ detail: { name } }) => {
        setActiveDimension(name);
      }}
    />
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
          on:click={clearAllFilters}
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
