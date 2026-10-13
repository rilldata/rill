<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import type {
    V1TimeGrain,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import TimeGrainReadOnly from "@rilldata/web-common/features/dashboards/filters/TimeGrainReadOnly.svelte";
  import TimeRangeReadOnly from "@rilldata/web-common/features/dashboards/filters/TimeRangeReadOnly.svelte";
  import ReadonlyDimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/ReadonlyDimensionFilter.svelte";
  import ReadonlyMeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/ReadonlyMeasureFilter.svelte";
  import { flip } from "svelte/animate";

  let {
    expressionFilterManager,
    displayTimeRange,
    displayComparisonTimeRange,
    displayTimeGrain = undefined,
    queryTimeStart = undefined,
    queryTimeEnd = undefined,
    hasBoldTimeRange = true,
    chipLayout = "wrap",
    ariaLabel = m.dashboard_readonly_filter_chips_aria(),
    showPinned = false,
  }: {
    expressionFilterManager: ExpressionFilterManager;
    displayTimeRange?: V1TimeRange | undefined;
    displayComparisonTimeRange?: V1TimeRange | undefined;
    // A grain the widget sets itself, shown as its own chip since it applies with or without a local range.
    displayTimeGrain?: V1TimeGrain | undefined;
    queryTimeStart?: string | undefined;
    queryTimeEnd?: string | undefined;
    hasBoldTimeRange?: boolean;
    chipLayout?: "wrap" | "scroll" | "col";
    ariaLabel?: string | undefined;
    showPinned?: boolean;
  } = $props();

  let scrollContainer: HTMLDivElement;

  let { dimensions: sortedDimensionManagers, measures: sortedMeasureManagers } =
    $derived(expressionFilterManager.sortedFilterManagers);

  let nonEmptyDimensionManagers = $derived(
    sortedDimensionManagers.filter(
      (dimensionManager) => showPinned || !!dimensionManager.expr,
    ),
  );
  let nonEmptyMeasureManager = $derived(
    sortedMeasureManagers.filter(
      (measureManager) => showPinned || !!measureManager.expr,
    ),
  );

  function handleWheel(event: WheelEvent) {
    if (chipLayout === "scroll" && event.deltaY !== 0) {
      scrollContainer.scrollLeft += event.deltaY;
      event.preventDefault();
    }
  }
</script>

<div
  class="relative flex flex-row items-center gap-x-2 gap-y-2 w-full max-w-full"
  class:scrollable-chips={chipLayout === "scroll"}
  class:flex-wrap={chipLayout === "wrap"}
  class:flex-col={chipLayout === "col"}
  aria-label={ariaLabel}
  bind:this={scrollContainer}
  onwheel={handleWheel}
>
  {#if displayTimeRange}
    <TimeRangeReadOnly
      timeRange={displayTimeRange}
      comparisonTimeRange={displayComparisonTimeRange}
      {hasBoldTimeRange}
    />
  {/if}

  {#if displayTimeGrain}
    <TimeGrainReadOnly timeGrain={displayTimeGrain} />
  {/if}

  {#each nonEmptyDimensionManagers as dimensionManager (dimensionManager.name)}
    <div animate:flip={{ duration: 200 }}>
      <ReadonlyDimensionFilter
        manager={expressionFilterManager}
        {dimensionManager}
        yamlConfigProvider={expressionFilterManager.yamlConfigProvider}
        timeStart={queryTimeStart}
        timeEnd={queryTimeEnd}
        {showPinned}
      />
    </div>
  {/each}

  {#each nonEmptyMeasureManager as measureManager (measureManager.name)}
    <div animate:flip={{ duration: 200 }}>
      <ReadonlyMeasureFilter
        {measureManager}
        yamlConfigProvider={expressionFilterManager.yamlConfigProvider}
        {showPinned}
      />
    </div>
  {/each}
</div>

<style lang="postcss">
  .scrollable-chips {
    @apply overflow-x-auto whitespace-nowrap;
    @apply overscroll-x-contain pr-2;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }
  .scrollable-chips::-webkit-scrollbar {
    @apply hidden;
  }
</style>
