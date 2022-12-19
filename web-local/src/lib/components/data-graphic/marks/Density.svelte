<script lang="ts">
  import { Area, Line } from "$lib/components/data-graphic/marks";
  import { extent } from "d3-array";
  import { density1d } from "fast-kde";
  import { tweened } from "svelte/motion";

  export let data;
  export let xMin: number = undefined;
  export let xMax: number = undefined;
  export let yMin = undefined;
  export let yMax = undefined;

  export let bandwidth = 0.1;

  export let xAccessor: string;
  export let yAccessor: string;
  export let bandwidthTweenProps = { duration: 50 };
  export let lineThickness = 1;
  export let transform = true;
  export let area = true;
  export let lineColor = "hsla(1, 80%, 30%, 1)";
  export let areaColor = "hsla(1, 80%, 90%, 1)";

  let bandwidthTween = tweened(bandwidth, bandwidthTweenProps);
  $: bandwidthTween.set(bandwidth);

  $: computedRange = extent(data, (d) => d[xAccessor]);

  $: innerXMin = xMin === undefined ? computedRange[0] : xMin;
  $: innerXMax = xMax === undefined ? computedRange[1] : xMax;
  $: if (innerXMin === innerXMax) {
    innerXMin = innerXMin * 0.8;
    innerXMax = innerXMax * 1.2;
  }

  $: kde = Array.from(
    density1d(data, {
      weight: (d) => d[yAccessor],
      x: (d) => d[xAccessor],
      bandwidth: $bandwidthTween * ((innerXMax - innerXMin) / 12),
      bins: 512,
      extent: [innerXMin, innerXMax],
    })
  );
  $: densityGrid = transform ? (data ? Array.from(kde) : []) : data;
</script>

{#if densityGrid}
  {#if area}
    <Area
      data={densityGrid}
      xAccessor={transform ? "x" : xAccessor}
      yAccessor={transform ? "y" : yAccessor}
      color={areaColor}
      {yMin}
      {yMax}
    />
  {/if}
  <Line
    data={densityGrid}
    {lineThickness}
    xAccessor={transform ? "x" : xAccessor}
    yAccessor={transform ? "y" : yAccessor}
    color={lineColor}
    {yMin}
    {yMax}
  />
{/if}
