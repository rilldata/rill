<script lang="ts">
  import { fade } from "svelte/transition";

  export let gridCellSize = 50;

  export let scrollOffset = 0;
</script>

<svg
  in:fade={{ duration: 200 }}
  out:fade={{ duration: 700 }}
  width="100%"
  height="100%"
  xmlns="http://www.w3.org/2000/svg"
  class="absolute opacity-10"
>
  <defs>
    <pattern
      y={scrollOffset * -1}
      fill="black"
      id="grid"
      width={gridCellSize}
      height={gridCellSize}
      patternUnits="userSpaceOnUse"
    >
      <rect
        stroke="rgb(0,0,50)"
        width={gridCellSize}
        height={gridCellSize}
        fill="none"
      />
    </pattern>
  </defs>

  <rect width="100%" height="100%" fill="url(#grid)" />
</svg>
