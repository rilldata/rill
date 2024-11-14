<script lang="ts">
  import { fade } from "svelte/transition";

  export let gridCell = 50;
  export let scrollOffset = 0;
  export let gapSize: number;
  export let radius: number;
  export let scale: number;

  $: clampedGap = Math.max(0.5, gapSize);
  $: cellSize = gridCell * scale;
  $: scaledGap = clampedGap * scale;
</script>

<svg
  in:fade={{ duration: 200 }}
  out:fade={{ duration: 700 }}
  width="100%"
  height="100%"
  xmlns="http://www.w3.org/2000/svg"
  class="bg-slate-100 pointer-events-none"
  style:position="absolute"
  style:opacity="1"
>
  <defs>
    <pattern
      y={scrollOffset * -1}
      id="cd-grid"
      width={cellSize}
      height={cellSize}
      patternUnits="userSpaceOnUse"
    >
      <rect
        rx={radius * scale}
        x={scaledGap}
        y={scaledGap}
        width={cellSize - scaledGap * 2}
        height={cellSize - scaledGap * 2}
        class="fill-slate-200"
      />
    </pattern>
  </defs>

  <rect width="100%" height="100%" fill="url(#cd-grid)" />
</svg>

<style>
  svg {
    transform-origin: top left;
  }
</style>
