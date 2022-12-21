<script lang="ts">
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import type { PlotConfig } from "../../utils";

  /** the starting value, in range space */
  export let start: number;
  /** the stopping value, in range space */
  export let stop: number;
  export let color: string;

  const plotConfig: Writable<PlotConfig> = getContext(
    "rill:data-graphic:plot-config"
  );
</script>

<rect
  x={Math.min(start, stop)}
  y={$plotConfig.plotTop + $plotConfig.buffer}
  width={Math.abs(start - stop)}
  height={$plotConfig.plotBottom - $plotConfig.plotTop}
  fill={color}
  style:mix-blend-mode="darken"
/>
<line
  x1={start}
  x2={start}
  y1={$plotConfig.plotTop + $plotConfig.buffer}
  y2={$plotConfig.plotBottom}
  stroke="rgb(100,100,100)"
/>
