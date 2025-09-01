<script lang="ts">
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { ScaleLinear, ScaleTime } from "d3-scale";
  import { area, curveLinear, line } from "d3-shape";

  type DataPoint = {
    date: Date;
    value: number | null | undefined;
  };

  export let data: DataPoint[];
  export let xScale: ScaleTime<number, number>;
  export let yScale: ScaleLinear<number, number>;
  export let color = MainLineColor;
  export let strokeWidth = 4;
  export let fill: boolean | undefined;

  $: curveFunction = curveLinear;

  $: lineFunction = line<DataPoint>()
    .defined(isDefined)
    .x(dateAccessor)
    .y(valueAccessor)
    .curve(curveFunction);

  $: areaFunction = area<DataPoint>()
    .defined(isDefined)
    .x(dateAccessor)
    .y0(valueAccessor)
    .y1(yScale.range()[0])
    .curve(curveFunction);

  $: areaPath = areaFunction(data);

  $: path = lineFunction(data);

  function isDefined(d: DataPoint) {
    return d.value !== null && d.value !== undefined;
  }

  function dateAccessor(d: DataPoint) {
    return xScale(d.date);
  }

  function valueAccessor(d: DataPoint): number {
    // We can safely assert this will be a number because we're using .defined()
    // to filter out null/undefined values before this accessor is called
    return yScale(d.value as number);
  }
</script>

{#if fill}
  <path d={areaPath} fill="url(#chart-gradient)" class="pointer-events-none" />

  <defs>
    <linearGradient id="chart-gradient" x1="0" x2="0" y1="0" y2="1">
      <stop
        offset="5%"
        stop-color={MainAreaColorGradientDark}
        stop-opacity={0.3}
      />
      <stop
        offset="95%"
        stop-color={MainAreaColorGradientLight}
        stop-opacity={0.15}
      />
    </linearGradient>
  </defs>
{/if}

<path
  d={path}
  fill="none"
  stroke={color}
  stroke-width={strokeWidth}
  vector-effect="non-scaling-stroke"
  class="pointer-events-none"
/>
