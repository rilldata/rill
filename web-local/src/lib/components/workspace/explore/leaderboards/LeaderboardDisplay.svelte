<script lang="ts">
  import {
    MetricsViewDimension,
    useRuntimeServiceMetricsViewTotals,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useMetaQuery } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import { getMapFromArray } from "@rilldata/web-local/lib/util/arrayUtils";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import {
    LeaderboardValue,
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import {
    determineScaleForValues,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "../../../../util/humanize-numbers";
  import LeaderboardMeasureSelector from "../../../leaderboard/LeaderboardMeasureSelector.svelte";
  import VirtualizedGrid from "../../../VirtualizedGrid.svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import {
    getAllLeaderboardValues,
    getLeaderboardStore,
  } from "./leaderboardStore";

  export let metricViewName: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  // query the `/meta` endpoint to get the metric's measures and dimensions
  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  let dimensions: Array<MetricsViewDimension>;
  $: dimensions = $metaQuery.data?.dimensions;
  $: measures = $metaQuery.data?.measures;

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  $: activeMeasure =
    measures &&
    measures.find(
      (measure) => measure.name === metricsExplorer?.leaderboardMeasureName
    );

  $: leaderboardStore = getLeaderboardStore(
    $runtimeStore.instanceId,
    metricViewName,
    activeMeasure?.name,
    dimensions
  );

  $: allLeaderboardValues = getAllLeaderboardValues(leaderboardStore);

  $: formatPreset =
    (activeMeasure?.format as NicelyFormattedTypes) ??
    NicelyFormattedTypes.HUMANIZE;

  /** create a scale for the valid leaderboards */
  let leaderboardFormatScale: ShortHandSymbols = "none";

  $: if (
    $allLeaderboardValues &&
    (formatPreset === NicelyFormattedTypes.HUMANIZE ||
      formatPreset === NicelyFormattedTypes.CURRENCY)
  ) {
    leaderboardFormatScale = determineScaleForValues($allLeaderboardValues);
  }

  let totalsQuery: UseQueryStoreResult<V1MetricsViewTotalsResponse, Error>;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    totalsQuery = useRuntimeServiceMetricsViewTotals(
      $runtimeStore.instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        timeStart: metricsExplorer.selectedTimeRange?.start,
        timeEnd: metricsExplorer.selectedTimeRange?.end,
      }
    );
  }

  let referenceValue: number;
  $: if (activeMeasure?.name && $totalsQuery?.data?.data) {
    referenceValue = $totalsQuery.data.data?.[activeMeasure.name];
  }

  const leaderboards = new Map<string, Array<LeaderboardValue>>();
  $: if (dimensions) {
    const dimensionNameMap = getMapFromArray(
      dimensions,
      (dimension) => dimension.name
    );
    [...leaderboards.keys()]
      .filter((dimensionName) => !dimensionNameMap.has(dimensionName))
      .forEach((dimensionName) => leaderboards.delete(dimensionName));
  }

  let leaderboardExpanded;

  function onSelectItem(event, item: MetricsViewDimension) {
    metricsExplorerStore.toggleFilter(
      metricViewName,
      item.name,
      event.detail.label
    );
  }

  function onLeaderboardValues(event) {
    leaderboards.set(event.detail.dimensionName, event.detail.values);
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
  bind:this={leaderboardContainer}
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  style:min-width="365px"
>
  <div
    class="grid grid-auto-cols justify-start grid-flow-col items-end p-1 pb-3"
  >
    <LeaderboardMeasureSelector {metricViewName} />
  </div>
  {#if metricsExplorer}
    <VirtualizedGrid {columns} height="100%" items={dimensions ?? []} let:item>
      <!-- the single virtual element -->
      <Leaderboard
        {formatPreset}
        {leaderboardFormatScale}
        isSummableMeasure={activeMeasure?.expression
          .toLowerCase()
          ?.includes("count(") ||
          activeMeasure?.expression?.toLowerCase()?.includes("sum(")}
        {metricViewName}
        dimensionName={item.name}
        on:expand={() => {
          if (leaderboardExpanded === item.name) {
            leaderboardExpanded = undefined;
          } else {
            leaderboardExpanded = item.name;
          }
        }}
        on:select-item={(event) => onSelectItem(event, item)}
        on:leaderboard-value={onLeaderboardValues}
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
