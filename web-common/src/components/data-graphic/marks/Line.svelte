<!-- @component
  Connects points together with a <path> element
    in the order they appear in the data.
-->
<script lang="ts">
  import { extent } from "d3-array";
  import { getContext, onDestroy, onMount } from "svelte";

  import { guidGenerator } from "../../../util/guid";
  import { contexts } from "../constants";
  import type { ExtremumResolutionStore, ScaleStore } from "../state/types";
  import { lineFactory, pathDoesNotDropToZero } from "../utils";

  const markID = guidGenerator();

  export let data;
  export let curve = "curveLinear";
  export let xAccessor = "x";
  export let yAccessor = "y";

  export let color = "hsla(217,70%, 60%, 1)";
  export let lineThickness = undefined;
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

  const config = getContext(contexts.config);

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

  $: totalTravelDistance = data
    .map((di, i) => {
      if (i === data.length - 1) {
        return 0;
      }
      const max = Math.max(
        $yScale(data[i + 1][yAccessor]),
        $yScale(data[i][yAccessor])
      );
      const min = Math.min(
        $yScale(data[i + 1][yAccessor]),
        $yScale(data[i][yAccessor])
      );
      return Math.abs(max - min);
    })
    .reduce((acc, v) => acc + v, 0);

  let computedLineDensity = 0.05;
  let devicePixelRatio = 3;

  $: computedLineDensity = Math.min(
    1,
    /** to determine the stroke width of the path, let's look at
     * the bigger of two values:
     * 1. the "y-ish" distance travelled
     * the inverse of "total travel distance", which is the Y
     * gap size b/t successive points divided by the zoom window size;
     * 2. time series length / available X pixels
     * the time series divided by the total number of pixels in the existing
     * zoom window.
     *
     * These heuristics could be refined, but this seems to provide a reasonable approximation for
     * the stroke width. (1) excels when lots of successive points are close together in the Y direction,
     * whereas (2) excels when a line is very, very noisy (and thus the X direction is the main constraint).
     */
    Math.max(
      2 /
        (totalTravelDistance /
          (($xScale.range()[1] - $xScale.range()[0]) *
            ($config.devicePixelRatio || 3))),
      (($xScale.range()[1] - $xScale.range()[0]) *
        ($config.devicePixelRatio || 3) *
        0.7) /
        data.length /
        1.5
    )
  );

  onMount(() => {
    devicePixelRatio = window.devicePixelRatio;
  });
</script>

{#if lineFcn}
  <path
    d={lineFcn(yAccessor)(data)}
    stroke={color}
    stroke-width={lineThickness ? lineThickness : computedLineDensity}
    fill="none"
    opacity={alpha}
  />
{/if}
