<script lang="ts">
  import { getContext, onDestroy } from "svelte";
  import { extent } from "d3-array";

  import { lineFactory } from "$lib/components/data-graphic/utils";
  import { guidGenerator } from "$lib/util/guid";

  const markID = guidGenerator();

  export let data;
  export let curve = "curveLinear";
  export let xAccessor = "x";
  export let yAccessor = "y";

  export let color = "black";
  export let lineThickness = 1;
  export let alpha = 1;

  export let xMin = undefined;
  export let xMax = undefined;
  export let yMin = undefined;
  export let yMax = undefined;

  const xMinStore = getContext("rill:data-graphic:x-min");
  const xMaxStore = getContext("rill:data-graphic:x-max");
  const yMinStore = getContext("rill:data-graphic:y-min");
  const yMaxStore = getContext("rill:data-graphic:y-max");

  // get extents
  $: [xMinValue, xMaxValue] = extent(data, (d) => d[xAccessor]);
  $: [yMinValue, yMaxValue] = extent(data, (d) => d[yAccessor]);
  // set your extrema here
  $: xMinStore.setWithKey(markID, xMin || xMinValue);
  $: xMaxStore.setWithKey(markID, xMax || xMaxValue);

  $: yMinStore.setWithKey(markID, yMin || yMinValue);
  $: yMaxStore.setWithKey(markID, yMax || yMaxValue);
  // we should set the extrema here.

  const xScale = getContext("rill:data-graphic:x-scale");
  const yScale = getContext("rill:data-graphic:y-scale");

  onDestroy(() => {
    xMinStore.removeKey(markID);
    xMaxStore.removeKey(markID);
    yMinStore.removeKey(markID);
    yMaxStore.removeKey(markID);
  });

  let lineFcn;
  $: if ($xScale) {
    lineFcn = lineFactory({
      xScale: $xScale,
      yScale: $yScale,
      curve,
      xAccessor,
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
