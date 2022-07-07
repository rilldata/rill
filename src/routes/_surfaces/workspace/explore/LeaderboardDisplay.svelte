<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import { toggleValueAndUpdateLeaderboard } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-apis";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import type { Readable } from "svelte/store";
  import { getMetricsLeaderboardById } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-readables";

  export let metricsDefId: string;
  export let referenceValue: number;

  let metricsLeaderboard: Readable<MetricsLeaderboardEntity>;
  $: metricsLeaderboard = getMetricsLeaderboardById(metricsDefId);

  const dispatch = createEventDispatcher();
  let leaderboardExpanded;

  let measure: Readable<MeasureDefinitionEntity>;
  $: if ($metricsLeaderboard?.measureId) {
    measure = getMeasureById($metricsLeaderboard?.measureId);
  }

  function onSelectItem(event, item) {
    dispatch("select-item", {
      fieldName: event.detail.label,
      dimensionName: item.displayName,
    });

    toggleValueAndUpdateLeaderboard(
      store.dispatch,
      metricsDefId,
      item.displayName,
      event.detail.label,
      !event.detail.isActive,
      $measure.expression
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
    observer.disconnect();
  });
</script>

<svelte:window on:resize={onResize} />
<!-- container for the metrics leaderboard components and controls -->
<div
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  bind:this={leaderboardContainer}
>
  {#if $metricsLeaderboard}
    <VirtualizedGrid
      {columns}
      height="100%"
      items={$metricsLeaderboard.leaderboards ?? []}
      let:item
    >
      <!-- the single virtual element -->
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
        activeValues={$metricsLeaderboard.activeValues[item.displayName] ?? []}
        displayName={item.displayName}
        values={item.values}
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
