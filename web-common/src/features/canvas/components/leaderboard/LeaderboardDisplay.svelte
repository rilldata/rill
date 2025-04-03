<script lang="ts">
  import type { LeaderboardSpec } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import {
    COMPARISON_COLUMN_WIDTH,
    valueColumn,
  } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-widths";
  import Leaderboard from "@rilldata/web-common/features/dashboards/leaderboard/Leaderboard.svelte";
  import { SortType } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { selectedDimensionValuesV2 } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Readable } from "svelte/store";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let canvasName: string;

  const DIMENSION_COLUMN_WIDTH = 164;

  let metricsViewName: string;
  let leaderboardMeasureNames: string[] = [];
  let dimensionNames: string[] = [];
  let numRows = 7;

  let parentElement: HTMLDivElement;
  let suppressTooltip = false;
  let showPercentOfTotal = false; // TODO: update this

  $: store = getCanvasStore(canvasName);
  $: ({
    canvasEntity: {
      spec: { getDimensionsForMetricView, getMeasuresForMetricView },
    },
  } = store);

  $: ({ instanceId } = $runtime);

  $: leaderboardProperties = rendererProperties as LeaderboardSpec;

  $: {
    metricsViewName = leaderboardProperties.metrics_view;
    leaderboardMeasureNames = leaderboardProperties.measures ?? [];
    dimensionNames = leaderboardProperties.dimensions ?? [];
    numRows = leaderboardProperties.num_rows ?? 7;
  }

  $: ({ dimensionFilters: whereFilter, dimensionThresholdFilters } =
    splitWhereFilter($timeAndFilterStore.where));

  $: allDimensions = getDimensionsForMetricView(metricsViewName);
  $: allMeasures = getMeasuresForMetricView(metricsViewName);

  $: visibleDimensions = $allDimensions.filter((d) =>
    dimensionNames.includes(d.name || (d.column as string)),
  );

  $: visibleMeasures = $allMeasures.filter((m) =>
    leaderboardMeasureNames.includes(m.name as string),
  );
  $: activeMeasureName = leaderboardMeasureNames?.[0] || "measure";

  $: measureFormatters = Object.fromEntries(
    visibleMeasures.map((m) => [
      m.name,
      createMeasureValueFormatter<null | undefined>(m),
    ]),
  );

  $: showDeltaPercent = !!$timeAndFilterStore.comparisonTimeRange;

  // Reset column widths when the measure changes
  $: if (leaderboardMeasureNames) {
    valueColumn.reset();
  }

  $: tableWidth =
    DIMENSION_COLUMN_WIDTH +
    $valueColumn +
    ($timeAndFilterStore.comparisonTimeRange
      ? COMPARISON_COLUMN_WIDTH * (showDeltaPercent ? 2 : 1)
      : showPercentOfTotal
        ? COMPARISON_COLUMN_WIDTH
        : 0);

  function getSelectedValues(dimensionName: string) {
    return selectedDimensionValuesV2(
      $runtime.instanceId,
      [metricsViewName],
      whereFilter,
      dimensionName,
      $timeAndFilterStore.timeRange.start,
      $timeAndFilterStore.timeRange.end,
    );
  }

  function isValidPercentOfTotal(measureName: string) {
    return (
      visibleMeasures.find((m) => m.name === measureName)
        ?.validPercentOfTotal ?? false
    );
  }

  $: console.log($timeAndFilterStore.comparisonTimeRange);
</script>

<div class="flex flex-col overflow-hidden size-full" aria-label="Leaderboards">
  <div
    bind:this={parentElement}
    class="overflow-y-auto leaderboard-display"
    on:scroll={() => {
      suppressTooltip = true;
    }}
    on:scrollend={() => {
      suppressTooltip = false;
    }}
  >
    {#if parentElement}
      <div class="leaderboard-grid overflow-hidden pb-4">
        {#each visibleDimensions as dimension (dimension.name)}
          {#if dimension.name}
            <Leaderboard
              slice={numRows}
              {instanceId}
              {isValidPercentOfTotal}
              {metricsViewName}
              {activeMeasureName}
              {leaderboardMeasureNames}
              visibleMeasures={leaderboardMeasureNames}
              {whereFilter}
              {dimensionThresholdFilters}
              {tableWidth}
              dimensionColumnWidth={DIMENSION_COLUMN_WIDTH}
              sortedAscending={false}
              sortType={SortType.VALUE}
              filterExcludeMode={false}
              timeRange={$timeAndFilterStore.timeRange}
              comparisonTimeRange={$timeAndFilterStore.comparisonTimeRange}
              {dimension}
              {parentElement}
              {suppressTooltip}
              timeControlsReady={true}
              allowExpandTable={false}
              allowDimensionComparison={false}
              selectedValues={getSelectedValues(dimension.name)}
              isBeingCompared={false}
              formatters={measureFormatters}
              toggleSort={() => {}}
              toggleDimensionValueSelection={() => {}}
              sortBy={null}
              measureLabel={(measureName) =>
                visibleMeasures.find((m) => m.name === measureName)
                  ?.displayName || measureName}
              leaderboardMeasureCountFeatureFlag={true}
            />
          {/if}
        {/each}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .leaderboard-grid {
    @apply flex flex-row flex-wrap gap-4 overflow-x-auto;
  }
</style>
