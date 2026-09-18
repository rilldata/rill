<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import ReadonlyDimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/ReadonlyDimensionFilter.svelte";
  import ReadonlyMeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/ReadonlyMeasureFilter.svelte";
  import { flip } from "svelte/animate";
  import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import ReadonlyTimeFilters from "@rilldata/web-common/features/dashboards/time-controls/ReadonlyTimeFilters.svelte";

  let {
    expressionFilterManager,
    timeFilterManager,
    hideTimePills = false,
    hasBoldTimeRange = true,
    chipLayout = "wrap",
    ariaLabel = m.dashboard_readonly_filter_chips_aria(),
    showPinned = false,
  }: {
    expressionFilterManager: ExpressionFilterManager;
    timeFilterManager?: TimeFilterManager;
    // Additional control to pass timeFilterManager but not show the pills
    hideTimePills?: boolean;
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
  {#if timeFilterManager && !hideTimePills}
    <ReadonlyTimeFilters {timeFilterManager} {hasBoldTimeRange} />
  {/if}

  {#each nonEmptyDimensionManagers as dimensionManager (dimensionManager.name)}
    <div animate:flip={{ duration: 200 }}>
      <ReadonlyDimensionFilter
        manager={expressionFilterManager}
        {dimensionManager}
        yamlConfigProvider={expressionFilterManager.yamlConfigProvider}
        timeStart={timeFilterManager?.timeStart}
        timeEnd={timeFilterManager?.timeEnd}
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
