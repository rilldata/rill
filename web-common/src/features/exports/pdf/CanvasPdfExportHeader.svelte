<script lang="ts">
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // Off-screen, read-only render of the canvas's active time range and filters.
  // It mirrors the explore "Download as PNG" summary so the PDF capture shows a
  // static, undistorted filter-bar summary instead of the live interactive bar.
  // The dashboard title and timestamp are drawn as vector text in assemble.ts.
  let {
    canvasName,
    instanceId,
    maxWidth,
  }: { canvasName: string; instanceId: string; maxWidth: number } = $props();

  let {
    canvasEntity: { expressionFilterManager, timeFilterManager },
  } = $derived(getCanvasStore(canvasName, instanceId));

  let {
    timeGrain,
    timeZone,
    interval,

    showComparison,
    comparisonInterval,
  } = $derived(timeFilterManager);

  // Exact, resolved range (e.g. "Jan 1 – Jan 7, 2024"), never the relative alias.
  let formattedTimeRange = $derived(
    interval ? prettyFormatTimeRange(interval, timeGrain) : "",
  );
  let formattedComparisonRange = $derived(
    showComparison && comparisonInterval
      ? prettyFormatTimeRange(comparisonInterval, timeGrain)
      : "",
  );
</script>

<div
  id="canvas-pdf-export-header"
  class="flex flex-col gap-y-3 bg-surface-background p-4"
  style:width="{maxWidth}px"
>
  {#if formattedTimeRange}
    <div class="text-sm text-fg-secondary">
      {formattedTimeRange}
      {#if formattedComparisonRange}
        <span> {m.time_vs()} {formattedComparisonRange}</span>
      {/if}
      <span class="text-fg-muted">· {timeZone}</span>
    </div>
  {/if}

  {#if expressionFilterManager.hasSomeFilter}
    <ReadonlyExpressionFilters
      {expressionFilterManager}
      ariaLabel={undefined}
    />
  {/if}
</div>
