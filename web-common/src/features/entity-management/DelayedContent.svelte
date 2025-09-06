<!-- @component 
  This component is used to only show content if the visible state is true after a delay.
  This is handy for preventing content from flickering when loading states change rapidly.
-->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";

  export let visible: boolean;
  export let delay: number = 300;

  const showContent = writable(false);

  let timeoutId: ReturnType<typeof setTimeout> | undefined;

  $: {
    clearTimeout(timeoutId);
    if (visible) {
      timeoutId = setTimeout(() => showContent.set(true), delay);
    } else {
      showContent.set(false);
    }
  }

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showContent}
  <slot />
{/if}
