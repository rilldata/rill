<script lang="ts">
  import { fade } from "svelte/transition";

  export let gridCell = 50;
  export let scrollOffset = 0;
  export let gapSize: number;

  $: clampedGap = Math.max(0.5, gapSize);
</script>

<svg
  in:fade={{ duration: 200 }}
  out:fade={{ duration: 700 }}
  width="100%"
  height="100%"
  xmlns="http://www.w3.org/2000/svg"
  class="bg-slate-100"
  style:position="absolute"
  style:opacity="1"
>
  <defs>
    <pattern
      y={scrollOffset * -1}
      id="cd-grid"
      width={gridCell}
      height={gridCell}
      patternUnits="userSpaceOnUse"
    >
      <rect
        rx="0.3%"
        x={clampedGap}
        y={clampedGap}
        width={gridCell - clampedGap * 2}
        height={gridCell - clampedGap * 2}
        class="fill-slate-200"
      />
    </pattern>
  </defs>

  <rect width="100%" height="100%" fill="url(#cd-grid)" />
</svg>
