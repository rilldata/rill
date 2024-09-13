<script lang="ts">
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";

  export let rows = 7;
  export let columns = 4;
  export let isLoading: boolean;
  export let delay: number = 300;

  const showLoadingState = writable(false);

  let timeoutId;

  $: {
    clearTimeout(timeoutId);
    if (isLoading) {
      timeoutId = setTimeout(() => showLoadingState.set(true), delay);
    } else {
      showLoadingState.set(false);
    }
  }

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showLoadingState}
  {#each { length: rows } as _, i (i)}
    <tr>
      <td />
      {#each { length: columns } as _, i (i)}
        <td>
          <div />
        </td>
      {/each}
    </tr>
  {/each}
{/if}

<style lang="postcss">
  td {
    height: 22px;
    @apply p-1 py-[5px];
  }
  div {
    @apply size-full bg-gray-200 animate-pulse rounded-full;
  }
</style>
