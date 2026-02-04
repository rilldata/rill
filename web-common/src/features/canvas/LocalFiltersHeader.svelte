<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { readable, type Readable } from "svelte/store";

  export let component: BaseCanvasComponent;

  let measures: Readable<MetricsViewSpecMeasure[]> = readable([]);
  let dimensions: Readable<MetricsViewSpecDimension[]> = readable([]);

  $: ({
    specStore,
    parent: {
      metricsView: { getDimensionsForMetricView, getMeasuresForMetricView },
    },
    timeAndFilterStore,
    localFilters,
    localTimeControls,
  } = component);

  $: metricsViewName =
    "metrics_view" in $specStore ? $specStore.metrics_view : null;

  $: if (metricsViewName) {
    measures = getMeasuresForMetricView(metricsViewName);
    dimensions = getDimensionsForMetricView(metricsViewName);
  }

  $: ({
    showTimeComparisonStore,
    interval: intervalStore,
    rangeStore,
    grainStore,
    comparisonRangeStore,
    comparisonIntervalStore,
  } = localTimeControls);

  $: showTimeComparison = $showTimeComparisonStore;
  $: activeTimeGrain = $grainStore;

  $: comparisonRange = $comparisonRangeStore;
  $: comparisonInterval = $comparisonIntervalStore;

  $: interval = $intervalStore;
  $: selectedRangeAlias = $rangeStore;

  $: selectedTimeRange = interval
    ? {
        name: selectedRangeAlias,
        start: interval?.start.toJSDate(),
        end: interval?.end.toJSDate(),
        interval: activeTimeGrain,
      }
    : undefined;

  // $: selectedTimeRange = $timeRangeStateStore?.selectedTimeRange;
  $: displayComparisonTimeRange =
    showTimeComparison && comparisonInterval && comparisonRange
      ? <V1TimeRange>{
          name: comparisonRange,
          start: comparisonInterval.start.toISO(),
          end: comparisonInterval.end.toISO(),
          interval: activeTimeGrain,
        }
      : undefined;

  $: ({ parsed } = localFilters);

  $: ({
    dimensionThresholdFilters,
    dimensionFilter,
    dimensionsWithInListFilter,
  } = $parsed);

  $: displayTimeRange = {
    ...$timeAndFilterStore.timeRange,
    isoDuration: selectedTimeRange?.name,
  };

  $: hasTimeFilters = "time_filters" in $specStore && $specStore.time_filters;
</script>

{#if "metrics_view" in $specStore}
  <div
    class="flex items-center gap-x-2 w-full max-w-full overflow-x-auto chip-scroll-container"
  >
    <Filter size="16px" className="text-fg-secondary" />

    <FilterChipsReadOnly
      metricsViewNames={[$specStore.metrics_view]}
      dimensions={$dimensions}
      measures={$measures}
      {dimensionThresholdFilters}
      dimensionsWithInlistFilter={dimensionsWithInListFilter}
      filters={dimensionFilter}
      {displayComparisonTimeRange}
      displayTimeRange={hasTimeFilters ? displayTimeRange : undefined}
      queryTimeStart={selectedTimeRange?.start?.toISOString()}
      queryTimeEnd={selectedTimeRange?.end?.toISOString()}
      hasBoldTimeRange={false}
      chipLayout="scroll"
    />
  </div>
{/if}

<style>
  .chip-scroll-container {
    mask-image: linear-gradient(to right, black 95%, transparent);
    -webkit-mask-image: linear-gradient(to right, black 95%, transparent);
    mask-size: 100% 100%;
    mask-repeat: no-repeat;
    -webkit-mask-size: 100% 100%;
    -webkit-mask-repeat: no-repeat;
  }
</style>
