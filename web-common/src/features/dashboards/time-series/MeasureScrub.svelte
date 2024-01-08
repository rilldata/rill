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
  import { createEventDispatcher, getContext } from "svelte";
  import type { Writable } from "svelte/store";

  export let start;
  export let stop;
  export let isScrubbing = false;
  export let showLabels = false;
  export let mouseoverTimeFormat;

  export let labelAccessor;
  export let timeGrainLabel;
  export let data;

  export let isOverStart;
  export let isOverEnd;
  export let isInsideScrub;

  // scrub local control points
  let justCreatedScrub = false;
  let moveStartDelta = 0;
  let moveEndDelta = 0;
  let isResizing: "start" | "end" = undefined;
  let isMovingScrub = false;

  const dispatch = createEventDispatcher();
  const plotConfig: Writable<PlotConfig> = getContext(contexts.config);
  const xScale = getContext(contexts.scale("x")) as ScaleStore;

  const strokeWidth = 1;
  const xLabelBuffer = 8;
  const yLabelBuffer = 10;
  $: y1 = $plotConfig.plotTop + $plotConfig.top + 5;
  $: y2 = $plotConfig.plotBottom - $plotConfig.bottom - 1;

  $: hasSubrangeSelected = Boolean(start && stop);

  export let cursorClass = "";
  $: cursorClass = isMovingScrub
    ? "cursor-grabbing"
    : isInsideScrub
      ? "cursor-grab"
      : isScrubbing || isOverStart || isOverEnd
        ? "cursor-ew-resize"
        : "";

  export let preventScrubReset;
  $: preventScrubReset = justCreatedScrub || isScrubbing || isResizing;

  export function startScrub(event) {
    if (hasSubrangeSelected) {
      const startX = event.detail?.start?.x;
      // check if we are scrubbing on the edges of scrub rect
      if (isOverStart || isOverEnd) {
        isResizing = isOverStart ? "start" : "end";
        dispatch("update", {
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

  export function moveScrub(event) {
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

        dispatch("update", {
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

        const insideBounds = $xScale(newStart) >= 0 && $xScale(newEnd) >= 0;
        if (insideBounds && newStart?.getTime() !== start?.getTime()) {
          dispatch("update", {
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
        dispatch("update", {
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
      dispatch("reset");
      return;
    }

    // Remove scrub if start and end are same
    if (hasSubrangeSelected && start?.getTime() === stop?.getTime()) {
      dispatch("reset");
      return;
    }

    isResizing = undefined;
    isMovingScrub = false;
    justCreatedScrub = true;

    // reset justCreatedScrub after 100 milliseconds
    setTimeout(() => {
      justCreatedScrub = false;
    }, 100);

    dispatch("update", {
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
    {@const xStart = xScale(Math.min(start, stop))}
    {@const xEnd = xScale(Math.max(start, stop))}
    <g>
      {#if showLabels}
        <text text-anchor="end" x={xStart - xLabelBuffer} y={y1 + yLabelBuffer}>
          {mouseoverTimeFormat(Math.min(start, stop))}
        </text>
        <circle
          cx={xStart}
          cy={y1}
          r={3}
          paint-order="stroke"
          class="fill-blue-700"
          stroke="white"
          stroke-width="3"
        />
        <text text-anchor="start" x={xEnd + xLabelBuffer} y={y1 + yLabelBuffer}>
          {mouseoverTimeFormat(Math.max(start, stop))}
        </text>
        <circle
          cx={xEnd}
          cy={y1}
          r={3}
          paint-order="stroke"
          class="fill-blue-700"
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
