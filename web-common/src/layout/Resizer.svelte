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
  export let onMouseUp: (() => void) | null = null;
  export let disabled = false;
  export let justify: "center" | "start" | "end" = "center";
  export let hang = true;

  let start = 0;
  let startingDimension = dimension;
  let hover = false;
  let timeout: ReturnType<typeof setTimeout> | null = null;

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
    window.addEventListener("mouseup", handleMouseUp);
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

  function handleMouseUp() {
    resizing = false;
    hover = false;
    if (onMouseUp) onMouseUp();
    window.removeEventListener("mousemove", onMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);
  }

  function handleDoubleClick() {
    dimension = basis;
    if (onUpdate) onUpdate(dimension);
  }
</script>

<button
  {disabled}
  class:absolute
  class="{direction} {side} justify-{justify}"
  class:hang
  on:mousedown|stopPropagation|preventDefault={handleMousedown}
  on:dblclick|stopPropagation={handleDoubleClick}
  on:mouseenter={() => {
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => (hover = true), 150);
  }}
  on:mouseleave={() => {
    if (timeout) clearTimeout(timeout);
    timeout = null;
    hover = false;
  }}
>
  {#if hover || resizing}
    <slot />
  {/if}
</button>

<style lang="postcss">
  button {
    @apply z-50 flex-none;
    @apply pointer-events-auto;
    @apply flex items-center;
    /* @apply bg-red-500 opacity-50; */
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

  .NS.maxed {
    @apply cursor-s-resize;
  }

  .NS.minned {
    @apply cursor-n-resize;
  }

  .EW.minned {
    @apply cursor-e-resize;
  }

  .EW.maxed {
    @apply cursor-w-resize;
  }

  .left.hang {
    @apply -left-1;
  }

  .right.hang {
    @apply -right-1;
  }

  .top.hang {
    @apply -top-1;
  }

  .bottom.hang {
    @apply -bottom-1;
  }

  .left {
    @apply left-0;
  }

  .right {
    @apply right-0;
  }

  .top {
    @apply top-0;
  }

  .bottom {
    @apply bottom-0;
  }
</style>
