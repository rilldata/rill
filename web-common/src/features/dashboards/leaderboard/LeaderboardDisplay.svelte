<script lang="ts">
  import VirtualizedGrid from "@rilldata/web-common/components/VirtualizedGrid.svelte";
  import {
    ROW_COUNT_INLINE_COL_EXPRESSION,
    ROW_COUNT_INLINE_COL_NAME,
    cancelDashboardQueries,
  } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MetricsViewDimension,
    V1MetricsViewTotalsResponse,
    createQueryServiceMetricsViewTotals,
    QueryServiceMetricsViewTotalsBody,
    V1InlineMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { CreateQueryResult, useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import {
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

  $: selectedMeasureNames = [
    ROW_COUNT_INLINE_COL_NAME,
    ...(metricsExplorer?.selectedMeasureNames ?? []),
  ];

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
  let totalsQueryFiltered: CreateQueryResult<
    V1MetricsViewTotalsResponse,
    Error
  >;

  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    let inlineMeasures: V1InlineMeasure[] = [
      {
        name: ROW_COUNT_INLINE_COL_NAME,
        expression: ROW_COUNT_INLINE_COL_EXPRESSION,
      },
    ];
    let totalsQueryParams: QueryServiceMetricsViewTotalsBody = {
      measureNames: selectedMeasureNames,
      inlineMeasures,
    };
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

    let totalsQueryFilteredParams = {
      ...totalsQueryParams,
      filter: metricsExplorer?.filters,
    };

    totalsQueryFiltered = createQueryServiceMetricsViewTotals(
      $runtime.instanceId,
      metricViewName,
      totalsQueryFilteredParams
    );
  }

  let referenceValue: number;
  $: if (activeMeasure?.name && $totalsQuery?.data?.data) {
    referenceValue = $totalsQuery.data.data?.[activeMeasure.name];
  }

  $: totalFilteredRowCount =
    $totalsQueryFiltered?.data?.data?.[ROW_COUNT_INLINE_COL_NAME] ?? 0;

  let leaderboardExpanded;

  function onSelectItem(event, item: MetricsViewDimension) {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(
      metricViewName,
      item.name,
      event.detail.label
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

  // FIXME: this is pending the remaining state work for show/hide measures and dimensions
  // $: availableDimensionKeys = selectDimensionKeys($metaQuery);
  // $: visibleDimensionKeys = metricsExplorer?.visibleDimensionKeys;
  // $: visibleDimensionsBitmask = availableDimensionKeys.map((k) =>
  //   visibleDimensionKeys.has(k)
  // );

  // $: dimensionsShown =
  //   dimensions?.filter((_, i) => visibleDimensionsBitmask[i]) ?? [];

  $: dimensionsShown = dimensions ?? [];
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
        {totalFilteredRowCount}
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
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
