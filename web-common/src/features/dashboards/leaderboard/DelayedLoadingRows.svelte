<!-- @component 
  This component is used to only show loading rows if the loading state is true after a delay.
  This is handy for preventing the loading rows from flickering.
-->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";
  import LoadingRows from "./LoadingRows.svelte";

  export let isLoading: boolean;
  export let delay: number = 300;
  export let rows: number = 7;
  export let columns: number = 4;

  const showLoading = writable(false);

  let timeoutId: NodeJS.Timeout;

  $: {
    clearTimeout(timeoutId);
    if (isLoading) {
      timeoutId = setTimeout(() => showLoading.set(true), delay);
    } else {
      showLoading.set(false);
    }
  }

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showLoading}
  <LoadingRows {rows} {columns} />
{/if}
