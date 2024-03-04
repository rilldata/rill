<script lang="ts">
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import { FloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Button } from "@rilldata/web-common/components/button";
  import Zoom from "@rilldata/web-common/components/icons/Zoom.svelte";
  import {
    useDashboardStore,
    metricsExplorerStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";

  export let metricViewName;

  let axisTop;
  $: dashboardStore = useDashboardStore(metricViewName);

  function onKeyDown(e) {
    if ($dashboardStore?.selectedScrubRange?.end) {
      // if key Z is pressed, zoom the scrub
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
