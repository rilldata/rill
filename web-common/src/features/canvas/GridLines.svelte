<script lang="ts">
  import { fade } from "svelte/transition";
  import * as defaults from "./constants";

  export let scale: number;
  export let height: number;

  $: columnWidth = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;
  $: columns = Array.from({ length: defaults.COLUMN_COUNT + 1 });
  $: rows = Array.from({ length: Math.ceil(height / defaults.ROW_HEIGHT) + 1 });
</script>

<svg
  in:fade={{ duration: 200 }}
  out:fade={{ duration: 700 }}
  width="100%"
  {height}
  xmlns="http://www.w3.org/2000/svg"
  class="pointer-events-none absolute opacity-100"
>
  <!-- Column lines -->
  {#each columns as _, i}
    <line
      x1={i * columnWidth * scale}
      y1={0}
      x2={i * columnWidth * scale}
      y2={height}
      class="stroke-slate-200"
      stroke-width="1"
    />
  {/each}

  <!-- Row lines -->
  {#each rows as _, i}
    <line
      x1={0}
      y1={i * defaults.ROW_HEIGHT * scale}
      x2={defaults.DASHBOARD_WIDTH * scale}
      y2={i * defaults.ROW_HEIGHT * scale}
      class="stroke-slate-200"
      stroke-width="1"
    />
  {/each}
</svg>

<style>
  svg {
    transform-origin: top left;
  }
</style>
