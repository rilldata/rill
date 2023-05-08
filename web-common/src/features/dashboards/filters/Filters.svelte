<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import {
    Chip,
    ChipContainer,
    RemovableListChip,
  } from "@rilldata/web-common/components/chip";
  import { defaultChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterRemove from "@rilldata/web-common/components/icons/FilterRemove.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import type {
    MetricsViewDimension,
    MetricsViewFilterCond,
    V1MetricsViewFilter,
  } from "@rilldata/web-common/runtime-client";
  import { createQueryServiceMetricsViewToplist } from "@rilldata/web-common/runtime-client";
  import { getMapFromArray } from "@rilldata/web-local/lib/util/arrayUtils";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import { getDisplayName } from "./getDisplayName";

  export let metricViewName;

  const queryClient = useQueryClient();

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let includeValues: Array<MetricsViewFilterCond>;
  $: includeValues = metricsExplorer?.filters.include;
  let excludeValues: Array<MetricsViewFilterCond>;
  $: excludeValues = metricsExplorer?.filters.exclude;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  let dimensions: Array<MetricsViewDimension>;
  $: dimensions = $metaQuery.data?.dimensions;

  function clearFilterForDimension(dimensionId, include: boolean) {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.clearFilterForDimension(
      metricViewName,
      dimensionId,
      include
    );
  }

  function isFiltered(filters: V1MetricsViewFilter): boolean {
    if (!filters) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }

  let topListQuery;
  let searchText = "";
  let searchedValues = [];
  let activeDimensionName;

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: addNull = "null".includes(searchText);

  $: if (activeDimensionName) {
    if (searchText == "") {
      searchedValues = [];
    } else {
      let topListParams = {
        dimensionName: activeDimensionName,
        limit: "15",
        offset: "0",
        sort: [],
        filter: {
          include: [
            {
              name: activeDimensionName,
              in: addNull ? [null] : [],
              like: [`%${searchText}%`],
            },
          ],
          exclude: [],
        },
      };

      if (hasTimeSeries) {
        topListParams = {
          ...topListParams,
          ...{
            timeStart: metricsExplorer?.selectedTimeRange?.start,
            timeEnd: metricsExplorer?.selectedTimeRange?.end,
          },
        };
      }

      // Use topList API to fetch the dimension names
      // We prune the measure values and use the dimension labels for the filter
      topListQuery = createQueryServiceMetricsViewToplist(
        $runtime.instanceId,
        metricViewName,
        topListParams
      );
    }
  }

  function setActiveDimension(name, value) {
    activeDimensionName = name;
    searchText = value;
  }

  $: if (!$topListQuery?.isFetching && searchText != "") {
    const topListData = $topListQuery?.data?.data ?? [];
    searchedValues =
      topListData.map((datum) => datum[activeDimensionName]) ?? [];
  }

  $: hasFilters = isFiltered(metricsExplorer?.filters);

  function clearAllFilters() {
    if (hasFilters) {
      cancelDashboardQueries(queryClient, metricViewName);
      metricsExplorerStore.clearFilters(metricViewName);
    }
  }

  /** prune the values and prepare for templating */
  let currentDimensionFilters = [];
  $: if (includeValues && excludeValues && dimensions) {
    const dimensionIdMap = getMapFromArray(
      dimensions,
      (dimension) => dimension.name
    );
    const currentDimensionIncludeFilters = includeValues.map(
      (dimensionValues) => ({
        name: dimensionValues.name,
        label: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
        selectedValues: dimensionValues.in,
        filterType: "include",
      })
    );
    const currentDimensionExcludeFilters = excludeValues.map(
      (dimensionValues) => ({
        name: dimensionValues.name,
        label: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
        selectedValues: dimensionValues.in,
        filterType: "exclude",
      })
    );
    currentDimensionFilters = [
      ...currentDimensionIncludeFilters,
      ...currentDimensionExcludeFilters,
    ];
    // sort based on name to make sure toggling include/exclude is not jarring
    currentDimensionFilters.sort((a, b) => (a.name > b.name ? 1 : -1));
  }

  function toggleDimensionValue(event, item) {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, item.name, event.detail);
  }

  function toggleFilterMode(dimensionName) {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
  }

  const excludeChipColors = {
    bgBaseClass: "bg-gray-100 dark:bg-gray-700",
    bgHoverClass: "bg-gray-200 dark:bg-gray-600",
    textClass: "ui-copy",
    bgActiveClass: "bg-gray-200 dark:bg-gray-600",
    outlineClass: "outline-gray-400 dark:outline-gray-500",
  };
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
  {#if currentDimensionFilters?.length}
    <ChipContainer>
      {#each currentDimensionFilters as { name, label, selectedValues, filterType } (name)}
        {@const isInclude = filterType === "include"}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => toggleFilterMode(name)}
            on:remove={() =>
              clearFilterForDimension(name, isInclude ? true : false)}
            on:apply={(event) => toggleDimensionValue(event, { name })}
            on:search={(event) => {
              setActiveDimension(name, event.detail);
            }}
            typeLabel="dimension"
            name={isInclude ? label : `Exclude ${label}`}
            excludeMode={isInclude ? false : true}
            colors={isInclude ? defaultChipColors : excludeChipColors}
            {selectedValues}
            {searchedValues}
          >
            <svelte:fragment slot="body-tooltip-content">
              Click to edit the the filters in this dimension
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
  {:else if currentDimensionFilters?.length === 0}
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
