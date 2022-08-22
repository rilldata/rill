<script lang="ts">
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import type {
    MetricViewMetaResponse,
    MetricViewTotalsRequest,
    MetricViewTotalsResponse,
  } from "$common/rill-developer-service/MetricViewActions";
  import LeaderboardMeasureSelector from "$lib/components/leaderboard/LeaderboardMeasureSelector.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { toggleLeaderboardActiveValue } from "$lib/redux-store/explore/explore-slice";
  import { store } from "$lib/redux-store/store-root";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
    getMetricViewTotals,
    getMetricViewTotalsQueryKey,
    invalidateMetricViewData,
  } from "$lib/svelte-query/queries/metric-view";
  import {
    getScaleForLeaderboard,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "$lib/util/humanize-numbers";
  import { useQuery, useQueryClient } from "@sveltestack/svelte-query";
  import { onDestroy, onMount } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import { MetricsExplorerStore } from "$lib/application-state-stores/explorer-stores";

  export let metricsDefId: string;
  export let whichReferenceValue: string;

  const queryClient = useQueryClient();

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $MetricsExplorerStore.entities[metricsDefId];

  // query the `/meta` endpoint to get the metric's measures and dimensions
  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }

  let dimensions: DimensionDefinitionEntity[];
  $: dimensions = $queryResult.data.dimensions;

  let measures: MeasureDefinitionEntity[];
  $: measures = $queryResult.data.measures;

  $: activeMeasure =
    measures &&
    measures.find(
      (measure) => measure.id === metricsExplorer?.leaderboardMeasureId
    );

  $: formatPreset =
    activeMeasure?.formatPreset ?? NicelyFormattedTypes.HUMANIZE;

  let referenceValue: number;

  function getTotalsRequest(noFilters = false): MetricViewTotalsRequest {
    return {
      measures: [metricsExplorer.leaderboardMeasureId],
      filter: noFilters ? undefined : metricsExplorer.filters,
      time: {
        start: metricsExplorer.selectedTimeRange?.start,
        end: metricsExplorer.selectedTimeRange?.end,
      },
    };
  }
  let totalsQueryKey = getMetricViewTotalsQueryKey(metricsDefId);
  const totalsQuery = useQuery<MetricViewTotalsResponse>(totalsQueryKey, () =>
    getMetricViewTotals(metricsDefId, getTotalsRequest())
  );
  // TODO: find a way to have a single request when there are no filters
  let referenceValueKey = getMetricViewTotalsQueryKey(metricsDefId, true);
  const referenceValueQuery = useQuery<MetricViewTotalsResponse>(
    referenceValueKey,
    () => getMetricViewTotals(metricsDefId, getTotalsRequest(true))
  );
  $: {
    totalsQueryKey = getMetricViewTotalsQueryKey(metricsDefId);
    totalsQuery.setOptions(totalsQueryKey, () =>
      getMetricViewTotals(metricsDefId, getTotalsRequest())
    );
    referenceValueKey = getMetricViewTotalsQueryKey(metricsDefId, true);
    referenceValueQuery.setOptions(referenceValueKey, () =>
      getMetricViewTotals(metricsDefId, getTotalsRequest(true))
    );
  }

  $: if ($totalsQuery && $referenceValueQuery && activeMeasure?.sqlName) {
    referenceValue =
      whichReferenceValue === "filtered"
        ? $totalsQuery.data.data?.[activeMeasure.sqlName]
        : $referenceValueQuery.data.data?.[activeMeasure.sqlName];
  }

  /** Filter out the leaderboards whose underlying dimensions do not pass the validation step. */
  // Q: We're doing this on the backend now, right? We can delete this?
  $: validLeaderboards =
    dimensions && metricsExplorer?.leaderboards
      ? metricsExplorer?.leaderboards.filter((leaderboard) => {
          const dimensionConfiguration = dimensions?.find(
            (dimension) => dimension.id === leaderboard.dimensionId
          );
          return dimensionConfiguration;
        })
      : [];

  /** create a scale for the valid leaderboards */
  let leaderboardFormatScale: ShortHandSymbols = "none";
  $: if (
    validLeaderboards &&
    (formatPreset === NicelyFormattedTypes.HUMANIZE ||
      formatPreset === NicelyFormattedTypes.CURRENCY)
  ) {
    leaderboardFormatScale = getScaleForLeaderboard(validLeaderboards);
  }

  let leaderboardExpanded;

  function onSelectItem(event, item: DimensionDefinitionEntity) {
    store.dispatch(
      toggleLeaderboardActiveValue(metricsDefId, item.id, event.detail.label)
    );
    invalidateMetricViewData(queryClient, metricsDefId);
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
  <div
    class="grid grid-auto-cols justify-start grid-flow-col items-end p-1 pb-3"
  >
    <LeaderboardMeasureSelector {metricsDefId} />
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
        {metricsDefId}
        dimensionId={item.id}
        seeMore={leaderboardExpanded === item.id}
        on:expand={() => {
          if (leaderboardExpanded === item.id) {
            leaderboardExpanded = undefined;
          } else {
            leaderboardExpanded = item.id;
          }
        }}
        on:select-item={(event) => onSelectItem(event, item)}
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
