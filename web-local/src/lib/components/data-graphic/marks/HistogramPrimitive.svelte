<!-- @component
This primitive component plots histogram data from duckdb as a single path.
It's not meant to be a general-purpose bar mark / geom, nor should you expect
it to do any automatic binning of data, which is done server-side.
-->
<script lang="ts">
  import { guidGenerator } from "$lib/util/guid";
  import { extent, max, min } from "d3-array";
  import { getContext, onDestroy } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { contexts } from "../constants";
  import type { ExtremumResolutionStore, ScaleStore } from "../state/types";
  import { barplotPolyline } from "../utils";

  const markID = guidGenerator();

  export let data;
  export let xLowAccessor;
  export let xHighAccessor;
  export let yAccessor;
  export let inferXExtent = true;
  export let inferYExtent = true;
  export let xMin: number = undefined;
  export let xMax: number = undefined;
  export let yMin = undefined;
  export let yMax = undefined;
  export let lineThickness = 1;

  export let separator = 0.5;
  export let closeBottom = false;

  export let outlineColor = "hsla(1,90%, 60%, .7)";
  export let color = "hsla(1,70%, 80%, .5)";
  export let stopOpacity = 0.4;

  const xMinStore = getContext(contexts.min("x")) as ExtremumResolutionStore;
  const xMaxStore = getContext(contexts.max("x")) as ExtremumResolutionStore;
  const yMinStore = getContext(contexts.min("y")) as ExtremumResolutionStore;
  const yMaxStore = getContext(contexts.max("y")) as ExtremumResolutionStore;

  // get extents
  $: xMinValue = min(data, (d) => d[xLowAccessor]);
  $: xMaxValue = max(data, (d) => d[xHighAccessor]);
  $: [yMinValue, yMaxValue] = extent(data, (d) => d[yAccessor]);
  // set your extrema here
  $: if (inferXExtent) xMinStore.setWithKey(markID, xMin || xMinValue);
  $: if (inferXExtent) xMaxStore.setWithKey(markID, xMax || xMaxValue);

  $: if (inferYExtent) yMinStore.setWithKey(markID, yMin || 0);
  $: if (inferYExtent) yMaxStore.setWithKey(markID, yMax || yMaxValue);
  // we should set the extrema here.

  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  onDestroy(() => {
    if (inferXExtent) xMinStore.removeKey(markID);
    if (inferXExtent) xMaxStore.removeKey(markID);
    if (inferYExtent) yMinStore.removeKey(markID);
    if (inferYExtent) yMaxStore.removeKey(markID);
  });

  const inflator = tweened(0, { duration: 800, easing: cubicOut });
  $: inflator.set(1);

  $: d = barplotPolyline(
    data,
    xLowAccessor,
    xHighAccessor,
    yAccessor,
    $xScale,
    $yScale,
    separator,
    closeBottom,
    $inflator
  );
</script>

{#if d?.length && $xScale && $yScale}
  <defs>
    <linearGradient id="gradient-{markID}" x1="0" x2="0" y1="0" y2="1">
      <stop offset="5%" stop-color={color} />
      <stop offset="95%" stop-color={color} stop-opacity={stopOpacity} />
    </linearGradient>
  </defs>
  <path {d} fill="url(#gradient-{markID})" />
  <path {d} stroke={outlineColor} fill="none" stroke-width={lineThickness} />
{/if}
