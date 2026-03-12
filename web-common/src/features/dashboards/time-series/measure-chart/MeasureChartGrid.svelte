<script lang="ts">
  import type { ScaleLinear } from "d3-scale";

  export let yTicks: number[];
  export let xTickIndices: number[];
  export let yScale: ScaleLinear<number, number>;
  export let xScale: ScaleLinear<number, number>;
  export let plotLeft: number;
  export let plotWidth: number;
  export let plotTop: number;
  export let plotHeight: number;
  export let axisFormatter: (value: number) => string;

  const DASH = "1,1.5";
</script>

<g class="y-axis">
  {#each yTicks as tick (tick)}
    <text
      class="fill-fg-muted text-[11px]"
      text-anchor="start"
      x={plotLeft + plotWidth + 4}
      y={yScale(tick) + 4}
    >
      {axisFormatter(tick)}
    </text>
    <line
      class="stroke-gray-300"
      x1={plotLeft}
      x2={plotLeft + plotWidth}
      y1={yScale(tick)}
      y2={yScale(tick)}
      stroke-width="0.75"
      stroke-dasharray={DASH}
    />
  {/each}
</g>

<g class="x-axis">
  {#each xTickIndices as idx (idx)}
    <line
      class="stroke-border"
      x1={xScale(idx)}
      x2={xScale(idx)}
      y1={plotTop}
      y2={plotTop + plotHeight}
      stroke-width="0.75"
      stroke-dasharray={DASH}
    />
  {/each}
</g>

<!-- Zero line -->
<line
  class="stroke-gray-300"
  x1={plotLeft}
  x2={plotLeft + plotWidth}
  y1={yScale(0)}
  y2={yScale(0)}
/>
