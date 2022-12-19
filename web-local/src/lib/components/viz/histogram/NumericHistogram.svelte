<script lang="ts">
  import { NUMERIC_TOKENS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { fly } from "svelte/transition";
  import HistogramBase from "./HistogramBase.svelte";
  export let data;
  export let width;
  export let height = 100;
  export let color = NUMERIC_TOKENS.vizFillClass; //;

  let left = 60;
  let right = 4;
  let top = 24;

  export let min: number;
  export let qlow: number;
  export let median: number;
  export let qhigh: number;
  export let mean: number;
  export let max: number;

  export let anchorBuffer = 8;
  export let labelOffset = 16;

  let histogramID = guidGenerator();

  $: effectiveWidth = Math.max(width - 8, 120);

  function transformValue(value, valueType) {
    if (valueType === "mean") {
      return Math.round(value * 10000) / 10000;
    }
    return value;
  }

  let fontSize = 12;
  let buffer = 4;
</script>

<HistogramBase
  separate={width > 300}
  bind:buffer
  {top}
  fillColor={NUMERIC_TOKENS.vizFillClass}
  baselineStrokeColor={NUMERIC_TOKENS.vizStrokeClass}
  {data}
  {left}
  {right}
  width={effectiveWidth}
  height={height + 6 * (fontSize + buffer + anchorBuffer) + anchorBuffer}
  bottom={anchorBuffer * 2 + 6 * (fontSize + buffer + anchorBuffer / 2)}
>
  <svelte:fragment let:x let:y let:buffer>
    <filter id="outline-{histogramID}">
      <feMorphology
        in="SourceAlpha"
        result="DILATED"
        operator="dilate"
        radius="1"
      />
      <feFlood flood-color="white" flood-opacity="1" result="PINK" />
      <feComposite in="PINK" in2="DILATED" operator="in" result="OUTLINE" />

      <feMerge>
        <feMergeNode in="OUTLINE" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>
    <g class="textElements">
      <!-- lines first -->
      {#each [["min", min], ["25%", qlow], ["median", median], ["mean", mean], ["75%", qhigh], ["max", max]] as [label, value], i}
        {@const yi =
          y(0) +
          anchorBuffer +
          i * (fontSize + buffer + anchorBuffer / 2) +
          anchorBuffer * 2}

        <line
          x1={left}
          x2={width - right}
          y1={yi - fontSize / 4}
          y2={yi - fontSize / 4}
          stroke-dasharray="2,1"
          class="stroke-gray-300"
        />
        <line
          x1={x(value)}
          x2={x(value)}
          y1={yi - fontSize / 4}
          y2={y(0) + 4}
          class="stroke-gray-300"
        />
      {/each}

      <!-- then everythign else -->
      {#each [["min", min], ["25%", qlow], ["median", median], ["mean", mean], ["75%", qhigh], ["max", max]] as [label, value], i}
        {@const yi =
          y(0) +
          anchorBuffer +
          i * (fontSize + buffer + anchorBuffer / 2) +
          anchorBuffer * 2}
        {@const anchor = x(value) < width / 2 ? "start" : "end"}
        {@const anchorPlacement =
          anchor === "start" ? anchorBuffer : -anchorBuffer}

        <text text-anchor="end" x={left - labelOffset} y={yi}>
          {label}
        </text>
        <text
          filter="url(#outline-{histogramID})"
          x={x(value) + anchorPlacement}
          y={yi}
          font-size="11"
          fill="hsl(217,1%,40%)"
          text-anchor={anchor}>{transformValue(value, label)}</text
        >
        <circle
          in:fly={{ duration: 500, y: -5 }}
          class={color}
          cx={x(value)}
          cy={yi - fontSize / 4}
          r="3"
        />
      {/each}
    </g>
  </svelte:fragment>
</HistogramBase>
