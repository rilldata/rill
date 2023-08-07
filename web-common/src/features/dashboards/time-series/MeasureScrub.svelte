<script lang="ts">
  import WithGraphicContexts from "@rilldata/web-common/components/data-graphic/functional-components/WithGraphicContexts.svelte";
  import { createEventDispatcher, getContext } from "svelte";
  import type { PlotConfig } from "@rilldata/web-common/components/data-graphic/utils";
  import type { Writable } from "svelte/store";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";

  export let start;
  export let stop;
  export let isScrubbing = false;
  export let showLabels = false;
  export let mouseoverTimeFormat;

  const dispatch = createEventDispatcher();
  const plotConfig: Writable<PlotConfig> = getContext(contexts.config);

  const strokeWidth = 1;
  const xLabelBuffer = 8;
  const yLabelBuffer = 10;
  const y1 = $plotConfig.plotTop + $plotConfig.top + 5;
  const y2 = $plotConfig.plotBottom - $plotConfig.bottom - 1;

  let showContextMenu = false;
  let contextMenuOpen = false;
  function onContextMenu() {
    showContextMenu = true;
  }

  function onKeyDown(e) {
    // if key Z is pressed, zoom the scrub
    if (e.key === "z") {
      dispatch("zoom");
    }
    if (!isScrubbing && e.key === "Escape") {
      dispatch("reset");
    }
  }
</script>

{#if start && stop}
  <WithGraphicContexts let:xScale let:yScale>
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
        stroke="#60A5FA"
        stroke-width={strokeWidth}
      />
      <line
        x1={xEnd}
        x2={xEnd}
        {y1}
        {y2}
        stroke="#60A5FA"
        stroke-width={strokeWidth}
      />
    </g>
    <g opacity={isScrubbing ? "0.4" : "0.2"}>
      <rect
        class:rect-shadow={isScrubbing}
        x={Math.min(xStart, xEnd)}
        y={y1}
        width={Math.abs(xStart - xEnd)}
        height={y2 - y1}
        fill="url('#scrubbing-gradient')"
      />
      <foreignObject
        x={Math.min(xStart, xEnd) + 20}
        y={y1 + 20}
        width="300"
        height="160"
      >
        <div on:contextmenu|preventDefault={() => onContextMenu()}>
          <!-- FIX ME: Unable to add menu on top of SVG  -->
          {#if showContextMenu}
            <!-- context menu -->
            <WithTogglableFloatingElement
              location="right"
              alignment="start"
              distance={16}
              let:toggleFloatingElement
              bind:active={contextMenuOpen}
            >
              <Menu
                maxWidth="300px"
                on:click-outside={toggleFloatingElement}
                on:escape={toggleFloatingElement}
                on:item-select={toggleFloatingElement}
                slot="floating-element"
              >
                <MenuItem on:select={() => console.log("zoom")}
                  >Zoom to subrange</MenuItem
                >
              </Menu>
            </WithTogglableFloatingElement>
          {/if}
        </div>
      </foreignObject>
    </g>
  </WithGraphicContexts>
{/if}

<svelte:window on:keydown|preventDefault={onKeyDown} />

<defs>
  <linearGradient id="scrubbing-gradient" gradientUnits="userSpaceOnUse">
    <stop stop-color="#558AFF" />
    <stop offset="0.36" stop-color="#4881FF" />
    <stop offset="1" stop-color="#2563EB" />
  </linearGradient>
</defs>

<style>
  .rect-shadow {
    filter: drop-shadow(0px 4px 6px rgba(0, 0, 0, 0.1))
      drop-shadow(0px 10px 15px rgba(0, 0, 0, 0.2));
  }

  g {
    transition: opacity ease 0.4s;
  }
</style>
