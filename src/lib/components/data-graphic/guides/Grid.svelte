<!--
  Draws grid lines according to the specified axis ticks.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";

  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;
  const config = getContext(contexts.config) as SimpleConfigurationStore;

  export let showX = true;
  export let showY = true;

  export let xColor = "rgba(0,0,0,.3)";
  export let xAlpha = 1;
  export let xDashArray = "1,1";
  export let xThickness = 1;
  export let yColor = "rgba(0,0,0,.3)";
  export let yAlpha = 1;
  export let yDashArray = "1,1";
  export let yThickness = 1;

  let xAxisLength;
  let yAxisLength;
  let xTickCount = 0;
  let yTickCount = 0;

  $: if ($config) {
    xAxisLength = $config.graphicWidth;
    // do we ensure different spacing in one case vs. another?
    xTickCount = ~~(xAxisLength / 20);
    xTickCount = Math.max(2, ~~(xTickCount / 100));

    yAxisLength = $config.graphicHeight;
    // do we ensure different spacing in one case vs. another?
    yTickCount = ~~(yAxisLength / 20);
    yTickCount = Math.max(2, ~~(yTickCount / 100));
  }
  $: xCopy = xScale.type === "date" ? $xScale.copy().nice() : $xScale;
  $: yCopy = yScale.type === "date" ? $yScale.copy().nice() : $yScale;
</script>

<g shape-rendering="crispEdges">
  {#if showX}
    {#each xCopy.ticks(xTickCount) as tick}
      <line
        x1={xCopy(tick)}
        x2={xCopy(tick)}
        y1={$config.bodyTop}
        y2={$config.bodyBottom}
        stroke={xColor}
        stroke-width={xThickness}
        stroke-dasharray={xDashArray}
        opacity={xAlpha}
      />
    {/each}
  {/if}
  {#if showY}
    {#each yCopy.ticks(yTickCount) as tick}
      <line
        y1={$yScale(tick)}
        y2={$yScale(tick)}
        x1={$config.bodyLeft}
        x2={$config.bodyRight}
        stroke={yColor}
        stroke-width={yThickness}
        stroke-dasharray={yDashArray}
        opacity={yAlpha}
      />
    {/each}
  {/if}
</g>
