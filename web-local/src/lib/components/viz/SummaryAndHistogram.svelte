<script lang="ts">
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { format } from "d3-format";
  import { scaleLinear } from "d3-scale";
  import { cubicOut as easing } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { fly } from "svelte/transition";

  interface HistogramBin {
    bucket: number;
    low: number;
    high: number;
    count: number;
  }

  export let min: number;
  export let qlow: number;
  export let median: number;
  export let qhigh: number;
  export let mean: number;
  export let max: number;

  export let data: HistogramBin[];

  export let width = 60;
  export let height = 19;
  export let time = 1000;
  export let color = "hsl(340, 70%, 70%)";
  export let dataType = "int";

  // rowsize for table
  export let left = 60;
  export let right = 56;
  export let fontSize = 20;
  export let bottom = 22;

  // dots and labels
  export let anchorBuffer = 8;
  export let labelOffset = 16;
  export let vizOffset = 0;

  // what do we have here? min, q25, q50, mean, q75, max

  const tw = tweened(0, { duration: time, easing });

  const lowValue = tweened(0, { duration: time / 2, easing });
  const highValue = tweened(0, { duration: time / 2, easing });

  $: minX = Math.min(...data.map((d) => d.low));
  $: maxX = Math.max(...data.map((d) => d.high));
  $: X = scaleLinear()
    .domain([minX, maxX])
    .range([left + vizOffset, width - right - vizOffset]);

  $: yVals = data.map((d) => d.count);
  $: maxY = Math.max(...yVals);
  $: Y = scaleLinear()
    .domain([0, maxY])
    .range([height - 4 - bottom, 4]);

  $: tw.set(1);

  $: tweeningFunction =
    dataType === "int" ? (v: number) => ~~v : (v: number) => v;

  let formatter: (number) => string;
  $: formatter = dataType === "int" ? format("") : format(".2d");
  $: $lowValue = data[0].low;
  $: $highValue = data.slice(-1)[0].high;
  $: formattedLowValue = formatter(tweeningFunction($lowValue));
  $: formattedHighValue = formatter(tweeningFunction($highValue));

  function transformValue(value, valueType) {
    if (valueType === "mean") {
      return Math.round(value * 10000) / 10000;
    }
    return value;
  }

  let histogramID = guidGenerator();
</script>

<svg {width} height={height + 6 * fontSize}>
  <!-- text outline filter -->
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
  <!-- histogram -->
  <g shape-rendering="crispEdges">
    {#each data as { low, high, count }, i}
      {@const x = X(low)}
      {@const width = X(high) - X(low)}
      {@const y = Y(0) * (1 - $tw) + Y(count) * $tw}
      {@const height = Math.min(Y(0), Y(0) * $tw - Y(count) * $tw)}

      <rect {x} {width} {y} {height} fill={color} />
    {/each}
    <line
      x1={X(X.domain()[0])}
      x2={width * $tw - right - vizOffset}
      y1={Y(0) + 4}
      y2={Y(0) + 4}
      stroke={color}
    />
  </g>

  <g class="lineElements">
    {#each [["min", min], ["q25", qlow], ["med", median], ["mean", mean], ["q75", qhigh], ["max", max]] as [label, value], i}
      {@const y = height + i * fontSize}
      <line
        x1={left}
        x2={width - right}
        y1={y - fontSize / 4}
        y2={y - fontSize / 4}
        stroke-dasharray="2,1"
        opacity=".3"
        stroke={color}
      />
      <line
        x1={X(value)}
        x2={X(value)}
        y1={y - fontSize / 4}
        y2={Y(0) + 4}
        opacity=".3"
        stroke={color}
      />
    {/each}
  </g>
  <g class="textElements">
    {#each [["min", min], ["25%", qlow], ["median", median], ["mean", mean], ["75%", qhigh], ["max", max]] as [label, value], i}
      {@const y = height + i * fontSize}
      {@const anchor = X(value) < width / 2 ? "start" : "end"}
      {@const anchorPlacement =
        anchor === "start" ? anchorBuffer : -anchorBuffer}
      <text text-anchor="end" x={left - labelOffset} {y}>
        {label}
      </text>
      <text
        filter="url(#outline-{histogramID})"
        x={X(value) + anchorPlacement}
        {y}
        font-size="11"
        fill="hsl(217,1%,40%)"
        text-anchor={anchor}>{transformValue(value, label)}</text
      >
      <circle
        in:fly={{ duration: 500, y: -5 }}
        fill={color}
        cx={X(value)}
        cy={y - fontSize / 4}
        r="3"
      />
    {/each}
  </g>
</svg>

<!-- temp: get json -->

<button
  on:click={() => {
    navigator.clipboard.writeText(JSON.stringify(data, null, 2));
    console.log("copied to clipboard.");
  }}>json</button
>
