<script lang="ts">
  import VirtualizedGrid from "@rilldata/web-common/components/VirtualizedGrid.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    selectDimensionKeys,
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import {
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
    V1MetricsViewTotalsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { CreateQueryResult, useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
    LeaderboardValue,
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import { NicelyFormattedTypes } from "../humanize-numbers";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardMeasureSelector from "./LeaderboardMeasureSelector.svelte";

  export let metricViewName: string;

  const queryClient = useQueryClient();

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  // query the `/meta` endpoint to get the metric's measures and dimensions
  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  let dimensions: Array<MetricsViewDimension>;
  $: dimensions = $metaQuery.data?.dimensions;
  $: measures = $metaQuery.data?.measures;

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  $: activeMeasure =
    measures &&
    measures.find(
      (measure) => measure.name === metricsExplorer?.leaderboardMeasureName
    );

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: formatPreset =
    (activeMeasure?.format as NicelyFormattedTypes) ??
    NicelyFormattedTypes.HUMANIZE;

  let totalsQuery: CreateQueryResult<V1MetricsViewTotalsResponse, Error>;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    let totalsQueryParams = { measureNames: selectedMeasureNames };
    if (hasTimeSeries) {
      totalsQueryParams = {
        ...totalsQueryParams,
        ...{
          timeStart: metricsExplorer.selectedTimeRange?.start,
          timeEnd: metricsExplorer.selectedTimeRange?.end,
        },
      };
    }
    totalsQuery = createQueryServiceMetricsViewTotals(
      $runtime.instanceId,
      metricViewName,
      totalsQueryParams
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
    cancelDashboardQueries(queryClient, metricViewName);
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

  $: availableDimensionKeys = selectDimensionKeys($metaQuery);
  $: visibleDimensionKeys = metricsExplorer?.visibleDimensionKeys;
  $: visibleDimensionsBitmask = availableDimensionKeys.map((k) =>
    visibleDimensionKeys.has(k)
  );

  $: dimensionsShown =
    dimensions?.filter((_, i) => visibleDimensionsBitmask[i]) ?? [];
</script>

<svelte:window on:resize={onResize} />
<!-- container for the metrics leaderboard components and controls -->
<div
  bind:this={leaderboardContainer}
  style:height="calc(100vh - 130px - 4rem)"
  style:min-width="365px"
>
  <div
    class="grid grid-auto-cols justify-between grid-flow-col items-center pl-1 pb-3"
  >
    <LeaderboardMeasureSelector {metricViewName} />
  </div>
  {#if metricsExplorer}
    <VirtualizedGrid {columns} height="100%" items={dimensionsShown} let:item>
      <!-- the single virtual element -->
      <Leaderboard
        {formatPreset}
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
