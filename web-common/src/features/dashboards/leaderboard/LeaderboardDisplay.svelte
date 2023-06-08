<script lang="ts">
  import VirtualizedGrid from "@rilldata/web-common/components/VirtualizedGrid.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { NicelyFormattedTypes } from "../humanize-numbers";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardMeasureSelector from "./LeaderboardMeasureSelector.svelte";

  export let metricViewName: string;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricViewName);

  // query the `/meta` endpoint to get the metric's measures and dimensions
  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  let dimensions: Array<MetricsViewDimension>;
  $: dimensions = $metaQuery.data?.dimensions;
  $: measures = $metaQuery.data?.measures;

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;

  $: activeMeasure =
    measures &&
    measures.find(
      (measure) => measure.name === $dashboardStore?.leaderboardMeasureName
    );

  $: metricTimeSeries = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $metricTimeSeries.data;

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime.instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: $dashboardStore?.selectedTimeRange?.start?.toISOString(),
      timeEnd: $dashboardStore?.selectedTimeRange?.end?.toISOString(),
      filter: $dashboardStore?.filters,
    },
    {
      query: {
        enabled:
          (hasTimeSeries ? !!$dashboardStore?.selectedTimeRange : true) &&
          !!$dashboardStore?.filters,
      },
    }
  );

  $: formatPreset =
    (activeMeasure?.format as NicelyFormattedTypes) ??
    NicelyFormattedTypes.HUMANIZE;

  let referenceValue: number;
  $: if (activeMeasure?.name && $totalsQuery?.data?.data) {
    referenceValue = $totalsQuery.data.data?.[activeMeasure.name];
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
  {#if $dashboardStore}
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
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
