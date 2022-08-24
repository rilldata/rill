<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { flip } from "svelte/animate";

  import { IconButton } from "$lib/components/button";
  import { ChipContainer, RemovableListChip } from "$lib/components/chip";
  import Filter from "$lib/components/icons/Filter.svelte";
  import ShiftKey from "$lib/components/tooltip/ShiftKey.svelte";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import {
    clearSelectedDimensionLeaderboardAndUpdate,
    clearSelectedLeaderboardValuesAndUpdate,
    toggleSelectedLeaderboardValueAndUpdate,
  } from "$lib/redux-store/explore/explore-apis";
  import { store } from "$lib/redux-store/store-root";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  export let metricsDefId;
  export let values;

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

  $: prunedValues = pruneValues(values);

  function onSelectItem(event, item) {
    event.detail.forEach((dimensionValue) => {
      console.log(dimensionValue);
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
  class="pt-3 pb-3 grid gap-x-2"
  style:grid-template-columns="max-content auto"
  style:min-height="44px"
>
  <Tooltip
    location="right"
    alignment="middle"
    distance={8}
    suppress={!hasFilters}
  >
    <IconButton disabled={!hasFilters} on:click={clearAllFilters}>
      <Filter />
    </IconButton>

    <TooltipContent slot="tooltip-content">
      <TooltipShortcutContainer padTop>
        <div>clear all filters</div>
        <Shortcut><ShiftKey /> + Click</Shortcut>
      </TooltipShortcutContainer>
    </TooltipContent>
  </Tooltip>
  {#if prunedValues?.length && $dimensions?.length}
    <ChipContainer>
      {#each prunedValues as [dimensionId, selectedValues] (dimensionId)}
        {@const dimension = $dimensions.find((dim) => dim.id === dimensionId)}
        {@const name = dimension?.labelSingle?.length
          ? dimension?.labelSingle
          : dimension?.dimensionColumn}
        <div animate:flip={{ duration: 200 }}>
          <RemovableListChip
            on:remove={() => clearFilterForDimension(dimensionId)}
            on:apply={(event) => onSelectItem(event, { dimensionId })}
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
    </ChipContainer>
  {:else if prunedValues?.length === 0}
    <div
      in:fly|local={{ duration: 200, x: 8 }}
      class="italic text-gray-400 ml-1 grid items-center"
      style:min-height="26px"
    >
      no filters selected
    </div>
  {:else}
    &nbsp;
  {/if}
</section>
