<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";

  export let component: BaseCanvasComponent;

  $: ({
    specStore,
    timeAndFilterStore,
    localExpressionFilters,
    localTimeControls,
  } = component);

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

    <ReadonlyExpressionFilters
      expressionFilterManager={localExpressionFilters}
      displayTimeRange={hasTimeFilters ? displayTimeRange : undefined}
      {displayComparisonTimeRange}
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
