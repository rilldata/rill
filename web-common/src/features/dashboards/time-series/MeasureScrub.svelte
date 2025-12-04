<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import type { PlotConfig } from "@rilldata/web-common/components/data-graphic/utils";
  import {
    ScrubArea0Color,
    ScrubArea1Color,
    ScrubArea2Color,
    ScrubBoxColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { getBisectedTimeFromCordinates } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import type { TimeSeriesDatum } from "./timeseries-data-store";
  import type { DateTimeUnit } from "luxon";

  export let start: Date | null;
  export let stop: Date | null;
  export let isScrubbing = false;
  export let showLabels = false;
  export let labelAccessor: string;
  export let timeGrainLabel: DateTimeUnit;
  export let data: TimeSeriesDatum[];
  export let isOverStart: boolean;
  export let isOverEnd: boolean;
  export let isInsideScrub: boolean;
  export let mouseoverTimeFormat: (d: unknown) => string;
  export let onUpdate: (data: {
    start: Date | null;
    stop: Date | null;
    isScrubbing: boolean;
  }) => void;
  export let onReset: () => void;

  // scrub local control points
  let justCreatedScrub = false;
  let moveStartDelta = 0;
  let moveEndDelta = 0;
  let isResizing: "start" | "end" | undefined = undefined;
  let isMovingScrub = false;

  const plotConfig: Writable<PlotConfig> = getContext(contexts.config);
  const xScale = getContext<ScaleStore>(contexts.scale("x"));

  const strokeWidth = 1;
  const xLabelBuffer = 8;
  const yLabelBuffer = 10;
  $: y1 = $plotConfig.plotTop + $plotConfig.top + 5;
  $: y2 = $plotConfig.plotBottom - $plotConfig.bottom - 1;

  $: hasSubrangeSelected = Boolean(start && stop);

  export let cursorClass = "cursor-pointer";
  $: cursorClass = isMovingScrub
    ? "cursor-grabbing"
    : isInsideScrub
      ? "cursor-grab"
      : isScrubbing || isOverStart || isOverEnd
        ? "cursor-ew-resize"
        : "cursor-pointer";

  export let preventScrubReset: boolean;
  $: preventScrubReset = justCreatedScrub || isScrubbing || Boolean(isResizing);

  export function startScrub(event: CustomEvent<{ start: { x: number } }>) {
    if (!start || !stop) return;

    if (hasSubrangeSelected) {
      const startX = event.detail?.start?.x;
      // check if we are scrubbing on the edges of scrub rect
      if (isOverStart || isOverEnd) {
        isResizing = isOverStart ? "start" : "end";
        onUpdate({
          start: start,
          stop: stop,
          isScrubbing: true,
        });

        return;
      } else if (isInsideScrub) {
        isMovingScrub = true;
        moveStartDelta = startX - $xScale(start);
        moveEndDelta = startX - $xScale(stop);

        return;
      }
    }
  }

  export function moveScrub(
    event: CustomEvent<{ start: { x: number }; stop: { x: number } }>,
  ) {
    const startX = event.detail?.start?.x;
    const scrubStartDate = getBisectedTimeFromCordinates(
      startX,
      $xScale,
      labelAccessor,
      data,
      timeGrainLabel,
    );

    let stopX = event.detail?.stop?.x;
    let intermediateScrubVal = getBisectedTimeFromCordinates(
      stopX,
      $xScale,
      labelAccessor,
      data,
      timeGrainLabel,
    );

    if (hasSubrangeSelected && (isResizing || isMovingScrub)) {
      if (
        isResizing &&
        intermediateScrubVal?.getTime() !== stop?.getTime() &&
        intermediateScrubVal?.getTime() !== start?.getTime()
      ) {
        /**
         * Adjust the ends of the subrange by dragging either end.
         * This snaps to the nearest time grain.
         */
        const newStart = isResizing === "start" ? intermediateScrubVal : start;
        const newEnd = isResizing === "end" ? intermediateScrubVal : stop;

        onUpdate({
          start: newStart,
          stop: newEnd,
          isScrubbing: true,
        });
      } else if (!isResizing && isMovingScrub) {
        /**
         * Pick up and shift the entire subrange left/right
         * This snaps to the nearest time grain
         */

        const startX = event.detail?.start?.x;
        const delta = stopX - startX;

        const newStart = getBisectedTimeFromCordinates(
          startX - moveStartDelta + delta,
          $xScale,
          labelAccessor,
          data,
          timeGrainLabel,
        );

        const newEnd = getBisectedTimeFromCordinates(
          startX - moveEndDelta + delta,
          $xScale,
          labelAccessor,
          data,
          timeGrainLabel,
        );

        if (!newStart || !newEnd) return;

        const insideBounds = $xScale(newStart) >= 0 && $xScale(newEnd) >= 0;
        if (insideBounds && newStart?.getTime() !== start?.getTime()) {
          onUpdate({
            start: newStart,
            stop: newEnd,
            isScrubbing: true,
          });
        }
      }
    } else {
      // Only make state changes when the bisected value changes
      if (
        scrubStartDate?.getTime() !== start?.getTime() ||
        intermediateScrubVal?.getTime() !== stop?.getTime()
      ) {
        onUpdate({
          start: scrubStartDate,
          stop: intermediateScrubVal,
          isScrubbing: true,
        });
      }
    }
  }

  export function endScrub() {
    // if the mouse leaves the svg area, reset the scrub
    // check if any parent of explicitOriginalTarget is a svg or not
    const hoverElem = Array.from(document.querySelectorAll(":hover")).pop();
    if (hoverElem?.nodeName !== "svg" && !hoverElem?.closest("svg")) {
      onReset();
      return;
    }

    // Remove scrub if start and end are same
    if (hasSubrangeSelected && start?.getTime() === stop?.getTime()) {
      onReset();
      return;
    }

    isResizing = undefined;
    isMovingScrub = false;
    justCreatedScrub = true;

    // reset justCreatedScrub after 100 milliseconds
    setTimeout(() => {
      justCreatedScrub = false;
    }, 100);

    onUpdate({
      start,
      stop,
      isScrubbing: false,
    });
  }

  /***
   * prevent unwanted scrub changes when clicked
   * inside a scrub without any cursor move
   */
  function onMouseUp() {
    isResizing = undefined;
    isMovingScrub = false;
    moveStartDelta = 0;
    moveEndDelta = 0;
  }
</script>

{#if start && stop}
  <WithGraphicContexts let:xScale>
    {@const numStart = Number(start)}
    {@const numStop = Number(stop)}
    {@const xStart = xScale(Math.min(numStart, numStop))}
    {@const xEnd = xScale(Math.max(numStart, numStop))}
    <g>
      {#if showLabels}
        <text text-anchor="end" x={xStart - xLabelBuffer} y={y1 + yLabelBuffer}>
          {mouseoverTimeFormat(Math.min(numStart, numStop))}
        </text>
        <circle
          cx={xStart}
          cy={y1}
          r={3}
          paint-order="stroke"
          class="fill-primary-700"
          stroke="white"
          stroke-width="3"
        />
        <text text-anchor="start" x={xEnd + xLabelBuffer} y={y1 + yLabelBuffer}>
          {mouseoverTimeFormat(Math.max(numStart, numStop))}
        </text>
        <circle
          cx={xEnd}
          cy={y1}
          r={3}
          paint-order="stroke"
          class="fill-primary-700"
          stroke="white"
          stroke-width="3"
        />
      {/if}
      <line
        x1={xStart}
        x2={xStart}
        {y1}
        {y2}
        stroke={ScrubBoxColor}
        stroke-width={strokeWidth}
      />
      <line
        x1={xEnd}
        x2={xEnd}
        {y1}
        {y2}
        stroke={ScrubBoxColor}
        stroke-width={strokeWidth}
      />
    </g>
    <g
      role="presentation"
      opacity={isScrubbing ? "0.4" : "0.2"}
      on:mouseup={() => onMouseUp()}
    >
      <rect
        class:rect-shadow={isScrubbing}
        x={Math.min(xStart, xEnd)}
        y={y1}
        width={Math.abs(xStart - xEnd)}
        height={y2 - y1}
        fill="url('#scrubbing-gradient')"
      />
    </g>
  </WithGraphicContexts>
{/if}

<defs>
  <linearGradient gradientUnits="userSpaceOnUse" id="scrubbing-gradient">
    <stop stop-color={ScrubArea0Color} />
    <stop offset="0.36" stop-color={ScrubArea1Color} />
    <stop offset="1" stop-color={ScrubArea2Color} />
  </linearGradient>
</defs>

<style>
  .rect-shadow {
    filter: drop-shadow(0px 4px 6px rgba(0, 0, 0, 0.1))
      drop-shadow(0px 10px 15px rgba(0, 0, 0, 0.2));
  }

  g {
    transition: opacity ease 0.3s;
  }
</style>
