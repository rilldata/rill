<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import { toggleValueAndUpdateLeaderboard } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-apis";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import MetricsExploreTimeChart from "$lib/components/leaderboard/MetricsExploreTimeChart.svelte";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import type { Readable } from "svelte/store";
  import { getMetricsLeaderboardById } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-readables";

  export let metricsDefId: string;
  export let columns: number;
  export let referenceValue: number;

  let metricsLeaderboard: Readable<MetricsLeaderboardEntity>;
  $: metricsLeaderboard = getMetricsLeaderboardById(metricsDefId);

  const dispatch = createEventDispatcher();
  let leaderboardExpanded;
  let anythingSelected: boolean;
  $: anythingSelected = isAnythingSelected($metricsLeaderboard?.activeValues);

  let measure: Readable<MeasureDefinitionEntity>;
  $: if ($metricsLeaderboard?.measureId) {
    measure = getMeasureById($metricsLeaderboard?.measureId);
  }

  function onSelectItem(event, item) {
    dispatch("select-item", {
      fieldName: event.detail,
      dimensionName: item.displayName,
    });

    toggleValueAndUpdateLeaderboard(
      store.dispatch,
      metricsDefId,
      item.displayName,
      event.detail,
      true,
      $measure.expression
    );
  }
</script>

<div
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  class="border-t border-gray-200 overflow-auto"
>
  {#if $metricsLeaderboard}
    <MetricsExploreTimeChart {metricsDefId} />
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
