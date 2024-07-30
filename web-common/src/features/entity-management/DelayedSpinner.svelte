<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { writable } from "svelte/store";

  export let isLoading: boolean;
  export let delay = 300;
  const showSpinner = writable(false);

  let timeoutId;

  $: {
    clearTimeout(timeoutId);
    if (isLoading) {
      timeoutId = setTimeout(() => showSpinner.set(true), delay);
    } else {
      showSpinner.set(false);
    }
  }

  onMount(() => {
    if (isLoading) {
      timeoutId = setTimeout(() => showSpinner.set(true), delay);
    }
  });

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showSpinner}
  <slot />
{/if}
