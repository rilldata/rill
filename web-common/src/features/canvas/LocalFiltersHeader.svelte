<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
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

  $: ({ showTimeComparison, timeRangeStateStore } = localTimeControls);

  $: selectedTimeRange = $timeRangeStateStore?.selectedTimeRange;

  $: ({ whereFilter, dimensionThresholdFilters, dimensionsWithInlistFilter } =
    localFilters);

  $: displayTimeRange = {
    ...$timeAndFilterStore.timeRange,
    isoDuration: selectedTimeRange?.name,
  };

  $: displayComparisonTimeRange = $timeAndFilterStore.comparisonTimeRange;

  $: hasTimeFilters = "time_filters" in $specStore && $specStore.time_filters;
</script>

{#if "metrics_view" in $specStore}
  <div
    class="flex items-center gap-x-2 w-full max-w-full overflow-x-auto chip-scroll-container"
  >
    <Filter size="16px" className="text-gray-400" />
    <FilterChipsReadOnly
      metricsViewNames={[$specStore.metrics_view]}
      dimensions={$dimensions}
      measures={$measures}
      dimensionThresholdFilters={$dimensionThresholdFilters}
      dimensionsWithInlistFilter={$dimensionsWithInlistFilter}
      filters={$whereFilter}
      displayComparisonTimeRange={$showTimeComparison
        ? displayComparisonTimeRange
        : undefined}
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
