<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/manager/DimensionFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/manager/MeasureFilter.svelte";
  import { fly } from "svelte/transition";
  import type { ExpressionFilterManager } from "./ExpressionFilterManager.svelte.ts";

  let {
    expressionFilterManager,
    filteredMeasures,

    timeEnd,
    timeStart,
    timeControlsReady,
    timeDimension,

    isUrlTooLongAfterInListFilter,
  }: {
    expressionFilterManager: ExpressionFilterManager;
    // Explore spec can restrict the available measure vs metrics views.
    // This is passed in explore context.
    filteredMeasures?: string[];

    timeStart: string | undefined;
    timeEnd: string | undefined;
    timeControlsReady: boolean | undefined;
    timeDimension: string | undefined;

    isUrlTooLongAfterInListFilter?: (name: string, values: string[]) => boolean;
  } = $props();

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  let metricsViewsProvider = $derived(
    expressionFilterManager.metricsViewsProvider,
  );
  let yamlConfigProvider = $derived(expressionFilterManager.yamlConfigProvider);

  let allDimensions = $derived(metricsViewsProvider.dimensions);

  let filteredSimpleMeasures = $derived(
    filteredMeasures
      ? metricsViewsProvider.simpleMeasures.filter((m) =>
          filteredMeasures.includes(m.name ?? ""),
        )
      : metricsViewsProvider.simpleMeasures,
  );

  let hasFilters = $derived(
    expressionFilterManager.filterManagers.dimensions.length > 0 ||
      expressionFilterManager.filterManagers.measures.length > 0,
  );
  // Required and pinned filters have a chip even without a value, so there is nothing to clear
  // unless a chip actually holds a filter.
  let hasClearableFilters = $derived(
    expressionFilterManager.filterManagers.dimensions.some(
      (dfm) => !!dfm.expr,
    ) ||
      expressionFilterManager.filterManagers.measures.some((mfm) => !!mfm.expr),
  );

  let measureHasFilter = $derived((name: string) =>
    expressionFilterManager.filterManagers.measures.some(
      (mfm) => mfm.name === name,
    ),
  );
  let dimensionHasFilter = $derived((name: string) =>
    expressionFilterManager.filterManagers.dimensions.some(
      (dfm) => dfm.name === name,
    ),
  );
</script>

<div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
    {#if !hasFilters}
      <div
        in:fly={{ duration: 200, x: 8 }}
        class="text-fg-muted grid ml-1 items-center"
        style:min-height={ROW_HEIGHT}
      >
        {m.dashboard_no_filters_selected()}
      </div>
    {:else}
      {#each expressionFilterManager.filterManagers.dimensions as dimensionManager (dimensionManager.name)}
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

      {#each expressionFilterManager.filterManagers.measures as measureManager (measureManager.name)}
        <MeasureFilter
          {measureManager}
          {yamlConfigProvider}
          {allDimensions}
          openOnMount={expressionFilterManager.temporaryFilterName ===
            measureManager.name}
        />
      {/each}
    {/if}

    <FilterButton
      {allDimensions}
      {filteredSimpleMeasures}
      {dimensionHasFilter}
      {measureHasFilter}
      setTemporaryFilterName={(name: string) =>
        expressionFilterManager.addNewFilter(name)}
    />
    <!-- if filters are present, place a chip at the end of the flex container
    that enables clearing all filters -->
    {#if hasClearableFilters}
      <Button type="text" onClick={() => expressionFilterManager.clear()}>
        {m.dashboard_clear_filters()}
      </Button>
    {/if}
  </div>
</div>
