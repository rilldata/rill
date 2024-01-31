<script lang="ts">
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import type { LayoutElement } from "./types";

  const outputLayout = getContext(
    "rill:app:output-layout",
  ) as Writable<LayoutElement>;

  export let className = "";

  onMount(() => {
    parentElement = splitter.parentElement;
  });

  let parentElement: HTMLElement | null = null;
  let splitter: HTMLButtonElement;

  let parentHeight = 0;

  function onMouseMove(e: MouseEvent) {
    $outputLayout.value = Math.min(
      parentHeight - 200,
      Math.max(200, parentHeight - e.clientY),
    );
  }

  function onMouseUp() {
    window.removeEventListener("mousemove", onMouseMove);
  }

  function startDrag() {
    if (!parentElement) return;

    parentHeight = parentElement.clientHeight;

    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
  }
</script>

<button class={className} bind:this={splitter} on:mousedown={startDrag}>
  <div class="line" />
  <span class="handle" />
</button>

<style lang="postcss">
  button {
    @apply cursor-move;
    @apply flex items-center justify-center;
    @apply w-full h-2;
  }

  .handle {
    @apply absolute;
    @apply border-gray-400 border bg-white;
    @apply rounded h-1 w-8;
  }

  .line {
    @apply h-[1px] w-full;
    @apply bg-gray-300;
  }
</style>
