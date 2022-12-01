<script lang="ts">
  import { eCDF } from "../utils";
  import Area from "./Area.svelte";
  import Line from "./Line.svelte";

  export let xAccessor: string;
  export let yAccessor: string = undefined;
  /** set transform to false if using with WithCumulative */
  export let transform = true;
  export let area = true;

  export let data;

  $: ecdf = transform ? eCDF(data, yAccessor) : data;
</script>

<g>
  <Line
    data={ecdf}
    {xAccessor}
    yAccessor="total"
    curve="curveStepAfter"
    color="hsla(1,10%, 60%, 1)"
  />
</g>
{#if area}
  <g>
    <Area
      data={ecdf}
      {xAccessor}
      yAccessor="total"
      curve="curveStepAfter"
      color="hsla(1,10%, 80%, .1)"
    />
  </g>
{/if}
