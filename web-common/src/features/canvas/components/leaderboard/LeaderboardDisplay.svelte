<script lang="ts">
  import type { LeaderboardComponent } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import { validateLeaderboardSchema } from "@rilldata/web-common/features/canvas/components/leaderboard/selector";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import {
    COMPARISON_COLUMN_WIDTH,
    deltaColumn,
    valueColumn,
  } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-widths";
  import Leaderboard from "@rilldata/web-common/features/dashboards/leaderboard/Leaderboard.svelte";
  import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { selectedDimensionValues } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import {
    getDimensionColumnWidth,
    LEADERBOARD_WRAPPER_PADDING,
    MIN_DIMENSION_COLUMN_WIDTH,
  } from "./util";

  let { component }: { component: LeaderboardComponent } = $props();

  const runtimeClient = useRuntimeClient();
  const { instanceId } = runtimeClient;

  let leaderboardWrapperWidth = $state(0);

  let specStore = $derived(component.specStore);
  let timeAndFilterStore = $derived(component.timeAndFilterStore);
  let leaderboardState = $derived(component.leaderboardState);
  let toggleSort = $derived(component.toggleSort);
  let dataEnabled = $derived(component.dataEnabled);

  let leaderboardProperties = $derived($specStore);

  let store = $derived(getCanvasStore(component.parent.name, instanceId));
  let metricsViewSelectors = $derived(store.canvasEntity.metricsView);
  let expressionFilterManager = $derived(
    store.canvasEntity.expressionFilterManager,
  );

  let metricsViewName = $derived(leaderboardProperties.metrics_view);
  let leaderboardMeasureNames = $derived(leaderboardProperties.measures ?? []);
  let dimensionNames = $derived(leaderboardProperties.dimensions ?? []);
  let numRows = $derived(leaderboardProperties.num_rows ?? 7);

  let metricsViewQuery = $derived(
    metricsViewSelectors.getMetricsViewFromName(metricsViewName),
  );

  let schema = $derived(
    validateLeaderboardSchema(leaderboardProperties, $metricsViewQuery),
  );

  let showTimeComparison = $derived($timeAndFilterStore.showTimeComparison);
  let comparisonTimeRange = $derived($timeAndFilterStore.comparisonTimeRange);
  let timeRange = $derived($timeAndFilterStore.timeRange);

  let whereFilter = $derived($timeAndFilterStore.where);

  let allDimensions = $derived(
    metricsViewSelectors.getDimensionsForMetricView(metricsViewName),
  );
  let allMeasures = $derived(
    metricsViewSelectors.getMeasuresForMetricView(metricsViewName),
  );

  let visibleDimensions = $derived(
    dimensionNames
      .map((name) =>
        $allDimensions.find((d) => (d.name || (d.column as string)) === name),
      )
      .filter((d) => d !== undefined),
  );

  let visibleMeasures = $derived(
    leaderboardMeasureNames
      .map((lm) => $allMeasures.find((m) => m.name === lm))
      .filter((m) => m !== undefined),
  );

  let measureFormatters = $derived(
    Object.fromEntries(
      visibleMeasures.map((m) => [
        m.name,
        createMeasureValueFormatter<null | undefined>(m),
      ]),
    ),
  );

  let measureTooltipFormatters = $derived(
    Object.fromEntries(
      visibleMeasures.map((m) => [
        m.name,
        createMeasureValueFormatter<null | undefined>(m, "tooltip"),
      ]),
    ),
  );

  let totalContextWidth = $derived(
    leaderboardMeasureNames.reduce(
      (sum, measureName) =>
        sum +
        $valueColumn +
        (showTimeComparison ? $deltaColumn + COMPARISON_COLUMN_WIDTH : 0) +
        (isValidPercentOfTotal(measureName) ? COMPARISON_COLUMN_WIDTH : 0),
      0,
    ),
  );

  let dimensionColumnWidth = $derived(
    getDimensionColumnWidth(leaderboardWrapperWidth, totalContextWidth),
  );

  let estimatedTableWidth = $derived(
    MIN_DIMENSION_COLUMN_WIDTH + totalContextWidth,
  );

  let filters = $derived({
    time_filters: leaderboardProperties.time_filters,
    dimension_filters: leaderboardProperties.dimension_filters,
  });

  let leaderboardSortByMeasureName = $derived(
    $leaderboardState.leaderboardSortByMeasureName,
  );
  let sortDirection = $derived($leaderboardState.sortDirection);
  let sortType = $derived($leaderboardState.sortType);

  // Reset column widths when the measures change. Uses `$effect.pre` so the reset
  // lands before the DOM update, keeping the width derivations above in sync for
  // the same render rather than one frame later.
  $effect.pre(() => {
    if (leaderboardMeasureNames) {
      valueColumn.reset();
      deltaColumn.reset();
    }
  });

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
    title={leaderboardProperties.title}
    description={leaderboardProperties.description}
    showDescriptionAsTooltip={leaderboardProperties.show_description_as_tooltip}
    {filters}
  />

  <div
    class={[
      "h-fit p-0 grow relative",
      visibleDimensions.length === 1 && "!p-0",
    ]}
  >
    <span class="border-overlay"></span>
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
              visible={$dataEnabled}
              {isValidPercentOfTotal}
              {metricsViewName}
              leaderboardSortByMeasureName={leaderboardSortByMeasureName ??
                leaderboardMeasureNames[0]}
              leaderboardMeasures={visibleMeasures}
              {whereFilter}
              tableWidth={dimensionColumnWidth + totalContextWidth}
              {dimensionColumnWidth}
              sortedAscending={sortDirection === SortDirection.ASCENDING}
              {sortType}
              filterExcludeMode={expressionFilterManager.filterManagers.dimensions.find(
                (dfm) => dfm.name === dimension.name,
              )?.exclude ?? false}
              {timeRange}
              comparisonTimeRange={showTimeComparison
                ? comparisonTimeRange
                : undefined}
              {dimension}
              allowExpandTable={false}
              allowDimensionComparison={false}
              selectedValues={selectedDimensionValues(
                runtimeClient,
                [metricsViewName],
                whereFilter,
                dimension.name,
                timeRange.start,
                timeRange.end,
              )}
              isBeingCompared={false}
              formatters={measureFormatters}
              tooltipFormatters={measureTooltipFormatters}
              {toggleSort}
              toggleDimensionValueSelection={async (
                _1,
                value,
                _2,
                exclusive,
              ) => {
                expressionFilterManager.dimensionFilterAction(
                  dimension.name!,
                  (dimensionManager) =>
                    dimensionManager.toggleValue(value, exclusive ?? false),
                );
              }}
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
    @apply absolute border-[12.5px] pointer-events-none border-surface-card size-full;
    z-index: 20;
  }

  @container component-container (inline-size < 440px) {
    .grid-wrapper {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
