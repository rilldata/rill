<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import type { LeaderboardComponent } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import { validateLeaderboardSchema } from "@rilldata/web-common/features/canvas/components/leaderboard/selector";
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
  import {
    getDimensionColumnWidth,
    LEADERBOARD_WRAPPER_PADDING,
    MIN_DIMENSION_COLUMN_WIDTH,
  } from "./util";

  export let component: LeaderboardComponent;

  let metricsViewName: string;
  let leaderboardMeasureNames: string[] = [];
  let dimensionNames: string[] = [];
  let numRows = 7;

  let parentElement: HTMLDivElement;
  let suppressTooltip = false;
  let showPercentOfTotal = false; // TODO: update this
  let leaderboardWrapperWidth = 0;

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
      spec: {
        getMetricsViewFromName,
        getDimensionsForMetricView,
        getMeasuresForMetricView,
      },
      filters: { isFilterExcludeMode, toggleDimensionValueSelection },
    },
  } = store);

  $: ({ instanceId } = $runtime);

  $: {
    metricsViewName = leaderboardProperties.metrics_view;
    leaderboardMeasureNames = leaderboardProperties.measures ?? [];
    dimensionNames = leaderboardProperties.dimensions ?? [];
    numRows = leaderboardProperties.num_rows ?? 7;
  }

  $: _metricViewSpec = getMetricsViewFromName(metricsViewName);
  $: metricsViewSpec = $_metricViewSpec.metricsView;

  $: schema = validateLeaderboardSchema(leaderboardProperties, metricsViewSpec);

  $: ({ showTimeComparison, comparisonTimeRange, timeRange, where } =
    $timeAndFilterStore);

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

  $: contextColumnWidth =
    $valueColumn +
    (showTimeComparison
      ? COMPARISON_COLUMN_WIDTH * (showDeltaPercent ? 2 : 1)
      : showPercentOfTotal
        ? COMPARISON_COLUMN_WIDTH
        : 0);

  $: dimensionColumnWidth = getDimensionColumnWidth(
    leaderboardWrapperWidth,
    contextColumnWidth,
    leaderboardMeasureNames,
  );

  $: tableWidthForPctBars = dimensionColumnWidth + contextColumnWidth;

  $: tableWidth =
    MIN_DIMENSION_COLUMN_WIDTH +
    contextColumnWidth * leaderboardMeasureNames.length;

  $: hasOverflow = tableWidth > parentElement?.clientWidth;

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

{#if schema.isValid}
  <ComponentHeader {title} {description} {filters} />

  <div
    class="h-fit p-0 grow relative"
    class:!p-0={visibleDimensions.length === 1}
  >
    <span class="border-overlay" />
    <div
      bind:this={parentElement}
      class="grid-wrapper gap-px size-full overflow-x-auto"
      style:grid-template-columns="repeat(auto-fit, minmax({tableWidth +
        LEADERBOARD_WRAPPER_PADDING}px, 1fr))"
      on:scroll={() => {
        suppressTooltip = true;
      }}
      on:scrollend={() => {
        suppressTooltip = false;
      }}
    >
      {#if parentElement}
        {#each visibleDimensions as dimension (dimension.name)}
          {#if dimension.name}
            <div
              class="leaderboard-wrapper"
              class:leaderboard-outline={!hasOverflow}
              bind:clientWidth={leaderboardWrapperWidth}
            >
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
                tableWidth={tableWidthForPctBars}
                {dimensionColumnWidth}
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
            </div>
          {/if}
        {/each}
      {/if}
    </div>
  </div>
{:else}
  <ComponentError error={schema.error} />
{/if}

<style lang="postcss">
  .grid-wrapper {
    @apply size-full grid;
    grid-auto-rows: auto;
  }

  .leaderboard-wrapper {
    @apply relative p-4 pr-6 grid justify-center;
  }

  .leaderboard-outline {
    @apply outline outline-1 outline-gray-200;
  }

  .border-overlay {
    @apply absolute border-[12.5px] pointer-events-none border-white size-full;
    z-index: 20;
  }
</style>
