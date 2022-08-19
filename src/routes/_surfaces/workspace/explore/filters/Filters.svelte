<script lang="ts">
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

  import Filter from "$lib/components/icons/Filter.svelte";
  import FilterRemove from "$lib/components/icons/FilterRemove.svelte";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import {
    clearSelectedDimensionLeaderboardAndUpdate,
    clearSelectedLeaderboardValuesAndUpdate,
    toggleSelectedLeaderboardValueAndUpdate,
  } from "$lib/redux-store/explore/explore-apis";
  import type { LeaderboardValues } from "$lib/redux-store/explore/explore-slice";
  import { store } from "$lib/redux-store/store-root";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import FilterContainer from "./FilterContainer.svelte";
  import FilterSet from "./FilterSet.svelte";
  export let metricsDefId;
  export let values;

  let dimensions: Readable<DimensionDefinitionEntity[]>;
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }

  function clearFilterForDimension(dimension) {
    clearSelectedDimensionLeaderboardAndUpdate(
      store.dispatch,
      metricsDefId,
      dimension
    );
  }

  $: hasFilters = isAnythingSelected(values);

  function pruneValues(set) {
    if (!set) return;
    return Object.keys(set)
      .filter((key) => set[key].length)
      .map((key) => {
        return [
          key,
          set[key].filter(([k, v]) => v === true).map(([k, v]) => k),
        ];
      });
  }

  $: prunedValues = pruneValues(values);

  function onSelectItem(event, item: LeaderboardValues) {
    toggleSelectedLeaderboardValueAndUpdate(
      store.dispatch,
      metricsDefId,
      item.dimensionId,
      event.detail,
      true
    );
  }
</script>

<div class="pt-3 pb-3" style:min-height="50px">
  <FilterContainer>
    <div
      class="grid place-items-center"
      style:width="24px"
      style:height="24px"
      style:font-size="18px"
    >
      <Filter />
    </div>

    {#if prunedValues?.length && $dimensions?.length}
      {#each prunedValues as [dimensionId, selectedValues]}
        {@const name = $dimensions.find(
          (dim) => dim.id === dimensionId
        ).dimensionColumn}
        <FilterSet
          on:remove-filters={() => clearFilterForDimension(id)}
          on:select={(event) => onSelectItem(event, { dimensionId })}
          {name}
          id={dimensionId}
          {selectedValues}
        />
      {/each}
    {/if}

    {#if hasFilters}
      <button
        transition:fly|local={{ duration: 200, y: 5 }}
        on:click={clearAllFilters}
        class="
            grid gap-x-2 items-center font-bold
            bg-red-50
            text-red-500
            p-1
            pl-2 pr-2
            rounded
        "
        style:grid-template-columns="max-content max-content"
      >
        <FilterRemove size="18px" />
        clear all filters
      </button>
    {/if}
  </FilterContainer>
</div>
