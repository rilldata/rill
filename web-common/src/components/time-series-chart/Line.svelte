<script lang="ts">
  import { line, curveLinear, area } from "d3-shape";
  import type { ScaleTime, ScaleLinear } from "d3-scale";
  import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { V1TimeSeriesValue } from "@rilldata/web-common/runtime-client";
  import type { Interval } from "luxon";

  type DataPoint = {
    interval: Interval<true>;
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

  $: areaFunction = area<V1TimeSeriesValue>()
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
    return xScale(d.interval.start.toJSDate());
  }

  function valueAccessor(d: DataPoint) {
    return d?.value !== null && d?.value !== undefined ? yScale(d.value) : null;
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
