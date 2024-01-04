<!-- @component
The main feature-set component for dashboard filters
 -->
<script context="module" lang="ts">
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
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { clearAllFilters } from "../actions";
  import { getFilterSearchList, useMetaQuery } from "../selectors/index";
  import { getStateManagers } from "../state-managers/state-managers";
  import FilterButton from "./FilterButton.svelte";
  import {
    getDimensionFilterItems,
    getMeasureFilterItems,
    potentialFilterName,
  } from "@rilldata/web-common/features/dashboards/filters/filter-items";
</script>

<script lang="ts">
  import { measureChipColors } from "@rilldata/web-common/components/chip/chip-types";

  const StateManagers = getStateManagers();

  const {
    dashboardStore,
    actions: {
      dimensionsFilter: {
        toggleDimensionNameSelection,
        toggleDimensionValueSelection,
        toggleDimensionFilterMode,
      },
      measuresFilter: { toggleMeasureFilter, setMeasureFilter },
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
  let allValues: Record<string, string[]> = {};
  let activeDimensionName: string;
  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;

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
    allValues[activeDimensionName] =
      topListData.map((datum) => datum[activeColumn]) ?? [];
  }

  $: dimensionIdMap = getMapFromArray(dimensions, (d) => d.name as string);

  $: currentDimensionFilters = getDimensionFilterItems(
    $dashboardStore.whereFilter,
    dimensionIdMap,
    $potentialFilterName
  );

  $: currentMeasureFilters = getMeasureFilterItems(
    $dashboardStore.havingFilter,
    getMapFromArray(measures, (m) => m.name as string),
    $potentialFilterName
  );

  $: hasFilters =
    currentDimensionFilters.length > 0 || currentMeasureFilters.length > 0;

  function setActiveDimension(name: string, value = "") {
    if (!dimensionIdMap.has(name)) return;
    activeDimensionName = name;
    searchText = value;
  }

  function getColorForChip(isInclude: boolean) {
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
    {#if !hasFilters}
      <div
        in:fly|local={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid items-center"
        style:min-height={ROW_HEIGHT}
      >
        No filters selected
      </div>
    {:else}
      {#each currentDimensionFilters as { name, label, selectedValues } (name)}
        {@const isInclude =
          !$dashboardStore.dimensionFilterExcludeMode.get(name)}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => toggleDimensionFilterMode(name)}
            on:remove={() => {
              if ($potentialFilterName) {
                $potentialFilterName = null;
              } else {
                toggleDimensionNameSelection(name);
              }
            }}
            on:apply={(event) => {
              if ($potentialFilterName) {
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
      {#each currentMeasureFilters as { name, label, expr } (name)}
        <div animate:flip={{ duration: 200 }}>
          <MeasureFilter
            on:remove={() => {
              if ($potentialFilterName) {
                $potentialFilterName = null;
              } else {
                toggleMeasureFilter(name);
              }
            }}
            on:apply={(event) => {
              if ($potentialFilterName) {
                $potentialFilterName = null;
              }
              setMeasureFilter(name, event.detail);
            }}
            colors={measureChipColors}
            {name}
            {label}
            {expr}
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
          on:click={() => {
            $potentialFilterName = null;
            clearAllFilters(StateManagers);
          }}
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
