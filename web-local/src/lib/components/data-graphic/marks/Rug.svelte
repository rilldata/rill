<script lang="ts">
  import { contexts } from "$lib/components/data-graphic/constants";
  import WithSimpleLinearScale from "$lib/components/data-graphic/functional-components/WithSimpleLinearScale.svelte";
  import type { ScaleStore } from "$lib/components/data-graphic/state/types";
  import { interpolateReds } from "d3-scale-chromatic";
  import { getContext } from "svelte";
  import type { SimpleConfigurationStore } from "../state/types";

  export let data;

  export let size = 12;
  export let xAccessor = "value";
  export let densityAccessor = "count";

  export let side: "top" | "bottom" | "left" | "right" = "bottom";
  // create a special path that jumps from point to point.
  // each point could represent an arc line.

  function drawSegment(x) {
    return `M${x},0 L${x},${size} L${x},0`;
  }

  function drawSegments(data, xScale) {
    return data.map((point) => drawSegment(xScale(point[xAccessor]))).join("");
  }

  // bin data according to count
  $: largestCount = Math.max(...data.map((d) => d[densityAccessor]));
  // let's split everything up into 8.
  let tiers = 32;
  let counts = Array.from({ length: tiers }).fill([]);
  $: {
    data.forEach((datum) => {
      if (datum[densityAccessor] === 0) return;
      let c = Math.min(
        tiers - 1,
        Math.floor((datum[densityAccessor] / largestCount) * tiers)
      );
      counts[c] = [...counts[c], datum];
    });
    counts = counts;
  }

  const config = getContext(contexts.config) as SimpleConfigurationStore;
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
</script>

{#if xScale}
  <g
    transform="translate(0 {side === 'bottom'
      ? $config.bodyBottom - size
      : $config.top})"
  >
    <WithSimpleLinearScale domain={[0, 1]} range={[0.2, 0.65]} clamp let:scale>
      {#each counts as countSet, i}
        <g>
          <path
            d={drawSegments(countSet, $xScale)}
            stroke-width={1}
            stroke={interpolateReds(scale(i / tiers))}
          />
        </g>
      {/each}
    </WithSimpleLinearScale>
  </g>
{/if}
