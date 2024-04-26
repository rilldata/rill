<script lang="ts" context="module">
  type Event = MouseEvent & {
    currentTarget: EventTarget & HTMLButtonElement;
  };
</script>

<script lang="ts">
  export let dimension: number;
  export let direction: "NS" | "EW" = "EW";
  export let side: "left" | "right" | "top" | "bottom" =
    direction === "EW" ? "left" : "top";
  export let max = 440;
  export let min = 200;
  export let basis = 200;
  export let resizing = false;
  export let absolute = true;
  export let onMouseDown: ((e: MouseEvent) => void) | null = null;
  export let onUpdate: ((dimension: number) => void) | null = null;
  export let disabled = false;

  let start = 0;
  let startingDimension = dimension;

  function handleMousedown(e: Event) {
    startingDimension = dimension;
    resizing = true;

    if (direction === "EW") {
      start = e.clientX;
    } else {
      start = e.clientY;
    }

    if (onMouseDown) onMouseDown(e);
    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
  }

  function onMouseMove(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    let delta = 0;
    if (direction === "EW") {
      if (side === "left") {
        delta = start - e.clientX;
      } else {
        delta = e.clientX - start;
      }
    } else {
      if (side === "top") {
        delta = start - e.clientY;
      } else {
        delta = e.clientY - start;
      }
    }
    requestAnimationFrame(() => {
      dimension = Math.min(max, Math.max(min, startingDimension + delta));
      if (onUpdate) onUpdate(dimension);
    });
  }

  function onMouseUp() {
    resizing = false;
    window.removeEventListener("mousemove", onMouseMove);
    window.removeEventListener("mouseup", onMouseUp);
  }

  function handleDoubleClick() {
    dimension = basis;
    if (onUpdate) onUpdate(dimension);
  }
</script>

<button
  {disabled}
  class:absolute
  class="{direction} {side}"
  on:mousedown|stopPropagation|preventDefault={handleMousedown}
  on:dblclick={handleDoubleClick}
>
  <slot />
</button>

<style lang="postcss">
  button {
    @apply z-10 flex-none;
    /* @apply bg-red-400; */
  }

  button:disabled {
    @apply cursor-default;
  }

  .NS {
    @apply w-full h-2 pr-8;
    @apply cursor-row-resize;
  }

  .EW {
    @apply w-2 h-full;
    @apply cursor-col-resize;
  }

  .left {
    @apply -left-1;
  }

  .right {
    @apply right-0;
  }

  .top {
    @apply -top-1;
  }

  .bottom {
    @apply -bottom-1;
  }
</style>
