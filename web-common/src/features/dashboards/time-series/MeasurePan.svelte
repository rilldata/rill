<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import type { PlotConfig } from "@rilldata/web-common/components/data-graphic/utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";

  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { getOffset } from "@rilldata/web-common/lib/time/transforms";
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

  $: y1 = $plotConfig.plotTop + $plotConfig.top;
  $: y2 = $plotConfig.plotBottom + $plotConfig.bottom - 10;

  $: midY = (y1 + y2) / 2;

  $: x1 = $plotConfig.plotLeft + $plotConfig.left + 5;
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
    const timeZone = $dashboardStore?.selectedTimezone || "UTC";
    const interval = selectedTimeRange?.interval;
    const allTimeRange = $timeControlsStore?.allTimeRange;

    if (!allTimeRange || !interval || !selectedTimeRange?.start) return;

    const offsetType =
      direction === "left" ? TimeOffsetType.SUBTRACT : TimeOffsetType.ADD;
    const panAmount = TIME_GRAIN[interval].duration;

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
        d="M6.66796 11.9962L15.484 4.11121C15.8061 3.82311 16.3173 4.05174 16.3173 4.48389L16.3173 19.5301C16.3173 19.9626 15.8054 20.1911 15.4835 19.9023L6.66796 11.9962Z"
        class="pan-button"
        on:click|self={() => panCharts("left")}
      />
    </g>
    <g transform={`translate(${x2}, ${midY})`} class="pan-controls">
      <!-- Right Pan Button -->
      <path
        role="presentation"
        d="M17.332 12.0038L8.516 19.8888C8.19389 20.1769 7.68268 19.9483 7.68268 19.5161L7.68268 4.46989C7.68268 4.03741 8.19455 3.80891 8.51651 4.09766L17.332 12.0038Z"
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
