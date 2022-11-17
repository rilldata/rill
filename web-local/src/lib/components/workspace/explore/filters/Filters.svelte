<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";

  import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { getContext } from "svelte";
  import { flip } from "svelte/animate";

  import type {
    MetricsViewDimensionValues,
    MetricsViewRequestFilter,
  } from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
  import { getMapFromArray } from "@rilldata/web-local/common/utils/arrayUtils";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import { getDimensionsByMetricsId } from "../../../../redux-store/dimension-definition/dimension-definition-readables";
  import { useTopListQuery } from "../../../../svelte-query/queries/metrics-views/top-list";
  import { Chip, ChipContainer, RemovableListChip } from "../../../chip";
  import { defaultChipColors } from "../../../chip/chip-types";
  import Filter from "../../../icons/Filter.svelte";
  import FilterRemove from "../../../icons/FilterRemove.svelte";
  import { getDisplayName } from "../utils";
  export let metricsDefId;

  const config = getContext<RootConfig>("config");

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  let includeValues: MetricsViewDimensionValues;
  $: includeValues = metricsExplorer?.filters.include;
  let excludeValues: MetricsViewDimensionValues;
  $: excludeValues = metricsExplorer?.filters.exclude;

  let dimensions: Readable<DimensionDefinitionEntity[]>;
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  function clearFilterForDimension(dimensionId, include: boolean) {
    metricsExplorerStore.clearFilterForDimension(
      metricsDefId,
      dimensionId,
      include
    );
  }

  function isFiltered(filters: MetricsViewRequestFilter): boolean {
    if (!filters) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }

  let topListQuery;
  let searchText = "";
  let searchedValues = [];
  let activeDimensionName;
  let activeDimensionId;

  $: addNull = "null".includes(searchText);

  $: if (activeDimensionName && activeDimensionId) {
    if (searchText == "") {
      searchedValues = [];
    } else {
      // Use topList API to fetch the dimension names
      // We prune the measure values and use the dimension labels for the filter
      topListQuery = useTopListQuery(config, metricsDefId, activeDimensionId, {
        measures: ["measure_0"], // Ideally should work with empty measures
        limit: 15,
        offset: 0,
        sort: [],
        time: {
          start: metricsExplorer?.selectedTimeRange?.start,
          end: metricsExplorer?.selectedTimeRange?.end,
        },
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
      });
    }
  }

  function setActiveDimension(name, dimensionId, value) {
    activeDimensionName = name;
    activeDimensionId = dimensionId;
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
      metricsExplorerStore.clearFilters(metricsDefId);
    }
  }

  /** prune the values and prepare for for templating */
  let currentDimensionFilters = [];
  $: if (includeValues && excludeValues) {
    const dimensionIdMap = getMapFromArray(
      $dimensions,
      (dimension) => dimension.id
    );
    const currentDimensionIncludeFilters = includeValues.map(
      (dimensionValues) => ({
        name: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
        sqlName: dimensionIdMap.get(dimensionValues.name)?.dimensionColumn,
        dimensionId: dimensionValues.name,
        selectedValues: dimensionValues.in,
        filterType: "include",
      })
    );
    const currentDimensionExcludeFilters = excludeValues.map(
      (dimensionValues) => ({
        name: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
        sqlName: dimensionIdMap.get(dimensionValues.name)?.dimensionColumn,
        dimensionId: dimensionValues.name,
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
    metricsExplorerStore.toggleFilter(
      metricsDefId,
      item.dimensionId,
      event.detail
    );
  }

  function togglerFilterMode(dimensionId) {
    metricsExplorerStore.toggleFilterMode(metricsDefId, dimensionId);
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
      {#each currentDimensionFilters as { name, sqlName, dimensionId, selectedValues, filterType } (dimensionId)}
        {@const isInclude = filterType === "include"}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:toggle={() => togglerFilterMode(dimensionId)}
            on:remove={() =>
              clearFilterForDimension(dimensionId, isInclude ? true : false)}
            on:apply={(event) => toggleDimensionValue(event, { dimensionId })}
            on:search={(event) => {
              setActiveDimension(sqlName, dimensionId, event.detail);
            }}
            typeLabel="dimension"
            name={isInclude ? name : `Exclude ${name}`}
            excludeMode={isInclude ? false : true}
            colors={isInclude ? defaultChipColors : excludeChipColors}
            {selectedValues}
            {searchedValues}
          >
            <svelte:fragment slot="body-tooltip-content">
              click to edit the the filters in this dimension
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
            <svelte:fragment slot="body">clear filters</svelte:fragment>
          </Chip>
        </div>
      {/if}
    </ChipContainer>
  {:else if currentDimensionFilters?.length === 0}
    <div
      in:fly|local={{ duration: 200, x: 8 }}
      class="italic ui-copy-disabled grid items-center"
      style:min-height="26px"
    >
      no filters selected
    </div>
  {:else}
    &nbsp;
  {/if}
</section>
