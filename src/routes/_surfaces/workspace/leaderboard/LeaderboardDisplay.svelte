<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard-slice";
  import {
    singleMetricsLeaderboardSelector,
    toggleValueAndUpdateLeaderboard,
  } from "$lib/redux-store/metrics-leaderboard-slice";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";

  export let metricsDefId: string;
  export let columns: number;
  export let referenceValue: number;

  let metricsLeaderboard: MetricsLeaderboardEntity;
  $: metricsLeaderboard =
    singleMetricsLeaderboardSelector(metricsDefId)($reduxReadable);

  const dispatch = createEventDispatcher();
  let leaderboardExpanded;
  let anythingSelected: boolean;
  $: anythingSelected = isAnythingSelected(metricsLeaderboard?.activeValues);

  function onSelectItem(event, item) {
    dispatch("select-item", {
      fieldName: event.detail,
      dimensionName: item.displayName,
    });

    toggleValueAndUpdateLeaderboard(
      store.dispatch,
      metricsDefId,
      item.displayName,
      event.detail
    );
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
      items={metricsLeaderboard.leaderboards ?? []}
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
