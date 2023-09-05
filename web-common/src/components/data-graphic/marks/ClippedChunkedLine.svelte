<script lang="ts">
  import { getContext } from "svelte";
  import { ChunkedLine } from "./";
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";

  export let start;
  export let end;

  export let data;
  export let xAccessor: string;
  export let yAccessor: string;

  /** time in ms to trigger a delay when the underlying data changes */
  export let delay;
  export let duration;

  export let lineColor = "hsla(217,60%, 55%, 1)";
  export let areaColor = "hsla(217,70%, 80%, .4)";

  const xScale = getContext(contexts.scale("x")) as ScaleStore;
</script>

<svg>
  <!-- Define the clip path using the x positions of the scrub start/end -->
  <clipPath id="clip">
    <rect
      x={$xScale(start)}
      width={$xScale(end) - $xScale(start)}
      height="100%"
    />
  </clipPath>

  <!-- Apply the clip path to the svg element -->
  <g clip-path="url(#clip)">
    <ChunkedLine
      {lineColor}
      {areaColor}
      {delay}
      {duration}
      {data}
      {xAccessor}
      {yAccessor}
    />
  </g>
</svg>
