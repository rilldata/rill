<script lang="ts">
  import LeaderboardMeasureSelector from "$lib/components/leaderboard/LeaderboardMeasureSelector.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import { toggleSelectedLeaderboardValueAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type {
    LeaderboardValues,
    MetricsExplorerEntity,
  } from "$lib/redux-store/explore/explore-slice";
  import {
    getMeasureFieldNameByIdAndIndex,
    getMeasuresByMetricsId,
  } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import {
    getScaleForLeaderboard,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "$lib/util/humanize-numbers";
  import { onDestroy, onMount } from "svelte";
  import type { Readable } from "svelte/store";
  import Leaderboard from "./Leaderboard.svelte";

  export let metricsDefId: string;
  export let whichReferenceValue: string;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let measureField: Readable<string>;
  $: if ($metricsExplorer?.leaderboardMeasureId)
    measureField = getMeasureFieldNameByIdAndIndex(
      $metricsExplorer.leaderboardMeasureId,
      $metricsExplorer.measureIds.indexOf(
        $metricsExplorer?.leaderboardMeasureId
      )
    );

  $: measures = getMeasuresByMetricsId(metricsDefId);
  $: leaderboardMeasureDefinition = $measures.find(
    (measure) => measure.id === $metricsExplorer?.leaderboardMeasureId
  );
  // get the expression so we can determine if the measure is summable
  $: expression = leaderboardMeasureDefinition?.expression;
  $: formatPreset =
    leaderboardMeasureDefinition?.formatPreset ?? NicelyFormattedTypes.HUMANIZE;

  let bigNumberEntity: Readable<BigNumberEntity>;
  $: bigNumberEntity = getBigNumberById(metricsDefId);
  let referenceValue: number;

  $: if ($bigNumberEntity && $measureField) {
    referenceValue =
      whichReferenceValue === "filtered"
        ? $bigNumberEntity.bigNumbers?.[$measureField]
        : $bigNumberEntity.referenceValues?.[$measureField];
  }

  let leaderboardFormatScale: ShortHandSymbols = "none";
  $: if (
    $metricsExplorer &&
    (formatPreset === NicelyFormattedTypes.HUMANIZE ||
      formatPreset === NicelyFormattedTypes.CURRENCY)
  ) {
    leaderboardFormatScale = getScaleForLeaderboard(
      $metricsExplorer.leaderboards
    );
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
    if (!leaderboardContainer) return;
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.max(1, Math.floor(availableWidth / (315 + 20)));
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
  style:min-width="365px"
  bind:this={leaderboardContainer}
>
  <div class="grid grid-auto-cols justify-end grid-flow-col items-end p-1 pb-3">
    <LeaderboardMeasureSelector {metricsDefId} />
  </div>
  {#if $metricsExplorer}
    <VirtualizedGrid
      {columns}
      height="100%"
      items={$metricsExplorer.leaderboards ?? []}
      let:item
    >
      <!-- the single virtual element -->
      <Leaderboard
        {formatPreset}
        {leaderboardFormatScale}
        isSummableMeasure={expression?.toLowerCase()?.includes("count(") ||
          expression?.toLowerCase()?.includes("sum(")}
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
        activeValues={$metricsExplorer.activeValues[item.dimensionId] ?? []}
        values={item.values}
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
