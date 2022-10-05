<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { flip } from "svelte/animate";

  import type {
    MetricsViewDimensionValues,
    MetricsViewRequestFilter,
  } from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
  import { getMapFromArray } from "@rilldata/web-local/common/utils/arrayUtils";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import { Chip, ChipContainer, RemovableListChip } from "../../../chip";
  import Filter from "../../../icons/Filter.svelte";
  import FilterRemove from "../../../icons/FilterRemove.svelte";
  import { getDimensionsByMetricsId } from "../../../../redux-store/dimension-definition/dimension-definition-readables";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import { getDisplayName } from "../utils";
  export let metricsDefId;

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

  $: hasFilters = isFiltered(metricsExplorer?.filters);

  function clearAllFilters() {
    if (hasFilters) {
      metricsExplorerStore.clearFilters(metricsDefId);
    }
  }

  /** prune the values and prepare for for templating */
  let currentDimensionIncludeFilters = [];
  let currentDimensionExcludeFilters = [];
  $: if (includeValues && excludeValues) {
    const dimensionIdMap = getMapFromArray(
      $dimensions,
      (dimension) => dimension.id
    );
    currentDimensionIncludeFilters = includeValues.map((dimensionValues) => ({
      name: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
      dimensionId: dimensionValues.name,
      selectedValues: dimensionValues.values,
    }));
    currentDimensionExcludeFilters = excludeValues.map((dimensionValues) => ({
      name: getDisplayName(dimensionIdMap.get(dimensionValues.name)),
      dimensionId: dimensionValues.name,
      selectedValues: dimensionValues.values,
    }));
  }

  function toggleDimensionValue(event, item, include: boolean) {
    event.detail.forEach((dimensionValue) => {
      metricsExplorerStore.toggleFilter(
        metricsDefId,
        item.dimensionId,
        dimensionValue,
        include
      );
    });
  }

  const excludeChipColors = {
    bgBaseColor: "bg-gray-100",
    bgHoverColor: "bg-gray-200",
    textColor: "text-gray-900",
    bgActiveColor: "bg-gray-200",
    ringColor: "ring-gray-400",
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
    class:text-gray-400={!hasFilters}
    class:text-gray-600={hasFilters}
  >
    <Filter size="16px" />
  </div>
  {#if currentDimensionIncludeFilters?.length || currentDimensionExcludeFilters?.length}
    <ChipContainer>
      {#each currentDimensionIncludeFilters as { name, dimensionId, selectedValues } (dimensionId)}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:remove={() => clearFilterForDimension(dimensionId, true)}
            on:apply={(event) =>
              toggleDimensionValue(event, { dimensionId }, true)}
            typeLabel="dimension"
            {name}
            {selectedValues}
          >
            <svelte:fragment slot="body-tooltip-content">
              click to edit the the filters in this dimension
            </svelte:fragment>
          </RemovableListChip>
        </div>
      {/each}
      {#each currentDimensionExcludeFilters as { name, dimensionId, selectedValues } (dimensionId)}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:remove={() => clearFilterForDimension(dimensionId, false)}
            on:apply={(event) =>
              toggleDimensionValue(event, { dimensionId }, false)}
            typeLabel="dimension"
            name={`Exclude ${name}`}
            {selectedValues}
            excludeMode
            colors={excludeChipColors}
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
            bgBaseColor="bg-white"
            bgHoverColor="bg-gray-100"
            textColor="text-gray-500"
            bgActiveColor="bg-gray-200"
            ringColor="ring-gray-400"
            on:click={clearAllFilters}
          >
            <svelte:fragment slot="icon"
              ><FilterRemove size="16px" /></svelte:fragment
            >
            <svelte:fragment slot="body">clear filters</svelte:fragment>
          </Chip>
        </div>
      {/if}
    </ChipContainer>
  {:else if currentDimensionIncludeFilters?.length === 0 && currentDimensionExcludeFilters?.length === 0}
    <div
      in:fly|local={{ duration: 200, x: 8 }}
      class="italic text-gray-400  grid items-center"
      style:min-height="26px"
    >
      no filters selected
    </div>
  {:else}
    &nbsp;
  {/if}
</section>
