<script lang="ts">
  import { dataGraphicContext } from "./TimestampDetail.svelte";

  /** the starting value, in range space */
  export let start: number;
  /** the stopping value, in range space */
  export let stop: number;
  export let color: string;

  const plotConfig = dataGraphicContext.plotConfig.get();
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
