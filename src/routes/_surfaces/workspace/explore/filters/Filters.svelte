<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { flip } from "svelte/animate";

  import { Chip, ChipContainer, RemovableListChip } from "$lib/components/chip";
  import Filter from "$lib/components/icons/Filter.svelte";
  import FilterRemove from "$lib/components/icons/FilterRemove.svelte";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import {
    clearSelectedDimensionLeaderboardAndUpdate,
    clearSelectedLeaderboardValuesAndUpdate,
    toggleSelectedLeaderboardValueAndUpdate,
  } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { store } from "$lib/redux-store/store-root";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import { getDisplayName } from "../utils";
  export let metricsDefId;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  $: values = $metricsExplorer?.activeValues;

  let dimensions: Readable<DimensionDefinitionEntity[]>;
  $: dimensions = getDimensionsByMetricsId(metricsDefId);

  function clearFilterForDimension(dimension) {
    clearSelectedDimensionLeaderboardAndUpdate(
      store.dispatch,
      metricsDefId,
      dimension
    );
  }

  $: hasFilters = isAnythingSelected(values);

  function clearAllFilters() {
    if (hasFilters)
      clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }

  function pruneValues(set) {
    if (!set) return;
    return Object.keys(set)
      .filter((key) => set[key].length)
      .map((key) => {
        return [key, set[key].filter(([_, v]) => v === true).map(([k]) => k)];
      });
  }

  /** prune the values and prepare for for templating */
  let currentDimensionFilters = [];
  $: if (values)
    currentDimensionFilters = pruneValues(values).map(
      ([dimensionId, selectedValues]) => {
        const dimension = $dimensions.find((dim) => dim.id === dimensionId);
        return {
          name: getDisplayName(dimension),
          dimensionId,
          selectedValues,
        };
      }
    );

  function toggleDimensionValue(event, item) {
    event.detail.forEach((dimensionValue) => {
      toggleSelectedLeaderboardValueAndUpdate(
        store.dispatch,
        metricsDefId,
        item.dimensionId,
        dimensionValue,
        true
      );
    });
  }
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
    class:text-gray-800={hasFilters}
  >
    <Filter size="16px" />
  </div>
  {#if currentDimensionFilters?.length}
    <ChipContainer>
      {#each currentDimensionFilters as { name, dimensionId, selectedValues } (dimensionId)}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:remove={() => clearFilterForDimension(dimensionId)}
            on:apply={(event) => toggleDimensionValue(event, { dimensionId })}
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
      <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
      {#if hasFilters}
        <div class="ml-auto">
          <Chip
            bgBaseColor="bg-white"
            bgHoverColor="bg-gray-100"
            textColor="text-gray-500"
            bgActiveColor="bg-gray-200"
            ringOffsetColor="ring-offset-gray-400"
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
  {:else if currentDimensionFilters?.length === 0}
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
