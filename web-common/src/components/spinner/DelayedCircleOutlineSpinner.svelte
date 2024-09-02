<!-- @component
  This component is used to only show a spinner if the loading state is true after a delay.
  This is handy for preventing the spinner from flickering.

  This variant is used in places where a spinner is shown instead of the action leading to the spinner.
  Use the slot to show the action and pass in the `isLoading` to control loading state.
-->
<script lang="ts">
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import { onDestroy } from "svelte";
  import { writable } from "svelte/store";

  export let isLoading: boolean;
  export let delay: number = 300;
  export let size: string = "1em";

  const showSpinner = writable(false);

  let timeoutId: ReturnType<typeof setTimeout>;

  $: {
    clearTimeout(timeoutId);
    if (isLoading) {
      timeoutId = setTimeout(() => showSpinner.set(true), delay);
    } else {
      showSpinner.set(false);
    }
  }

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showSpinner}
  <LoadingCircleOutline {size} />
{:else}
  <slot />
{/if}
