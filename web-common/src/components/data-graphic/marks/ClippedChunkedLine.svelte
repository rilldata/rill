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

  export let lineColor: string;
  export let areaGradientColors: [string, string];

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
      {areaGradientColors}
      {delay}
      {duration}
      {data}
      {xAccessor}
      {yAccessor}
    />
  </g>
</svg>
