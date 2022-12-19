<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterRemove from "@rilldata/web-common/components/icons/FilterRemove.svelte";
  import type {
    MetricsViewDimension,
    V1MetricsViewRequestFilter,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeServiceMetricsViewToplist } from "@rilldata/web-common/runtime-client";
  import type { MetricsViewDimensionValues } from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { getMapFromArray } from "@rilldata/web-local/lib/util/arrayUtils";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import { Chip, ChipContainer, RemovableListChip } from "../../../chip";
  import { defaultChipColors } from "../../../chip/chip-types";
  import { getDisplayName } from "../utils";

  export let metricViewName;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let includeValues: MetricsViewDimensionValues;
  $: includeValues = metricsExplorer?.filters.include;
  let excludeValues: MetricsViewDimensionValues;
  $: excludeValues = metricsExplorer?.filters.exclude;

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  let dimensions: Array<MetricsViewDimension>;
  $: dimensions = $metaQuery.data?.dimensions;

  function clearFilterForDimension(dimensionId, include: boolean) {
    metricsExplorerStore.clearFilterForDimension(
      metricViewName,
      dimensionId,
      include
    );
  }

  function isFiltered(filters: V1MetricsViewRequestFilter): boolean {
    if (!filters) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }

  let topListQuery;
  let searchText = "";
  let searchedValues = [];
  let activeDimensionName;

  $: addNull = "null".includes(searchText);

  $: if (activeDimensionName) {
    if (searchText == "") {
      searchedValues = [];
    } else {
      // Use topList API to fetch the dimension names
      // We prune the measure values and use the dimension labels for the filter
      topListQuery = useRuntimeServiceMetricsViewToplist(
        $runtimeStore.instanceId,
        metricViewName,
        {
          dimensionName: activeDimensionName,
          limit: "15",
          offset: "0",
          sort: [],
          timeStart: metricsExplorer?.selectedTimeRange?.start,
          timeEnd: metricsExplorer?.selectedTimeRange?.end,
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
        }
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
      metricsExplorerStore.clearFilters(metricViewName);
    }
  }

  /** prune the values and prepare for templating */
  let currentDimensionFilters = [];
  $: if (includeValues && excludeValues) {
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
  }

  function toggleDimensionValue(event, item) {
    metricsExplorerStore.toggleFilter(metricViewName, item.name, event.detail);
  }

  function toggleFilterMode(dimensionName) {
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
  class="pl-2 pt-2 pb-3 grid gap-x-2"
  style:grid-template-columns="max-content auto"
  style:min-height="44px"
>
  <div
    style:width="24px"
    style:height="24px"
    class="grid place-items-center"
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
      style:min-height="26px"
    >
      No filters selected
    </div>
  {:else}
    &nbsp;
  {/if}
</section>
