<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    DashboardTimeControls,
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";

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

  function onKeyDown(e: KeyboardEvent) {
    const targetTagName = (e.target as HTMLElement).tagName;
    if (["INPUT", "TEXTAREA", "SELECT"].includes(targetTagName)) {
      return;
    }
    if (e.key === "ArrowLeft") {
      if ($canPanLeft) {
        const panRange = $getNewPanRange("left");
        if (panRange) updatePanRange(panRange.start, panRange.end);
      }
    } else if (e.key === "ArrowRight") {
      if ($canPanRight) {
        const panRange = $getNewPanRange("right");
        if (panRange) updatePanRange(panRange.start, panRange.end);
      }
    } else if ($dashboardStore?.selectedScrubRange?.end) {
      if (e.key === "z") {
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

    metricsExplorerStore.selectTimeRange(
      metricViewName,
      timeRange,
      timeGrain,
      comparisonTimeRange,
    );
  }

  function zoomScrub() {
    if (
      $dashboardStore?.selectedScrubRange?.start instanceof Date &&
      $dashboardStore?.selectedScrubRange?.end instanceof Date
    ) {
      const { start, end } = getOrderedStartEnd(
        $dashboardStore.selectedScrubRange.start,
        $dashboardStore.selectedScrubRange.end,
      );
      metricsExplorerStore.setSelectedTimeRange(metricViewName, {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      });
    }
  }
</script>

{#if $dashboardStore?.selectedScrubRange?.end && !$dashboardStore?.selectedScrubRange?.isScrubbing}
  <div class="absolute flex justify-center left-1/2 -top-8 -translate-x-1/2">
    <Button compact type="highlighted" on:click={() => zoomScrub()}>
      <div class="flex items-center gap-x-2">
        <Zoom size="16px" />
        Zoom
        <span class="font-semibold">(Z)</span>
      </div>
    </Button>
  </div>
{/if}

<!-- Only to be used on singleton components to avoid multiple state dispatches -->
<svelte:window on:keydown={onKeyDown} />
