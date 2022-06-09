<script lang="ts">
  import { tweened } from "svelte/motion";
  import { fly, fade } from "svelte/transition";
  import { cubicOut as easing } from "svelte/easing";
  import { scaleLinear } from "d3-scale";
  import { format } from "d3-format";
  // FIXME: move util to $lib or add a $util
  import { guidGenerator } from "$lib/util/guid";

  interface HistogramBin {
    bucket: number;
    low: number;
    high: number;
    count: number;
  }

  export let data: HistogramBin[];

  export let width = 60;
  export let height = 19;
  export let time = 1000;
  export let fillColor: string; //'hsl(340, 70%, 70%)';
  export let baselineStrokeColor: string;
  export let dataType = "int";

  // s
  export let separate = true;
  $: separateQuantity = separate ? 0.25 : 0;

  // rowsize for table
  export let left = 60;
  export let right = 56;
  export let top = 0;
  export let bottom = 22;

  export let buffer = 4;

  // dots and labels
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
    .range([height - buffer - bottom, top + buffer]);

  $: tw.set(1);

  $: tweeningFunction =
    dataType === "int" ? (v: number) => ~~v : (v: number) => v;

  let formatter: (number) => string;
  $: formatter = dataType === "int" ? format("") : format(".2d");
  $: $lowValue = data[0].low;
  $: $highValue = data.slice(-1)[0].high;

  function transformValue(value, valueType) {
    if (valueType === "mean") {
      return Math.round(value * 10000) / 10000;
    }
    return value;
  }

  let histogramID = guidGenerator();

  // reduce data to construct path for polyline
  $: lineData = data.reduce((pointsPathString, datum) => {
    const {low, high, count} = datum
    const x = X(low) + separateQuantity
    const width = X(high) - X(low) - separateQuantity * 2
    const y = Y(0) * (1 - $tw) + Y(count) * $tw
    const height = Math.min(Y(0), Y(0) * $tw - Y(count) * $tw)

    const currentPoints = `${x},${y+height} ${x},${y} ${x+width},${y}, ${x+width},${y+height} `

    return pointsPathString + currentPoints
  }, "")

</script>
<svg {width} {height}>
  <!-- histogram -->
  <g shape-rendering="crispEdges">
    <polyline
      class={fillColor}
      points={lineData}
    />
    <line
      x1={left + vizOffset}
      x2={width * $tw - right - vizOffset}
      y1={Y(0) + buffer}
      y2={Y(0) + buffer}
      class={baselineStrokeColor}
    />
  </g>
  <slot x={X} y={Y} {buffer} />
</svg>
