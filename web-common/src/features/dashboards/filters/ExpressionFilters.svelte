<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import { Button } from "web-common/src/components/button";
  import MeasureFilter from "./measure-filters/MeasureFilter.svelte";
  import { fly } from "svelte/transition";
  import type { ExpressionFilterManager } from "./ExpressionFilterManager.svelte.ts";
  import AddExpressionFilterButton from "@rilldata/web-common/features/dashboards/filters/AddExpressionFilterButton.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import type { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";

  let {
    expressionFilterManager,
    dashboardConfigProvider,

    timeEnd,
    timeStart,
    timeControlsReady,
    timeDimension,

    isUrlTooLongAfterInListFilter,
  }: {
    expressionFilterManager: ExpressionFilterManager;
    dashboardConfigProvider: DashboardConfigProvider;

    timeStart: string | undefined;
    timeEnd: string | undefined;
    timeControlsReady: boolean | undefined;
    timeDimension?: string | undefined;

    isUrlTooLongAfterInListFilter?: (name: string, values: string[]) => boolean;
  } = $props();

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  let { metricsViewsProvider, yamlConfigProvider } = $derived(
    dashboardConfigProvider,
  );

  let { restrictedMeasures, restrictedDimensions } =
    $derived(yamlConfigProvider);

  let allDimensions = $derived(
    restrictedDimensions
      ? metricsViewsProvider.dimensions.filter((d) =>
          restrictedDimensions.includes(d.name!),
        )
      : metricsViewsProvider.dimensions,
  );

  let { dimensions: sortedDimensionManagers, measures: sortedMeasureManagers } =
    $derived(expressionFilterManager.sortedFilterManagers);

  let hasFilters = $derived(
    sortedDimensionManagers.length > 0 || sortedMeasureManagers.length > 0,
  );
  // Required and pinned filters have a chip even without a value,
  // so there is nothing to clear unless a chip actually holds a filter.
  let hasClearableFilters = $derived(
    sortedDimensionManagers.some((dfm) => !!dfm.expr) ||
      sortedMeasureManagers.some((mfm) => !!mfm.expr),
  );

  let excludedDimensions = $derived(
    Object.fromEntries(
      restrictedDimensions
        ? metricsViewsProvider.dimensions
            .filter((d) => !restrictedDimensions.includes(d.name!))
            .map((m) => [m.name!, true])
        : [],
    ),
  );
  let excludedMeasures = $derived(
    Object.fromEntries(
      restrictedMeasures
        ? metricsViewsProvider.measures
            .filter((m) => !restrictedMeasures.includes(m.name!))
            .map((m) => [m.name!, true])
        : [],
    ),
  );
</script>

<div
  class="relative flex flex-row gap-x-2 gap-y-2 items-start pointer-events-auto"
>
  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
    {#if expressionFilterManager.isComplexFilter}
      {#each Object.entries(expressionFilterManager.exprByMetricsView) as [mv, expr] (mv)}
        <AdvancedFilter advancedFilter={expr} />
      {/each}
    {:else}
      {#if !hasFilters}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="text-fg-muted grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          {m.dashboard_no_filters_selected()}
        </div>
      {:else}
        {#each sortedDimensionManagers as dimensionManager (dimensionManager.name)}
          <DimensionFilter
            manager={expressionFilterManager}
            {dimensionManager}
            {yamlConfigProvider}
            {timeStart}
            {timeEnd}
            {timeControlsReady}
            {timeDimension}
            openOnMount={expressionFilterManager.temporaryFilterName ===
              dimensionManager.name}
            isUrlTooLongAfterInListFilter={isUrlTooLongAfterInListFilter
              ? (values) =>
                  isUrlTooLongAfterInListFilter(dimensionManager.name, values)
              : undefined}
          />
        {/each}

        {#each sortedMeasureManagers as measureManager (measureManager.name)}
          <MeasureFilter
            {measureManager}
            {yamlConfigProvider}
            {allDimensions}
            openOnMount={expressionFilterManager.temporaryFilterName ===
              measureManager.name}
          />
        {/each}
      {/if}

      <AddExpressionFilterButton
        {expressionFilterManager}
        {excludedDimensions}
        {excludedMeasures}
      />

      <!-- if filters are present, place a chip at the end of the flex container
      that enables clearing all filters -->
      {#if hasClearableFilters}
        <Button type="text" onClick={() => expressionFilterManager.clear()}>
          {m.dashboard_clear_filters()}
        </Button>
      {/if}
    {/if}
  </div>
</div>
