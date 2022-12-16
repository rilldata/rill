<script>
  import SimpleDataGraphic from "@rilldata/web-local/lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import WithTween from "@rilldata/web-local/lib/components/data-graphic/functional-components/WithTween.svelte";
  import { extent } from "d3-array";
  import { interpolatePath, pathCommandsFromString } from "d3-interpolate-path";
  import { cubicOut } from "svelte/easing";
  import ChunkedLine from "./ChunkedLine.svelte";
  import WithLineChartPath from "./WithLineChartPath.svelte";
  export let xMin;
  export let xMax;
  export let data;
  export let xAccessor = "ts";
  export let yAccessor = "value";
  export let groundOnZero = true;
  $: [xExtentMin, xExtentMax] = extent(data, (d) => d[xAccessor]);
  $: [yExtentMin, yExtentMax] = extent(data, (d) => d[yAccessor]);
  $: internalXMin = xMin || xExtentMin;
  $: internalXMax = xMax || xExtentMax;
</script>

{yExtentMin} - {yExtentMax}
<SimpleDataGraphic
  xMin={internalXMin}
  xMax={internalXMax}
  yMin={yExtentMin > 0 || groundOnZero ? 0 : yExtentMin}
  yMax={yExtentMax}
  xType="date"
  yType="number"
  width={500}
  height={500}
>
  <ChunkedLine {data} {xAccessor} {yAccessor} />
  <WithLineChartPath {data} {xAccessor} {yAccessor} />
  {#if false}
    <WithLineChartPath {data} {xAccessor} {yAccessor} let:d>
      <WithTween
        value={d}
        tweenProps={{
          duration: 1000,
          interpolate: (a, b) =>
            interpolatePath(a, b, (ai, bi) => {
              return !(ai.type !== "M" && bi.type === "M");
            }),
          easing: cubicOut,
        }}
        let:output={dt}
      >
        {@const commands = pathCommandsFromString(dt)}
        {#each commands as command}
          <text
            x={command.x}
            y={command.y - 3}
            style:font-weight={command.type === "M" ? "bold" : "normal"}
            r={5}>{command.type}</text
          >
        {/each}
        <path d={dt} stroke="black" fill="none" />
      </WithTween>
    </WithLineChartPath>
  {/if}
</SimpleDataGraphic>
