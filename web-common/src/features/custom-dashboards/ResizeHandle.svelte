<script lang="ts">
  import type { Vector } from "./types";

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let position: Vector;
  export let dimensions: Vector;
  export let side: [number, number];
  export let i: number;
</script>

<button
  style:left="{side[0] * 100}%"
  style:top="{side[1] * 100}%"
  style:width={side[0] === 0.5 ? "calc(100% - 12px)" : "12px"}
  style:height={side[1] === 0.5 ? "calc(100% - 12px)" : "12px"}
  class:cursor-ns-resize={side[0] === 0.5}
  class:cursor-ew-resize={side[1] === 0.5}
  class:cursor-nesw-resize={(side[0] === 1 && side[1] === 0) ||
    (side[0] === 0 && side[1] === 1)}
  class:cursor-nwse-resize={(side[0] === 0 && side[1] === 0) ||
    (side[0] === 1 && side[1] === 1)}
  data-index={i}
  on:mousedown={(e) => {
    dispatch("change", {
      e,
      dimensions,
      position,
      changeDimensions: [
        side[0] === 0.5 ? 0 : side[0] ? 1 : -1,
        side[1] === 0.5 ? 0 : side[1] ? 1 : -1,
      ],
      changePosition: [side[0] ? 0 : 1, side[1] ? 0 : 1],
    });
  }}
/>

<style lang="postcss">
  button {
    @apply z-40 absolute;
    @apply -translate-y-1/2 -translate-x-1/2;
  }
</style>
