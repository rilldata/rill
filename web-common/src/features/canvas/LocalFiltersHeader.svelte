<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import { INHERIT_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";

  let { component }: { component: BaseCanvasComponent } = $props();

  let { specStore, localExpressionFilters, localTimeFilterManager } =
    $derived(component);

  let {
    urlTimeRange,
    urlTimeGrain,
    urlComparisonTimeRange,
    timeRange,
    timeGrain,
    timeZone,
    interval,
    showComparison,
    comparisonTimeRange,
    comparisonInterval,
  } = $derived(localTimeFilterManager);

  // Only a range, comparison or grain the component sets itself counts as a local filter.
  let hasLocalTimeRange = $derived(
    !!urlTimeRange && urlTimeRange !== INHERIT_TIME_RANGE_ALIAS,
  );
  let hasLocalComparison = $derived(
    !!urlComparisonTimeRange &&
      urlComparisonTimeRange !== INHERIT_TIME_RANGE_ALIAS,
  );

  let displayTimeRange = $derived<V1TimeRange | undefined>(
    hasLocalTimeRange && interval
      ? {
          start: interval.start.toUTC().toISO(),
          end: interval.end.toUTC().toISO(),
          timeZone,
          isoDuration: timeRange,
        }
      : undefined,
  );

  let displayComparisonTimeRange = $derived<V1TimeRange | undefined>(
    hasLocalComparison && showComparison && comparisonInterval
      ? <V1TimeRange>{
          name: comparisonTimeRange,
          start: comparisonInterval.start.toUTC().toISO(),
          end: comparisonInterval.end.toUTC().toISO(),
          interval: timeGrain,
        }
      : undefined,
  );
</script>

{#if "metrics_view" in $specStore}
  <div
    class="flex items-center gap-x-2 w-full max-w-full overflow-x-auto chip-scroll-container"
  >
    <Filter size="16px" className="text-fg-secondary" />

    <ReadonlyExpressionFilters
      expressionFilterManager={localExpressionFilters}
      {displayTimeRange}
      {displayComparisonTimeRange}
      displayTimeGrain={urlTimeGrain ? timeGrain : undefined}
      queryTimeStart={interval?.start.toUTC().toISO()}
      queryTimeEnd={interval?.end.toUTC().toISO()}
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
