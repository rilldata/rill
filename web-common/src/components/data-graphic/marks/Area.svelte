<!-- @component
  Draws an "area under the curve" shape as a <path>
    in the order the points appear in the data.
-->
<script lang="ts">
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { extent } from "d3-array";
  import { getContext, onDestroy } from "svelte";
  import { contexts } from "../constants";
  import type { ExtremumResolutionStore, ScaleStore } from "../state/types";
  import { areaFactory } from "../utils";

  const markID = guidGenerator();

  const xMinStore = getContext<ExtremumResolutionStore>(contexts.min("x"));
  const xMaxStore = getContext<ExtremumResolutionStore>(contexts.max("x"));
  const yMinStore = getContext<ExtremumResolutionStore>(contexts.min("y"));
  const yMaxStore = getContext<ExtremumResolutionStore>(contexts.max("y"));

  const xScale = getContext<ScaleStore>("rill:data-graphic:x-scale");
  const yScale = getContext<ScaleStore>("rill:data-graphic:y-scale");

  export let data: Record<string, number | Date>[];
  export let curve = "curveLinear";
  export let xAccessor = "x";
  export let yAccessor = "y";

  export let color = "hsla(217,70%, 80%, .4)";
  export let alpha = 1;
  export let stopOpacity = 0.3;

  export let xMin: number | Date | undefined = undefined;
  export let xMax: number | Date | undefined = undefined;
  export let yMin: number | Date | undefined = undefined;
  export let yMax: number | Date | undefined = undefined;

  $: [xMinValue, xMaxValue] = extent(data, (d) => d[xAccessor]);
  $: [yMinValue, yMaxValue] = extent(data, (d) => d[yAccessor]);

  $: finalXMin = xMin || xMinValue;
  $: finalXMax = xMax || xMaxValue;
  $: finalYMin = yMin || yMinValue;
  $: finalYMax = yMax || yMaxValue;

  $: if (finalXMin) xMinStore.setWithKey(markID, finalXMin);
  $: if (finalXMax) xMaxStore.setWithKey(markID, finalXMax);
  $: if (finalYMin) yMinStore.setWithKey(markID, finalYMin);
  $: if (finalYMax) yMaxStore.setWithKey(markID, finalYMax);

  $: areaFcn =
    $xScale &&
    areaFactory({
      xScale: $xScale,
      yScale: $yScale,
      curve,
      xAccessor,
    });

  onDestroy(() => {
    xMinStore.removeKey(markID);
    xMaxStore.removeKey(markID);
    yMinStore.removeKey(markID);
    yMaxStore.removeKey(markID);
  });
</script>

{#if areaFcn}
  <defs>
    <linearGradient id="gradient-{markID}" x1="0" x2="0" y1="0" y2="1">
      <stop offset="5%" stop-color={color} />
      <stop offset="95%" stop-color={color} stop-opacity={stopOpacity} />
    </linearGradient>
  </defs>
  <path
    d={areaFcn(yAccessor)(data)}
    fill="url(#gradient-{markID})"
    opacity={alpha}
  />
{/if}
