<script lang="ts">
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { DateTime, Interval } from "luxon";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import RangeDisplay from "../time-controls/super-pill/components/RangeDisplay.svelte";

  let priorRange = $state<string | undefined>(undefined);
  let button = $state<HTMLButtonElement | undefined>(undefined);

  const explainEnabled = measureSelection.getEnabledStore();

  const StateManagers = getStateManagers();
  const { dashboardStore, metricsViewName, timeFilterManager } = StateManagers;

  let { timeGrain, timeZone, canPanLeft, canPanRight } =
    $derived(timeFilterManager);

  let { selectedScrubRange } = $derived($dashboardStore);

  let selectedSubRange = $derived(
    selectedScrubRange?.start && selectedScrubRange?.end
      ? getOrderedStartEnd(selectedScrubRange.start, selectedScrubRange.end)
      : null,
  );

  let subInterval = $derived(
    selectedSubRange
      ? Interval.fromDateTimes(
          DateTime.fromJSDate(selectedSubRange.start).setZone(timeZone),
          DateTime.fromJSDate(selectedSubRange.end).setZone(timeZone),
        )
      : null,
  );

  function onKeyDown(e: KeyboardEvent) {
    const target = e.target as HTMLElement;

    if (
      ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName) ||
      target.isContentEditable
    ) {
      return;
    }

    const isMac = window.navigator.userAgent.includes("Macintosh");
    const isExplainKey =
      $explainEnabled && e.key === "e" && !e.metaKey && !e.ctrlKey;

    if (e.key === "ArrowLeft" && !e.metaKey && !e.altKey) {
      if (canPanLeft) {
        timeFilterManager.onPan("left");
      }
    } else if (e.key === "ArrowRight" && !e.metaKey && !e.altKey) {
      if (canPanRight) {
        timeFilterManager.onPan("right");
      }
    } else if ($dashboardStore?.selectedScrubRange?.end) {
      if (e.key === "z" && !e.metaKey && !e.ctrlKey) {
        zoomScrub();
      } else if (
        !$dashboardStore.selectedScrubRange?.isScrubbing &&
        e.key === "Escape"
      ) {
        timeFilterManager.resetScrubRange();
      } else if (isExplainKey) {
        measureSelection.startAnomalyExplanationChat($metricsViewName);
      }
    } else if (
      priorRange &&
      e.key === "z" &&
      ((isMac && e.metaKey) || (!isMac && e.ctrlKey))
    ) {
      e.preventDefault();
      undoZoom();
    } else if (isExplainKey) {
      measureSelection.startAnomalyExplanationChat($metricsViewName);
    }
  }

  function zoomScrub() {
    if (
      selectedScrubRange?.start instanceof Date &&
      selectedScrubRange?.end instanceof Date
    ) {
      priorRange = timeFilterManager.timeRange;

      const { start, end } = getOrderedStartEnd(
        selectedScrubRange.start,
        selectedScrubRange.end,
      );
      void timeFilterManager.onSelectRange(
        `${start.toISOString()} to ${end.toISOString()}`,
      );

      window.addEventListener("click", cancelUndo, true);
    }
  }

  function clearPriorRange() {
    priorRange = undefined;
  }

  function undoZoom() {
    if (priorRange) {
      void timeFilterManager.onSelectRange(priorRange);
      clearPriorRange();
    }
  }

  function cancelUndo(e: MouseEvent) {
    window.removeEventListener("click", cancelUndo, true);

    if (
      !priorRange ||
      (e.target instanceof HTMLElement && e.target === button)
    ) {
      return;
    }

    clearPriorRange();
  }

  function handleClick() {
    if (priorRange) {
      undoZoom();
    } else {
      zoomScrub();
    }
  }
</script>

{#if priorRange || (subInterval?.isValid && !subInterval.start.equals(subInterval.end))}
  <button
    bind:this={button}
    onclick={(e) => {
      e.stopPropagation();
      handleClick();
    }}
    aria-label={priorRange ? m.chart_undo_zoom() : m.chart_zoom()}
  >
    <div class="content-wrapper">
      <span class="flex-none text-icon-muted">
        <Zoom size="16px" />
      </span>

      {#if subInterval?.isValid && timeGrain}
        <RangeDisplay interval={subInterval} {timeGrain} />
      {/if}

      <span class="font-medium line-clamp-1 flex-none whitespace-nowrap">
        {#if priorRange}
          {m.chart_undo_zoom_label()} (<MetaKey plusses={false} action="Z" />)
        {:else}
          {m.chart_zoom_label()} (Z)
        {/if}
      </span>
    </div>
  </button>
{/if}

<!-- Only to be used on singleton components to avoid multiple state dispatches -->
<svelte:window onkeydown={onKeyDown} />

<style lang="postcss">
  button {
    @apply border rounded-[2px] bg-surface-subtle pointer-events-auto;
    @apply absolute top-0 -translate-x-1/2 z-50;
    /* Center over the plot body, not the full chart (40px right margin for y-axis labels) */
    left: calc(50% - 20px);
  }

  .content-wrapper {
    @apply py-1 px-2 flex gap-x-1 w-fit flex-none;
    @apply pointer-events-none;
  }
</style>
