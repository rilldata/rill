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
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filter/MeasureFilter.svelte";
  import type { FilteredDimension } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import type { FilteredMeasure } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
  import { useMetaQuery, getFilterSearchList } from "../selectors/index";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import { clearAllFilters } from "../actions";
  import FilterButton from "./FilterButton.svelte";

  const StateManagers = getStateManagers();
  const {
    selectors: {
      dimensionFilters: { dimensionHasFilter, getAllFilteredDimensions },
      measureFilters: { measureHasFilter, getAllMeasureFilters },
    },
    actions: {
      dimensionsFilter: {
        toggleDimensionNameSelection,
        toggleDimensionValueSelection,
        toggleDimensionFilterMode,
      },
      measuresFilter: { toggleMeasureFilter },
    },
  } = StateManagers;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";

  const metaQuery = useMetaQuery(StateManagers);
  $: dimensions = $metaQuery.data?.dimensions ?? [];
  $: measures = $metaQuery.data?.measures ?? [];

  let searchText = "";
  let allValues: string[] | null = null;
  let activeDimensionName: string;
  let activeMeasureName: string;

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

  $: hasFilters =
    $dimensionHasFilter(activeDimensionName) ||
    $measureHasFilter(activeMeasureName);

  /** prune the values and prepare for templating */
  let currentDimensionFilters: FilteredDimension[] = [];

  $: {
    const dimensionIdMap = getMapFromArray(
      dimensions,
      (dimension) => dimension.name as string
    );
    currentDimensionFilters = $getAllFilteredDimensions(dimensionIdMap);
    // sort based on name to make sure toggling include/exclude is not jarring
    currentDimensionFilters.sort((a, b) => (a.name > b.name ? 1 : -1));
  }

  let currentMeasureFilters: FilteredMeasure[] = [];

  $: {
    const measureIdMap = getMapFromArray(
      measures,
      (measure) => measure.name as string
    );
    currentMeasureFilters = $getAllMeasureFilters(measureIdMap);
    // sort based on name to make sure toggling include/exclude is not jarring
    currentMeasureFilters.sort((a, b) => (a.name > b.name ? 1 : -1));
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
    {#if !currentDimensionFilters.length && !currentMeasureFilters.length}
      <div
        in:fly|local={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid items-center"
        style:min-height={ROW_HEIGHT}
      >
        No filters selected
      </div>
    {:else}
      {#each currentDimensionFilters as { name, label, selectedValues, isInclude } (name)}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => toggleDimensionFilterMode(name)}
            on:remove={() => toggleDimensionNameSelection(name)}
            on:apply={(event) => {
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
            excludeMode={!isInclude}
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
      {#each currentMeasureFilters as { name, label, expr } (name)}
        <div animate:flip={{ duration: 200 }}>
          <MeasureFilter
            on:remove={() => toggleMeasureFilter(name, expr)}
            colors={defaultChipColors}
            {name}
            {label}
            {expr}
          />
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
