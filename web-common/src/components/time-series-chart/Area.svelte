<script lang="ts">
  import { area, curveLinear } from "d3-shape";
  import type { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import type { ScaleLinear, ScaleTime } from "d3-scale";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";

  export let data: TimeSeriesDatum[];
  export let xKey: string;
  export let yKey: string;
  export let xScaler: ScaleTime<number, number>;
  export let yScaler: ScaleLinear<number, number>;

  $: areaFunction = area<TimeSeriesDatum>()
    .defined((d) => d[yKey] !== null && d[yKey] !== undefined)
    .x((data) => xScaler(data[xKey] as Date))
    .y1(100)
    .y0((data) => yScaler((data[yKey] as number) ?? 100))
    .curve(curveLinear);

  $: path = areaFunction(data);
</script>

<path d={path} fill="url(#chart-gradient)" />

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
      stop-opacity={0.3}
    />
  </linearGradient>
</defs>
