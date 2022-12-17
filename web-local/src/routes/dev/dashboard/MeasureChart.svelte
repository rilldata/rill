<script>
  import Body from "@rilldata/web-local/lib/components/data-graphic/elements/Body.svelte";
  import SimpleDataGraphic from "@rilldata/web-local/lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import Axis from "@rilldata/web-local/lib/components/data-graphic/guides/Axis.svelte";
  import Grid from "@rilldata/web-local/lib/components/data-graphic/guides/Grid.svelte";
  import { extent } from "d3-array";
  import { cubicOut } from "svelte/easing";
  import WithLineChartPath from "./WithLineChartPath.svelte";
  export let xMin;
  export let xMax;
  export let yMin;
  export let yMax;
  export let data;
  export let xAccessor = "ts";
  export let yAccessor = "value";
  export let groundOnZero = true;
  $: [xExtentMin, xExtentMax] = extent(data, (d) => d[xAccessor]);
  $: [yExtentMin, yExtentMax] = extent(data, (d) => d[yAccessor]);
  $: internalXMin = xMin || xExtentMin;
  $: internalXMax = xMax || xExtentMax;
  $: inflate = (yExtentMax - yExtentMin) / yExtentMax;
</script>

<SimpleDataGraphic
  xMin={internalXMin}
  xMax={internalXMax}
  yMin={yMin || yExtentMin * inflate}
  yMax={yMax || yExtentMax / inflate}
  xType="date"
  yType="number"
  width={500}
  height={200}
  right={64}
  yMinTweenProps={{ duration: 1000, easing: cubicOut }}
  yMaxTweenProps={{ duration: 1000, easing: cubicOut }}
  xMaxTweenProps={{ duration: 1000, easing: cubicOut }}
  xMinTweenProps={{ duration: 1000, easing: cubicOut }}
>
  <!-- <ChunkedLine {data} {xAccessor} {yAccessor} /> -->

  <Body>
    <WithLineChartPath {data} {xAccessor} {yAccessor} />
  </Body>

  <Axis side="bottom" />
  <Axis side="right" />
  <Grid />
</SimpleDataGraphic>
