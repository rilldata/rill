<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type {
    LeaderboardValues,
    MetricsExploreEntity,
  } from "$lib/redux-store/explore/explore-slice";
  import { toggleSelectedLeaderboardValueAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import type { Readable } from "svelte/store";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import LeaderboardMeasureSelector from "$lib/components/leaderboard/LeaderboardMeasureSelector.svelte";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import { getMeasureFieldNameByIdAndIndex } from "$lib/redux-store/measure-definition/measure-definition-readables";

  export let metricsDefId: string;
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

  let leaderboardExpanded;

  function onSelectItem(event, item: LeaderboardValues) {
    toggleSelectedLeaderboardValueAndUpdate(
      store.dispatch,
      metricsDefId,
      item.dimensionId,
      event.detail.label,
      true
    );
  }

  /** Functionality for resizing the virtual leaderboard */
  let columns = 3;
  let availableWidth = 0;
  let leaderboardContainer: HTMLElement;
  let observer: ResizeObserver;

  function onResize() {
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.floor(availableWidth / (315 + 20));
  }

  onMount(() => {
    onResize();
    const observer = new ResizeObserver(() => {
      onResize();
    });
    observer.observe(leaderboardContainer);
  });

  onDestroy(() => {
    observer?.disconnect();
  });
</script>

<svelte:window on:resize={onResize} />
<!-- container for the metrics leaderboard components and controls -->
<div
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  bind:this={leaderboardContainer}
>
  <div class="grid grid-auto-cols justify-end grid-flow-col items-end p-1 pb-3">
    <LeaderboardMeasureSelector {metricsDefId} />
  </div>
  {#if $metricsLeaderboard}
    <VirtualizedGrid
      {columns}
      height="100%"
      items={$metricsLeaderboard.leaderboards ?? []}
      let:item
    >
      <!-- the single virtual element -->
      <Leaderboard
        dimensionId={item.dimensionId}
        seeMore={leaderboardExpanded === item.dimensionId}
        on:expand={() => {
          if (leaderboardExpanded === item.dimensionId) {
            leaderboardExpanded = undefined;
          } else {
            leaderboardExpanded = item.dimensionId;
          }
        }}
        on:select-item={(event) => onSelectItem(event, item)}
        activeValues={$metricsLeaderboard.activeValues[item.dimensionId] ?? []}
        values={item.values}
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
