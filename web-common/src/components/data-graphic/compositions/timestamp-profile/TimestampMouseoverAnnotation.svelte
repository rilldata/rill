<script lang="ts">
  import {
    datePortion,
    formatInteger,
    timePortion,
  } from "@rilldata/web-common/lib/formatters";
  import { removeLocalTimezoneOffset } from "@rilldata/web-common/lib/time/timezone";
  import { fly } from "svelte/transition";
  import { outline } from "../../actions/outline";
  import { dataGraphicContext } from "./TimestampDetail.svelte";

  const X = dataGraphicContext.x.get();
  const Y = dataGraphicContext.y.get();
  const config = dataGraphicContext.plotConfig.get();

  export let point;
  export let xAccessor: string;
  export let yAccessor: string;

  $: xLabel = removeLocalTimezoneOffset(point[xAccessor]);
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
    in:fly|global={{ duration: 200, x: -16 }}
    out:fly|global={{ duration: 200, x: -16 }}
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
      {formatInteger(Math.trunc(point[yAccessor]))} row{#if point[yAccessor] !== 1}s{/if}
    </text>
  </g>
</g>
