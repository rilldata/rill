<script lang="ts">
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { getMapFromArray } from "$common/utils/arrayUtils";
  import {
    LeaderboardValue,
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import LeaderboardMeasureSelector from "$lib/components/leaderboard/LeaderboardMeasureSelector.svelte";
  import VirtualizedGrid from "$lib/components/VirtualizedGrid.svelte";
  import {
    useMetaQuery,
    useTotalsQuery,
  } from "$lib/svelte-query/queries/metrics-view";
  import {
    getScaleForLeaderboard,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "$lib/util/humanize-numbers";
  import { onDestroy, onMount } from "svelte";
  import Leaderboard from "./Leaderboard.svelte";

  export let metricsDefId: string;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  // query the `/meta` endpoint to get the metric's measures and dimensions
  $: metaQuery = useMetaQuery(metricsDefId);
  $: dimensions = $metaQuery.data?.dimensions;
  $: measures = $metaQuery.data?.measures;

  $: activeMeasure =
    measures &&
    measures.find(
      (measure) => measure.id === metricsExplorer?.leaderboardMeasureId
    );

  $: formatPreset =
    activeMeasure?.formatPreset ?? NicelyFormattedTypes.HUMANIZE;

  let totalsQuery;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    totalsQuery = useTotalsQuery(metricsDefId, {
      measures: metricsExplorer?.selectedMeasureIds,
      time: {
        start: metricsExplorer?.selectedTimeRange?.start,
        end: metricsExplorer?.selectedTimeRange?.end,
      },
    });
  }

  let referenceValue: number;
  $: if (activeMeasure?.sqlName && $totalsQuery?.data?.data) {
    referenceValue = $totalsQuery.data.data?.[activeMeasure.sqlName];
  }

  const leaderboards = new Map<string, Array<LeaderboardValue>>();
  $: if (dimensions) {
    const dimensionIdMap = getMapFromArray(
      dimensions,
      (dimension) => dimension.id
    );
    [...leaderboards.keys()]
      .filter((dimensionId) => !dimensionIdMap.has(dimensionId))
      .forEach((dimensionId) => leaderboards.delete(dimensionId));
  }

  /** create a scale for the valid leaderboards */
  let leaderboardFormatScale: ShortHandSymbols = "none";

  let leaderboardExpanded;

  function onSelectItem(event, item: DimensionDefinitionEntity) {
    metricsExplorerStore.toggleFilter(
      metricsDefId,
      item.id,
      event.detail.label
    );
  }

  function onLeaderboardValues(event) {
    leaderboards.set(event.detail.dimensionId, event.detail.values);
    if (
      formatPreset === NicelyFormattedTypes.HUMANIZE ||
      formatPreset === NicelyFormattedTypes.CURRENCY
    ) {
      leaderboardFormatScale = getScaleForLeaderboard(leaderboards);
    }
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
        on:expand={() => {
          if (leaderboardExpanded === item.id) {
            leaderboardExpanded = undefined;
          } else {
            leaderboardExpanded = item.id;
          }
        }}
        on:select-item={(event) => onSelectItem(event, item)}
        on:leaderboard-value={onLeaderboardValues}
        referenceValue={referenceValue || 0}
      />
    </VirtualizedGrid>
  {/if}
</div>
