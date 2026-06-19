<script lang="ts">
  import CanvasFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/CanvasFilterChipsReadOnly.svelte";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";

  // Off-screen, read-only render of the canvas's active time range and filters.
  // It mirrors the explore "Download as PNG" summary so the PDF capture shows a
  // static, undistorted filter-bar summary instead of the live interactive bar.
  // The dashboard title and timestamp are drawn as vector text in assemble.ts.
  export let canvasName: string;
  export let instanceId: string;
  export let maxWidth: number;

  $: ({
    canvasEntity: {
      timeManager: {
        state: {
          interval: intervalStore,
          grainStore,
          timeZoneStore,
          comparisonIntervalStore,
          showTimeComparisonStore,
        },
      },
      filterManager: { activeUIFiltersStore },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: grain = $grainStore;
  // Exact, resolved range (e.g. "Jan 1 – Jan 7, 2024"), never the relative alias.
  $: formattedTimeRange = $intervalStore
    ? prettyFormatTimeRange($intervalStore, grain)
    : "";
  $: formattedComparisonRange =
    $showTimeComparisonStore && $comparisonIntervalStore
      ? prettyFormatTimeRange($comparisonIntervalStore, grain)
      : "";

  // Drop pinned-but-empty filters (interactive affordances with no applied
  // value); a static PDF should only show filters that actually constrain data.
  $: uiFilters = {
    ...$activeUIFiltersStore,
    dimensionFilters: new Map(
      [...$activeUIFiltersStore.dimensionFilters].filter(
        ([, f]) =>
          (f.selectedValues?.length ?? 0) > 0 ||
          (!!f.inputText && f.inputText.length > 0),
      ),
    ),
    measureFilters: new Map(
      [...$activeUIFiltersStore.measureFilters].filter(([, f]) => !!f.filter),
    ),
  };
  $: hasFilters =
    uiFilters.dimensionFilters.size > 0 || uiFilters.measureFilters.size > 0;
</script>

<div
  id="canvas-pdf-export-header"
  class="flex flex-col gap-y-3 bg-surface-background p-4"
  style:width="{maxWidth}px"
>
  {#if formattedTimeRange}
    <div class="text-sm text-fg-secondary">
      {formattedTimeRange}
      {#if formattedComparisonRange}vs {formattedComparisonRange}{/if}
      <span class="text-fg-muted">· {$timeZoneStore}</span>
    </div>
  {/if}

  {#if hasFilters}
    <CanvasFilterChipsReadOnly {uiFilters} col={false} />
  {/if}
</div>
