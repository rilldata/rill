<!--
  Wraps text with animated trailing dots that cycle: "" → "." → ".." → "..."
  Used for loading states like "Thinking", "Querying metrics", etc.
-->
<script lang="ts">
  import { onDestroy } from "svelte";

  /** Interval between dot changes in milliseconds */
  export let interval = 300;

  let dotCount = 0;
  let dotInterval: ReturnType<typeof setInterval> | null = null;

  // Start animation on mount
  dotInterval = setInterval(() => {
    dotCount = (dotCount + 1) % 4;
  }, interval);

  onDestroy(() => {
    if (dotInterval) clearInterval(dotInterval);
  });

  $: dots = ".".repeat(dotCount);
</script>

<span class="animated-dots"><slot /><span class="dots">{dots}</span></span>

<style lang="postcss">
  .animated-dots {
    @apply inline-flex;
  }

  .dots {
    @apply inline-block;
    @apply w-[1em] text-left;
  }
</style>
