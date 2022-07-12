<!-- @component
  Connects points together with a <path> element
    in the order they appear in the data.
-->
<script lang="ts">
  import { getContext, onDestroy } from "svelte";
  import { extent } from "d3-array";

  import {
    lineFactory,
    pathDoesNotDropToZero,
  } from "$lib/components/data-graphic/utils";
  import { guidGenerator } from "$lib/util/guid";
  import { contexts } from "../constants";
  import type { ExtremumResolutionStore, ScaleStore } from "../state/types";
  import { tweened } from "svelte/motion";

  const markID = guidGenerator();

  export let data;
  export let curve = "curveLinear";
  export let xAccessor = "x";
  export let yAccessor = "y";

  export let color = "hsla(217,70%, 60%, 1)";
  export let lineThickness = 1;
  export let alpha = 1;
  export let pathDefined = pathDoesNotDropToZero;

  export let xMin = undefined;
  export let xMax = undefined;
  export let yMin = undefined;
  export let yMax = undefined;

  const xMinStore = getContext(contexts.min("x")) as ExtremumResolutionStore;
  const xMaxStore = getContext(contexts.max("x")) as ExtremumResolutionStore;
  const yMinStore = getContext(contexts.min("y")) as ExtremumResolutionStore;
  const yMaxStore = getContext(contexts.max("y")) as ExtremumResolutionStore;

  // get extents
  $: [xMinValue, xMaxValue] = extent(data, (d) => d[xAccessor]);
  $: [yMinValue, yMaxValue] = extent(data, (d) => d[yAccessor]);
  // set your extrema here
  $: xMinStore.setWithKey(markID, xMin || xMinValue);
  $: xMaxStore.setWithKey(markID, xMax || xMaxValue);

  $: yMinStore.setWithKey(markID, yMin || yMinValue);
  $: yMaxStore.setWithKey(markID, yMax || yMaxValue);
  // we should set the extrema here.

  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  onDestroy(() => {
    xMinStore.removeKey(markID);
    xMaxStore.removeKey(markID);
    yMinStore.removeKey(markID);
    yMaxStore.removeKey(markID);
  });

  let lineFcn;
  $: if ($xScale && $yScale) {
    lineFcn = lineFactory({
      xScale: $xScale,
      yScale: $yScale,
      curve,
      xAccessor,
      pathDefined,
    });
  }
</script>

{#if lineFcn}
  <path
    d={lineFcn(yAccessor)(data)}
    stroke={color}
    stroke-width={lineThickness}
    fill="none"
    opacity={alpha}
  />
{/if}
