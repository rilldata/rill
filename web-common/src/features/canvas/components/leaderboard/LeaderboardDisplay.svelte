<script lang="ts">
  import type { LeaderboardComponent } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import {
    COMPARISON_COLUMN_WIDTH,
    valueColumn,
  } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-widths";
  import Leaderboard from "@rilldata/web-common/features/dashboards/leaderboard/Leaderboard.svelte";
  import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { selectedDimensionValuesV2 } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ComponentHeader from "../../ComponentHeader.svelte";

  export let component: LeaderboardComponent;

  const DIMENSION_COLUMN_WIDTH = 164;

  let metricsViewName: string;
  let leaderboardMeasureNames: string[] = [];
  let dimensionNames: string[] = [];
  let numRows = 7;

  let parentElement: HTMLDivElement;
  let suppressTooltip = false;
  let showPercentOfTotal = false; // TODO: update this

  $: ({
    specStore,
    timeAndFilterStore,
    leaderboardState,
    sortByMeasure,
    toggleSort,
    parent: { name: canvasName },
  } = component);
  $: leaderboardProperties = $specStore;

  $: store = getCanvasStore(canvasName);
  $: ({
    canvasEntity: {
      spec: { getDimensionsForMetricView, getMeasuresForMetricView },
      filters: { isFilterExcludeMode, toggleDimensionValueSelection },
    },
  } = store);

  $: ({ instanceId } = $runtime);

  $: ({ showTimeComparison, comparisonTimeRange, timeRange, where } =
    $timeAndFilterStore);

  $: {
    metricsViewName = leaderboardProperties.metrics_view;
    leaderboardMeasureNames = leaderboardProperties.measures ?? [];
    dimensionNames = leaderboardProperties.dimensions ?? [];
    numRows = leaderboardProperties.num_rows ?? 7;
  }

  $: ({ dimensionFilters: whereFilter, dimensionThresholdFilters } =
    splitWhereFilter(where));

  $: allDimensions = getDimensionsForMetricView(metricsViewName);
  $: allMeasures = getMeasuresForMetricView(metricsViewName);

  $: visibleDimensions = $allDimensions.filter((d) =>
    dimensionNames.includes(d.name || (d.column as string)),
  );

  $: visibleMeasures = $allMeasures.filter((m) =>
    leaderboardMeasureNames.includes(m.name as string),
  );

  $: activeMeasureName =
    $sortByMeasure || leaderboardMeasureNames?.[0] || "measure";

  $: measureFormatters = Object.fromEntries(
    visibleMeasures.map((m) => [
      m.name,
      createMeasureValueFormatter<null | undefined>(m),
    ]),
  );

  $: showDeltaPercent = showTimeComparison;

  // Reset column widths when the measure changes
  $: if (leaderboardMeasureNames) {
    valueColumn.reset();
  }

  $: tableWidth =
    DIMENSION_COLUMN_WIDTH +
    $valueColumn +
    (showTimeComparison
      ? COMPARISON_COLUMN_WIDTH * (showDeltaPercent ? 2 : 1)
      : showPercentOfTotal
        ? COMPARISON_COLUMN_WIDTH
        : 0);

  $: ({ title, description, time_filters, dimension_filters } =
    leaderboardProperties);

  $: filters = {
    time_filters,
    dimension_filters,
  };

  function isValidPercentOfTotal(measureName: string) {
    return (
      visibleMeasures.find((m) => m.name === measureName)
        ?.validPercentOfTotal ?? false
    );
  }
</script>

<ComponentHeader {title} {description} {filters} />

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
              sortedAscending={$leaderboardState.sortDirection ===
                SortDirection.ASCENDING}
              sortType={$leaderboardState.sortType}
              filterExcludeMode={$isFilterExcludeMode(dimension.name)}
              {timeRange}
              comparisonTimeRange={showTimeComparison
                ? comparisonTimeRange
                : undefined}
              {dimension}
              {parentElement}
              {suppressTooltip}
              timeControlsReady={true}
              allowExpandTable={false}
              allowDimensionComparison={false}
              selectedValues={selectedDimensionValuesV2(
                $runtime.instanceId,
                [metricsViewName],
                whereFilter,
                dimension.name,
                timeRange.start,
                timeRange.end,
              )}
              isBeingCompared={false}
              formatters={measureFormatters}
              {toggleSort}
              {toggleDimensionValueSelection}
              sortBy={$sortByMeasure}
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
