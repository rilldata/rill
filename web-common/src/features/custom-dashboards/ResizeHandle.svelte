<script lang="ts">
  import type { Vector } from "./types";

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let position: Vector;
  export let dimensions: Vector;
  export let side: [number, number];
  export let i: number;
  export let selected: boolean;
  export let scale: number;

  $: ew = side[0] !== 0.5;
  $: ns = side[1] !== 0.5;
  $: corner = ns && ew;
  $: span = corner ? 12 : dimensions[Number(ew)] - 6;

  function handleMouseDown(e: MouseEvent) {
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
  }
</script>

<button
  data-index={i}
  style:width="{span}px"
  style:left="{side[0] * 100}%"
  style:top="{side[1] * 100}%"
  class:rotate-90={ew}
  class:!z-50={corner}
  class:cursor-ns-resize={!corner && ns}
  class:cursor-ew-resize={!corner && ew}
  class:cursor-nwse-resize={corner && side[0] === side[1]}
  class:cursor-nesw-resize={corner && side[0] !== side[1]}
  on:mousedown={handleMouseDown}
>
  {#if !corner}
    <span class="line" style:height="{1 / scale}px" class:hide={!selected} />
  {/if}
  <span
    class="square"
    style:width="{6 / scale}px"
    style:height="{6 / scale}px"
    style:border-width="{1 / scale}px"
    class:hide={!selected}
  />
</button>

<style lang="postcss">
  button {
    @apply z-40 absolute h-3;
    @apply -translate-y-1/2 -translate-x-1/2;
    @apply flex items-center justify-center;
  }

  .square {
    @apply aspect-square bg-white border-primary-400 z-50;
  }

  .line {
    @apply absolute bg-primary-400 w-full;
  }

  .hide {
    @apply invisible;
  }
</style>
