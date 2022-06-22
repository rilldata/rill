<script lang="ts">
  import { createEventDispatcher, getContext } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import { browser } from "$app/env";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import type {
    ActiveValues,
    MetricsLeaderboardEntity,
  } from "$lib/redux-store/metrics-leaderboard-slice";
  import {
    setBigNumber,
    setReferenceValue,
    toggleLeaderboardActiveValue,
  } from "$lib/redux-store/metrics-leaderboard-slice";
  import { MetricsExploreClient } from "$lib/components/leaderboard/MetricsExploreClient";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import { updateDisplay } from "./utils";

  export let metricsDefId: string;
  export let columns: number;
  export let referenceValue: number;

  let metricsLeaderboard: MetricsLeaderboardEntity;
  $: if (metricsDefId && $reduxReadable)
    metricsLeaderboard =
      $reduxReadable.metricsLeaderboard.entities[metricsDefId];

  const dispatch = createEventDispatcher();
  let leaderboardExpanded;
  let anythingSelected: boolean;
  $: anythingSelected = isAnythingSelected(metricsLeaderboard?.activeValues);

  function onSelectItem(event, item) {
    dispatch("select-item", {
      fieldName: event.detail,
      dimensionName: item.displayName,
    });

    store.dispatch(
      toggleLeaderboardActiveValue(metricsDefId, item.displayName, event.detail)
    );

    if (browser && metricsLeaderboard.measureId) {
      updateDisplay(metricsDefId, metricsLeaderboard, anythingSelected);
    }
  }
</script>

<div
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  class="border-t border-gray-200 overflow-auto"
>
  {#if metricsLeaderboard}
    <VirtualizedGrid
      {columns}
      height="100%"
      items={metricsLeaderboard.leaderboards}
      let:item
    >
      <!-- the single virtual element -->
      <div style:width="315px">
        <Leaderboard
          seeMore={leaderboardExpanded === item.displayName}
          on:expand={() => {
            if (leaderboardExpanded === item.displayName) {
              leaderboardExpanded = undefined;
            } else {
              leaderboardExpanded = item.displayName;
            }
          }}
          on:select-item={(event) => onSelectItem(event, item)}
          activeValues={metricsLeaderboard.activeValues[item.displayName] || []}
          displayName={item.displayName}
          values={item.values}
          referenceValue={referenceValue || 0}
        />
      </div>
    </VirtualizedGrid>
  {/if}
</div>
