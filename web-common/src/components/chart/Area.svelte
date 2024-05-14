<script lang="ts">
  import { area, curveLinear } from "d3-shape";
  import { tweened } from "svelte/motion";
  import { interpolatePath } from "d3-interpolate-path";
  import {
    MainAreaColorGradientDark,
    MainAreaColorGradientLight,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import { cubicOut } from "svelte/easing";
  import type { Point } from "./Chart.svelte";

  export let data: Point[];

  const areaFunction = area<Point>()
    .x(({ x }) => x)
    .y0(0)
    .y1(({ y }) => y)
    .curve(curveLinear)
    .defined(({ y }) => y !== null && y !== undefined);

  const tweenedPath = tweened(areaFunction(data), {
    duration: 400,
    interpolate: interpolatePath,
    easing: cubicOut,
  });

  $: tweenedPath.set(areaFunction(data)).catch((e) => console.error(e));
</script>

<path
  vector-effect="non-scaling-stroke"
  d={$tweenedPath}
  fill="url(#gradient-okay)"
/>

<defs>
  <linearGradient id="gradient-okay" x1="0" x2="0" y1="0" y2="1">
    <stop
      offset="5%"
      stop-color={MainAreaColorGradientLight}
      stop-opacity={0.3}
    />
    <stop
      offset="95%"
      stop-color={MainAreaColorGradientDark}
      stop-opacity={0.3}
    />
  </linearGradient>
</defs>
