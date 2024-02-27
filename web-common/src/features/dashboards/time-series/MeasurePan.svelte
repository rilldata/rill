<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import type { PlotConfig } from "@rilldata/web-common/components/data-graphic/utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";

  import {
    getDurationFromMS,
    getOffset,
    getTimeWidth,
  } from "@rilldata/web-common/lib/time/transforms";
  import {
    TimeOffsetType,
    TimeRange,
  } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher, getContext } from "svelte";
  import type { Writable } from "svelte/store";

  export let hovering = true;

  type PanDirection = "left" | "right";

  const dispatch = createEventDispatcher();
  const plotConfig: Writable<PlotConfig> = getContext(contexts.config);

  const StateManagers = getStateManagers();
  const { dashboardStore } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  $: y1 = $plotConfig.plotTop + $plotConfig.top - 20;
  $: y2 = $plotConfig.plotBottom + $plotConfig.bottom - 20;

  $: midY = (y1 + y2) / 2;

  $: x1 = $plotConfig.plotLeft + $plotConfig.left - 20;
  $: x2 = $plotConfig.plotRight - 10;

  function isRangeValid(
    start: Date,
    end: Date,
    allTimeRange: TimeRange,
    direction: PanDirection,
  ) {
    if (direction === "right" && start.getTime() > allTimeRange.end.getTime()) {
      return false;
    } else if (
      direction === "left" &&
      end.getTime() < allTimeRange.start.getTime()
    ) {
      return false;
    }
    return true;
  }

  function panCharts(direction: PanDirection) {
    const selectedTimeRange = $timeControlsStore?.selectedTimeRange;
    if (!selectedTimeRange) return;

    const timeZone = $dashboardStore?.selectedTimezone || "UTC";
    const { start, end, interval } = selectedTimeRange;
    const allTimeRange = $timeControlsStore?.allTimeRange;

    if (!allTimeRange || !interval || !start || !end) return;

    const offsetType =
      direction === "left" ? TimeOffsetType.SUBTRACT : TimeOffsetType.ADD;

    const currentRangeWidth = getTimeWidth(start, end);
    const panAmount = getDurationFromMS(currentRangeWidth);

    const newStart = getOffset(
      selectedTimeRange?.start,
      panAmount,
      offsetType,
      timeZone,
    );
    const newEnd = getOffset(
      selectedTimeRange?.end,
      panAmount,
      offsetType,
      timeZone,
    );

    const isValid = isRangeValid(newStart, newEnd, allTimeRange, direction);
    if (!isValid) return;

    dispatch("pan", { start: newStart, end: newEnd });
  }
</script>

{#if hovering}
  <WithGraphicContexts>
    <g transform={`translate(${x1}, ${midY})`} class="pan-controls">
      <!-- Left Pan Button -->
      <path
        role="presentation"
        d="M9.335 16.795L21.678 5.756C22.129 5.352 22.844 5.672 22.844 6.277L22.844 27.342C22.844 27.948 22.128 28.268 21.677 27.863L9.335 16.795Z"
        class="pan-button"
        on:click|self={() => panCharts("left")}
      />
    </g>
    <g transform={`translate(${x2}, ${midY})`} class="pan-controls">
      <!-- Right Pan Button -->
      <path
        role="presentation"
        d="M24.265 16.805L11.922 27.844C11.471 28.248 10.756 27.928 10.756 27.323L10.756 6.258C10.756 5.652 11.472 5.332 11.923 5.737L24.265 16.805Z"
        class="pan-button"
        on:click|self={() => panCharts("right")}
      />
    </g>
  </WithGraphicContexts>
{/if}

<style>
  .pan-button {
    cursor: pointer;
    fill: #ccc;
  }
  .pan-button:hover {
    fill: #ddd;
  }
</style>
