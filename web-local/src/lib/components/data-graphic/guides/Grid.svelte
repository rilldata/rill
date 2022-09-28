<!--
  Draws grid lines according to the specified axis ticks.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";
  import { getTicks } from "../utils";

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

  let xTicks = [];
  let yTicks = [];

  $: if ($config) {
    xTicks = getTicks(
      "x",
      $xScale,
      $config.graphicWidth,
      $config[`xType`] === "date"
    );

    yTicks = getTicks(
      "y",
      $yScale,
      $config.graphicHeight,
      $config[`yType`] === "date"
    );
  }
</script>

<g shape-rendering="crispEdges">
  {#if showX}
    {#each xTicks as tick}
      <line
        x1={$xScale(tick)}
        x2={$xScale(tick)}
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
    {#each yTicks as tick}
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
