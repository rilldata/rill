<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../contexts";

  const xScale = getContext(contexts.scale("x"));
  const yScale = getContext(contexts.scale("y"));
  const config = getContext(contexts.config);

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

  let container;
  let xAxisLength;
  let yAxisLength;
  let xTickCount = 0;
  let yTickCount = 0;
  $: if (container) {
    xAxisLength = container.getBoundingClientRect().width;
    // do we ensure different spacing in one case vs. another?
    xTickCount = ~~(xAxisLength / 20);
    xTickCount = Math.max(3, ~~(xTickCount / 100));

    yAxisLength = container.getBoundingClientRect().height;
    // do we ensure different spacing in one case vs. another?
    yTickCount = ~~(yAxisLength / 20);
    yTickCount = Math.max(3, ~~(yTickCount / 100));
  }
  $: xCopy = xScale.type === "date" ? $xScale.copy().nice() : $xScale;
  $: yCopy = yScale.type === "date" ? $yScale.copy().nice() : $yScale;
</script>

<g bind:this={container} shape-rendering="crispEdges">
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
