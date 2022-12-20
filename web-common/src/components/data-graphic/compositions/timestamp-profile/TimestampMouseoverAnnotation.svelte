<script lang="ts">
  import {
    datePortion,
    formatInteger,
    removeTimezoneOffset,
    timePortion,
  } from "@rilldata/web-common/lib/formatters";
  import type { ScaleLinear } from "d3-scale";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { fly } from "svelte/transition";
  import { outline } from "../../actions/outline";
  import type { PlotConfig } from "../../utils";

  const X: Writable<ScaleLinear<number, number>> = getContext(
    "rill:data-graphic:X"
  );
  const Y: Writable<ScaleLinear<number, number>> = getContext(
    "rill:data-graphic:Y"
  );
  const config: Writable<PlotConfig> = getContext(
    "rill:data-graphic:plot-config"
  );

  export let point;
  export let xAccessor: string;
  export let yAccessor: string;
  $: xLabel = removeTimezoneOffset(point[xAccessor]);
</script>

<g>
  <line
    x1={$X(point[xAccessor])}
    x2={$X(point[xAccessor])}
    y1={$config.plotTop + $config.buffer}
    y2={$config.plotBottom}
    stroke="rgb(100,100,100)"
  />
  {#each [[yAccessor, "rgb(100,100,100)"]] as [accessor, color]}
    {@const cx = $X(point[xAccessor])}
    {@const cy = $Y(point[accessor])}
    {#if cx && cy}
      <circle {cx} {cy} r={3} fill={color} />
    {/if}
  {/each}
  <g
    in:fly={{ duration: 200, x: -16 }}
    out:fly={{ duration: 200, x: -16 }}
    font-size={$config.fontSize}
    style:user-select={"none"}
  >
    <text
      x={$config.plotLeft}
      y={$config.fontSize}
      class="fill-gray-500"
      use:outline
    >
      {datePortion(xLabel)}
    </text>
    <text
      x={$config.plotLeft}
      y={$config.fontSize * 2 + $config.textGap}
      class="fill-gray-500"
      use:outline
    >
      {timePortion(xLabel)}
    </text>
    <text
      x={$config.plotLeft}
      y={$config.fontSize * 3 + $config.textGap * 2}
      class="fill-gray-500"
      use:outline
    >
      {formatInteger(~~point[yAccessor])} row{#if point[yAccessor] !== 1}s{/if}
    </text>
  </g>
</g>
