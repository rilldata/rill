<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import { Button } from "web-common/src/components/button";
  import MeasureFilter from "./measure-filters/MeasureFilter.svelte";
  import { fly } from "svelte/transition";
  import type { ExpressionFilterManager } from "./ExpressionFilterManager.svelte.ts";
  import AddExpressionFilterButton from "@rilldata/web-common/features/dashboards/filters/AddExpressionFilterButton.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import { onDestroy } from "svelte";

  let {
    expressionFilterManager,
    filteredMeasures,
    filteredDimensions,

    timeEnd,
    timeStart,
    timeControlsReady,
    timeDimension,

    selfSync = false,

    isUrlTooLongAfterInListFilter,
  }: {
    expressionFilterManager: ExpressionFilterManager;
    // Explore spec can restrict the available dimension/measure vs metrics views.
    // This is passed in explore context.
    filteredDimensions?: string[];
    filteredMeasures?: string[];

    timeStart: string | undefined;
    timeEnd: string | undefined;
    timeControlsReady: boolean | undefined;
    timeDimension?: string | undefined;

    selfSync?: boolean;

    isUrlTooLongAfterInListFilter?: (name: string, values: string[]) => boolean;
  } = $props();

  let stateChangeUnsub: () => void = () => {};
  // svelte-ignore state_referenced_locally
  if (selfSync) {
    // Sync back url params
    expressionFilterManager.createListener();
    stateChangeUnsub = expressionFilterManager.on("state-changed", () =>
      expressionFilterManager.foldManagersIntoParam(),
    );
  }

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  let metricsViewsProvider = $derived(
    expressionFilterManager.metricsViewsProvider,
  );
  let yamlConfigProvider = $derived(expressionFilterManager.yamlConfigProvider);

  let allDimensions = $derived(
    filteredDimensions
      ? metricsViewsProvider.dimensions.filter((d) =>
          filteredDimensions.includes(d.name!),
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
      filteredDimensions
        ? metricsViewsProvider.dimensions
            .filter((d) => !filteredDimensions.includes(d.name!))
            .map((m) => [m.name!, true])
        : [],
    ),
  );
  let excludedMeasures = $derived(
    Object.fromEntries(
      filteredMeasures
        ? metricsViewsProvider.measures
            .filter((m) => !filteredMeasures.includes(m.name!))
            .map((m) => [m.name!, true])
        : [],
    ),
  );

  onDestroy(() => {
    stateChangeUnsub();
  });
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
