<!-- @component
  Draws an "area under the curve" shape as a <path>
    in the order the points appear in the data.
-->
<script lang="ts">
  import { getContext, onDestroy } from "svelte";
  import { extent } from "d3-array";

  import { areaFactory } from "$lib/components/data-graphic/utils";
  import { guidGenerator } from "$lib/util/guid";
  import type { ExtremumResolutionStore, ScaleStore } from "../state/types";
  import { contexts } from "../constants";

  const markID = guidGenerator();

  export let data;
  export let curve = "curveLinear";
  export let xAccessor = "x";
  export let yAccessor = "y";

  export let color = "hsla(217,70%, 80%, .4)";
  export let alpha = 1;

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

  const xScale = getContext("rill:data-graphic:x-scale") as ScaleStore;
  const yScale = getContext("rill:data-graphic:y-scale") as ScaleStore;

  onDestroy(() => {
    xMinStore.removeKey(markID);
    xMaxStore.removeKey(markID);
    yMinStore.removeKey(markID);
    yMaxStore.removeKey(markID);
  });

  let areaFcn;
  $: if ($xScale) {
    areaFcn = areaFactory({
      xScale: $xScale,
      yScale: $yScale,
      curve,
      xAccessor,
    });
  }
</script>

{#if areaFcn}
  <defs>
    <linearGradient id="gradient-{markID}" x1="0" x2="0" y1="0" y2="1">
      <stop offset="5%" stop-color={color} />
      <stop offset="95%" stop-color={color} stop-opacity={0.3} />
    </linearGradient>
  </defs>
  <path
    d={areaFcn(yAccessor)(data)}
    fill="url(#gradient-{markID})"
    opacity={alpha}
  />
{/if}
