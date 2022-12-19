<script lang="ts">
  import { DATA_TYPE_COLORS } from "@rilldata/web-common/lib/duckdb-data-types";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { fly } from "svelte/transition";
  import HistogramBase from "./HistogramBase.svelte";

  export let data;
  export let width;
  export let height = 30;

  let left = 60;
  let right = 4;
  let top = 4;

  export let mean: number;
  export let sd: number;
  export let min: number;
  export let max: number;

  export let anchorBuffer = 8;

  const outlierDeviationThreshold = 6;

  const histogramID = guidGenerator();

  $: effectiveWidth = Math.max(width - 8, 120);

  // add count for histogram base
  $: data = data.map((datum) => ({
    ...datum,
    count: 1,
  }));

  function addDeviationLabels() {
    const intervals: [string, number][] = [["µ", mean]];

    if (max >= mean + outlierDeviationThreshold * sd && sd !== 0) {
      const deviation = (max - mean) / sd;
      const label = `${Math.round(deviation * 100) / 100}σ`;
      intervals.push([label, max]);
    }

    if (min <= mean - outlierDeviationThreshold * sd && sd !== 0) {
      const deviation = (min - mean) / sd;
      const label = `${Math.round(deviation * 100) / 100}σ`;
      intervals.push([label, min]);
    }

    // push labels only if no extreme outliers present
    if (intervals.length == 1) {
      ["1σ", "2σ", "3σ"].forEach((label, i) => {
        // push interval only if it can be displayed on the plot
        if (mean + (i + 1) * sd <= max)
          intervals.push([label, mean + (i + 1) * sd]);
        if (mean - (i + 1) * sd >= min)
          intervals.push(["-" + label, mean - (i + 1) * sd]);
      });
    }
    return intervals;
  }

  $: intervals = addDeviationLabels();
</script>

<HistogramBase
  {data}
  {left}
  {right}
  {top}
  buffer={0}
  width={effectiveWidth}
  separate={false}
  height={height + anchorBuffer * 3}
  bottom={anchorBuffer * 3}
  fillColor={DATA_TYPE_COLORS["DOUBLE"].vizFillClass}
  baselineStrokeColor={DATA_TYPE_COLORS["DOUBLE"].vizStrokeClass}
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
      {#each intervals as [label, value] (label)}
        {@const anchor = value == max ? "end" : "middle"}
        <line
          x1={x(value)}
          x2={x(value)}
          y1={y(0)}
          y2={y(0) + anchorBuffer * 1.2}
          class="stroke-gray-300"
        />
        <text
          in:fly={{ duration: 500, y: y(0) }}
          filter="url(#outline-{histogramID})"
          x={x(value)}
          y={y(0) + anchorBuffer * 2}
          font-size="11"
          fill="hsl(217,1%,40%)"
          text-anchor={anchor}>{label}</text
        >
      {/each}
    </g>
  </svelte:fragment>
</HistogramBase>
