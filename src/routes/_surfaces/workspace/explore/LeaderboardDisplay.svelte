<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { toggleSelectedLeaderboardValueAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import type { Readable } from "svelte/store";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import LeaderboardMeasureSelector from "$lib/components/leaderboard/LeaderboardMeasureSelector.svelte";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import { getMeasureFieldNameByIdAndIndex } from "$lib/redux-store/measure-definition/measure-definition-readables";

  export let metricsDefId: string;
  export let columns: number;
  export let whichReferenceValue: string;

  let metricsLeaderboard: Readable<MetricsExploreEntity>;
  $: metricsLeaderboard = getMetricsExploreById(metricsDefId);

  let measureField: Readable<string>;
  $: if ($metricsLeaderboard?.leaderboardMeasureId)
    measureField = getMeasureFieldNameByIdAndIndex(
      $metricsLeaderboard.leaderboardMeasureId,
      $metricsLeaderboard.measureIds.indexOf(
        $metricsLeaderboard?.leaderboardMeasureId
      )
    );

  let bigNumberEntity: Readable<BigNumberEntity>;
  $: bigNumberEntity = getBigNumberById(metricsDefId);
  let referenceValue: number;
  $: if ($bigNumberEntity && $measureField) {
    referenceValue =
      whichReferenceValue === "filtered"
        ? $bigNumberEntity.bigNumbers?.[$measureField]
        : $bigNumberEntity.referenceValues?.[$measureField];
  }

  const dispatch = createEventDispatcher();
  let leaderboardExpanded;

  function onSelectItem(event, item) {
    dispatch("select-item", {
      fieldName: event.detail,
      dimensionName: item.displayName,
    });

    toggleSelectedLeaderboardValueAndUpdate(
      store.dispatch,
      metricsDefId,
      item.displayName,
      event.detail,
      true
    );
  }
</script>

<!-- container for the metrics leaderboard components and controls -->
<div
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  class="border-t border-gray-200 overflow-auto"
>
  <LeaderboardMeasureSelector {metricsDefId} />
  {#if $metricsLeaderboard}
    <VirtualizedGrid
      {columns}
      height="100%"
      items={$metricsLeaderboard.leaderboards ?? []}
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
          activeValues={$metricsLeaderboard.activeValues[item.displayName] ??
            []}
          displayName={item.displayName}
          values={item.values}
          referenceValue={referenceValue || 0}
        />
      </div>
    </VirtualizedGrid>
  {/if}
</div>
