<script lang="ts">
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";
  import LoadingRows from "./LoadingRows.svelte";

  export let isLoading: boolean;
  export let isFetching: boolean;
  export let isPending: boolean;
  export let delay: number = 300;
  export let rowCount: number;
  export let columnCount: number = 4;

  const showPlaceholder = writable(true);

  let timeoutId: ReturnType<typeof setTimeout> | undefined = undefined;
  let previousRowCount = rowCount ?? 7;

  $: {
    if (timeoutId) clearTimeout(timeoutId);

    if (isLoading || isPending) {
      showPlaceholder.set(true);
    } else if (isFetching) {
      timeoutId = setTimeout(() => {
        showPlaceholder.set(true);
      }, delay);
    } else {
      showPlaceholder.set(false);
      previousRowCount = rowCount;
    }
  }

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showPlaceholder}
  <LoadingRows rows={previousRowCount || 7} columns={columnCount} />
{:else}
  <slot />
{/if}
