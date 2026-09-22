<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import { resolveComparisonRange } from "@rilldata/web-common/features/canvas/components/comparison-range";
  import type { ComponentFilterProperties } from "@rilldata/web-common/features/canvas/components/types";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";

  export let component: BaseCanvasComponent;

  $: ({
    specStore,
    timeAndFilterStore,
    localExpressionFilters,
    localTimeControls,
  } = component);

  $: ({ interval: intervalStore, rangeStore, grainStore } = localTimeControls);

  $: activeTimeGrain = $grainStore;

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

  $: ({ showTimeComparison, comparisonTimeRangeState, timeGrain } =
    $timeAndFilterStore);
  $: selectedComparisonTimeRange =
    comparisonTimeRangeState?.selectedComparisonTimeRange;

  // Only a comparison the component sets itself counts as a local filter.
  $: hasLocalComparison =
    resolveComparisonRange($specStore as ComponentFilterProperties).mode ===
    "local";

  $: displayComparisonTimeRange =
    hasLocalComparison && showTimeComparison && selectedComparisonTimeRange
      ? <V1TimeRange>{
          name: selectedComparisonTimeRange.name,
          start: selectedComparisonTimeRange.start.toISOString(),
          end: selectedComparisonTimeRange.end.toISOString(),
          interval: timeGrain,
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
