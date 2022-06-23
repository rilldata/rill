<script lang="ts">
  /**
   * Body.svelte
   * -----------
   * Creates a clip path group element that hides parts of children out of the viewport range
   */
  import { getContext } from "svelte";

  const plotConfig = getContext("rill:data-graphic:plot-config");
</script>

<clipPath id="data-graphic-{$plotConfig.id}">
  <rect
    x={$plotConfig.plotLeft}
    y={$plotConfig.plotTop}
    width={$plotConfig.plotRight - $plotConfig.plotLeft}
    height={$plotConfig.plotBottom - $plotConfig.plotTop}
  />
</clipPath>

<g clip-path="url(#data-graphic-{$plotConfig.id})">
  <slot />
</g>
