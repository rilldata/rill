<script lang="ts">
  import type { LeaderboardComponent } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import { validateLeaderboardSchema } from "@rilldata/web-common/features/canvas/components/leaderboard/selector";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import {
    COMPARISON_COLUMN_WIDTH,
    valueColumn,
  } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-widths";
  import Leaderboard from "@rilldata/web-common/features/dashboards/leaderboard/Leaderboard.svelte";
  import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
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
  let leaderboardWrapperWidth = 0;

  $: ({ instanceId } = $runtime);

  $: ({
    specStore,
    timeAndFilterStore,
    leaderboardState,
    toggleSort,
    parent: { name: canvasName },
    visible,
  } = component);
  $: leaderboardProperties = $specStore;

  $: store = getCanvasStore(canvasName, instanceId);
  $: ({
    canvasEntity: {
      metricsView: {
        getMetricsViewFromName,
        getDimensionsForMetricView,
        getMeasuresForMetricView,
      },
      filters: { isFilterExcludeMode, toggleDimensionValueSelection },
    },
  } = store);

  $: {
    metricsViewName = leaderboardProperties.metrics_view;
    leaderboardMeasureNames = leaderboardProperties.measures ?? [];
    dimensionNames = leaderboardProperties.dimensions ?? [];
    numRows = leaderboardProperties.num_rows ?? 7;
  }

  $: metricsViewQuery = getMetricsViewFromName(metricsViewName);

  $: schema = validateLeaderboardSchema(
    leaderboardProperties,
    $metricsViewQuery,
  );

  $: ({ showTimeComparison, comparisonTimeRange, timeRange, where } =
    $timeAndFilterStore);

  $: ({ dimensionFilters: whereFilter, dimensionThresholdFilters } =
    splitWhereFilter(where));

  $: allDimensions = getDimensionsForMetricView(metricsViewName);
  $: allMeasures = getMeasuresForMetricView(metricsViewName);

  $: visibleDimensions = dimensionNames
    .map((name) =>
      $allDimensions.find((d) => (d.name || (d.column as string)) === name),
    )
    .filter((d) => d !== undefined);

  $: visibleMeasures = leaderboardMeasureNames
    .map((lm) => $allMeasures.find((m) => m.name === lm))
    .filter(Boolean) as MetricsViewSpecMeasure[];

  $: measureFormatters = Object.fromEntries(
    visibleMeasures.map((m) => [
      m.name,
      createMeasureValueFormatter<null | undefined>(m),
    ]),
  );

  // Reset column widths when the measure changes
  $: if (leaderboardMeasureNames) {
    valueColumn.reset();
  }

  $: totalContextWidth = leaderboardMeasureNames.reduce(
    (sum, measureName) =>
      sum +
      $valueColumn +
      (showTimeComparison ? COMPARISON_COLUMN_WIDTH * 2 : 0) +
      (isValidPercentOfTotal(measureName) ? COMPARISON_COLUMN_WIDTH : 0),
    0,
  );

  $: dimensionColumnWidth = getDimensionColumnWidth(
    leaderboardWrapperWidth,
    totalContextWidth,
  );

  $: estimatedTableWidth = MIN_DIMENSION_COLUMN_WIDTH + totalContextWidth;

  $: ({
    title,
    description,
    show_description_as_tooltip,
    time_filters,
    dimension_filters,
  } = leaderboardProperties);

  $: filters = {
    time_filters,
    dimension_filters,
  };

  $: ({ leaderboardSortByMeasureName, sortDirection, sortType } =
    $leaderboardState);

  function isValidPercentOfTotal(measureName: string) {
    return (
      visibleMeasures.find((m) => m.name === measureName)
        ?.validPercentOfTotal ?? false
    );
  }
</script>

{#if schema.isValid}
  <ComponentHeader
    {component}
    {title}
    {description}
    showDescriptionAsTooltip={show_description_as_tooltip}
    {filters}
  />

  <div
    class="h-fit p-0 grow relative"
    class:!p-0={visibleDimensions.length === 1}
  >
    <span class="border-overlay" />
    <div
      class="grid-wrapper gap-px overflow-x-auto"
      style:grid-template-columns="repeat(auto-fit, minmax({estimatedTableWidth +
        LEADERBOARD_WRAPPER_PADDING}px, 1fr))"
    >
      {#each visibleDimensions as dimension (dimension.name)}
        {#if dimension.name}
          <div
            class="leaderboard-wrapper"
            bind:clientWidth={leaderboardWrapperWidth}
          >
            <Leaderboard
              leaderboardShowContextForAllMeasures
              timeControlsReady
              slice={numRows}
              {instanceId}
              visible={$visible}
              {isValidPercentOfTotal}
              {metricsViewName}
              leaderboardSortByMeasureName={leaderboardSortByMeasureName ??
                leaderboardMeasureNames?.[0]}
              leaderboardMeasures={visibleMeasures}
              {whereFilter}
              {dimensionThresholdFilters}
              tableWidth={dimensionColumnWidth + totalContextWidth}
              {dimensionColumnWidth}
              sortedAscending={sortDirection === SortDirection.ASCENDING}
              {sortType}
              filterExcludeMode={$isFilterExcludeMode(dimension.name)}
              {timeRange}
              comparisonTimeRange={showTimeComparison
                ? comparisonTimeRange
                : undefined}
              {dimension}
              allowExpandTable={false}
              allowDimensionComparison={false}
              selectedValues={selectedDimensionValues(
                instanceId,
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
              measureLabel={(measureName) =>
                visibleMeasures.find((m) => m.name === measureName)
                  ?.displayName || measureName}
            />
          </div>
        {/if}
      {/each}
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
    @apply relative p-4 pr-6 grid outline outline-1 outline-gray-200;
  }

  .border-overlay {
    @apply absolute border-[12.5px] pointer-events-none border-surface size-full;
    z-index: 20;
  }

  @container component-container (inline-size < 440px) {
    .grid-wrapper {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
