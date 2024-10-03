<script lang="ts">
  import { line, curveLinear } from "d3-shape";
  import type { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
  import type { ScaleTime, ScaleLinear } from "d3-scale";
  import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";

  export let data: TimeSeriesDatum[];
  export let xKey: string;
  export let yKey: string;
  export let xScaler: ScaleTime<number, number>;
  export let yScaler: ScaleLinear<number, number>;

  $: lineFunction = line<TimeSeriesDatum>()
    .defined((d) => d[yKey] !== null && d[yKey] !== undefined)
    .x((data) => xScaler(data[xKey] as Date))
    .y((data) => yScaler(data[yKey] as number))
    .curve(curveLinear);

  $: path = lineFunction(data);
</script>

<path
  vector-effect="non-scaling-stroke"
  d={path}
  fill="none"
  stroke={MainLineColor}
  stroke-width={1}
/>
