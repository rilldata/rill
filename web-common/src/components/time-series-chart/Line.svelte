<script lang="ts">
  import {
    createLineGenerator,
    createAreaGenerator,
  } from "@rilldata/web-common/components/data-graphic/utils";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import type { ScaleLinear } from "d3-scale";
  import type { ChartDataPoint } from "./types";

  export let data: ChartDataPoint[];
  export let xScale: ScaleLinear<number, number>;
  export let yScale: ScaleLinear<number, number>;
  export let color = MainLineColor;
  export let strokeWidth = 4;
  export let fill: boolean | undefined;

  $: lineFunction = createLineGenerator<ChartDataPoint>({
    x: (d) => xScale(d.index),
    y: (d) => yScale(d.value as number),
    defined: (d) => d.value !== null && d.value !== undefined,
  });

  $: areaFunction = createAreaGenerator<ChartDataPoint>({
    x: (d) => xScale(d.index),
    y0: (d) => yScale(d.value as number),
    y1: yScale.range()[0],
    defined: (d) => d.value !== null && d.value !== undefined,
  });

  $: areaPath = areaFunction(data);

  $: path = lineFunction(data);
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
