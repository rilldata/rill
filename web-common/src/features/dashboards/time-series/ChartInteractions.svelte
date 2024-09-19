<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RangeDisplay from "../time-controls/super-pill/components/RangeDisplay.svelte";
  import { Interval, DateTime } from "luxon";

  export let metricViewName: string;
  export let showComparison = false;
  export let timeGrain: V1TimeGrain | undefined;

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    selectors: {
      charts: { canPanLeft, canPanRight, getNewPanRange },
    },
  } = StateManagers;

  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: metricsView = useMetricsView($runtime.instanceId, metricViewName);

  $: validSpec = $metricsView?.data;

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
    const targetTagName = (e.target as HTMLElement).tagName;
    if (["INPUT", "TEXTAREA", "SELECT"].includes(targetTagName)) {
      return;
    }
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
        metricsExplorerStore.setSelectedScrubRange(metricViewName, undefined);
      }
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

    if (validSpec)
      metricsExplorerStore.selectTimeRange(
        metricViewName,
        timeRange,
        timeGrain,
        comparisonTimeRange,
        validSpec.measures,
      );
  }

  function zoomScrub() {
    if (
      selectedScrubRange?.start instanceof Date &&
      selectedScrubRange?.end instanceof Date
    ) {
      const { start, end } = getOrderedStartEnd(
        selectedScrubRange.start,
        selectedScrubRange.end,
      );
      metricsExplorerStore.setSelectedTimeRange(metricViewName, {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      });
    }
  }
</script>

{#if $dashboardStore?.selectedScrubRange?.end}
  <div
    class="absolute flex justify-center left-1/2 -top-8 -translate-x-1/2 z-50 bg-white"
  >
    <Button compact type="plain" on:click={() => zoomScrub()}>
      <div class="flex items-center gap-x-2">
        <span class="flex-none">
          <Zoom size="16px" />
        </span>
        {#if subInterval?.isValid && timeGrain}
          <RangeDisplay interval={subInterval} grain={timeGrain} />
        {/if}
        <span class="font-semibold">(Z)</span>
      </div>
    </Button>
  </div>
{/if}

<!-- Only to be used on singleton components to avoid multiple state dispatches -->
<svelte:window on:keydown={onKeyDown} />
