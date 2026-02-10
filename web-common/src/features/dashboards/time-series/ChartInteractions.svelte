<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    type DashboardTimeControls,
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { DateTime, Interval } from "luxon";
  import ScrubActionMenu from "./ScrubActionMenu.svelte";

  export let exploreName: string;
  export let showComparison = false;
  export let timeGrain: V1TimeGrain | undefined;
  export let measureSelectionEnabled = false;

  let priorRange: DashboardTimeControls | null = null;

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
    validSpecStore,
    metricsViewName,
  } = StateManagers;

  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: ({ selectedScrubRange } = $dashboardStore);

  $: selectedSubRange =
    selectedScrubRange?.start && selectedScrubRange?.end
      ? getOrderedStartEnd(selectedScrubRange.start, selectedScrubRange.end)
      : null;

  $: subInterval = selectedSubRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedSubRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedSubRange.end).setZone(activeTimeZone),
      )
    : null;

  function onKeyDown(e: KeyboardEvent) {
    const target = e.target as HTMLElement;

    if (
      ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName) ||
      target.isContentEditable
    ) {
      return;
    }

    const isMac = window.navigator.userAgent.includes("Macintosh");
    const isExplainKey = e.key === "e" && !e.metaKey && !e.ctrlKey;

    if (e.key === "ArrowLeft" && !e.metaKey && !e.altKey) {
      if ($canPanLeft) {
        const panRange = $getNewPanRange("left");
        if (panRange) updatePanRange(panRange.start, panRange.end);
      }
    } else if (e.key === "ArrowRight" && !e.metaKey && !e.altKey) {
      if ($canPanRight) {
        const panRange = $getNewPanRange("right");
        if (panRange) updatePanRange(panRange.start, panRange.end);
      }
    } else if ($dashboardStore?.selectedScrubRange?.end) {
      if (e.key === "z" && !e.metaKey && !e.ctrlKey) {
        zoomScrub();
      } else if (
        !$dashboardStore.selectedScrubRange?.isScrubbing &&
        e.key === "Escape"
      ) {
        metricsExplorerStore.setSelectedScrubRange(exploreName, undefined);
      } else if (isExplainKey && measureSelectionEnabled) {
        measureSelection.startAnomalyExplanationChat($metricsViewName);
      }
    } else if (
      priorRange &&
      e.key === "z" &&
      ((isMac && e.metaKey) || (!isMac && e.ctrlKey))
    ) {
      e.preventDefault();
      undoZoom();
    } else if (isExplainKey && measureSelectionEnabled) {
      measureSelection.startAnomalyExplanationChat($metricsViewName);
    }
  }

  function updatePanRange(start: Date, end: Date) {
    if (!timeGrain) return;
    const timeRange = {
      name: TimeRangePreset.CUSTOM,
      start: start,
      end: end,
    };

    const comparisonTimeRange = showComparison
      ? ({
          name: TimeComparisonOption.CONTIGUOUS,
        } as DashboardTimeControls) // FIXME wrong typecasting across application
      : undefined;

    metricsExplorerStore.selectTimeRange(
      exploreName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
      $validSpecStore.data?.metricsView ?? {},
    );
  }

  function zoomScrub() {
    if (
      selectedScrubRange?.start instanceof Date &&
      selectedScrubRange?.end instanceof Date
    ) {
      if ($dashboardStore.selectedTimeRange) {
        priorRange = $dashboardStore.selectedTimeRange;
      }

      const { start, end } = getOrderedStartEnd(
        selectedScrubRange.start,
        selectedScrubRange.end,
      );
      metricsExplorerStore.setSelectedTimeRange(exploreName, {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      });

      window.addEventListener("click", cancelUndo, true);
    }
  }

  function clearPriorRange() {
    priorRange = null;
  }

  function undoZoom() {
    if (priorRange) {
      metricsExplorerStore.setSelectedTimeRange(exploreName, priorRange);
      clearPriorRange();
    }
  }

  function cancelUndo(e: MouseEvent) {
    window.removeEventListener("click", cancelUndo, true);

    if (!priorRange) {
      return;
    }

    clearPriorRange();
  }
</script>

<ScrubActionMenu
  {subInterval}
  {timeGrain}
  metricsViewName={$metricsViewName}
  {measureSelectionEnabled}
  onZoom={zoomScrub}
/>

<!-- Only to be used on singleton components to avoid multiple state dispatches -->
<svelte:window on:keydown={onKeyDown} />
