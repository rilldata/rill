<!-- @component 
  This component is used to only show a spinner if the loading state is true after a delay.
  This is handy for preventing the spinner from flickering.
-->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { writable } from "svelte/store";
  import Spinner from "./Spinner.svelte";

  export let isLoading: boolean;
  export let delay: number = 300;
  export let duration: number = 500;
  export let size: string = "1em";

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

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

{#if $showSpinner}
  <Spinner status={EntityStatus.Running} {duration} {size} />
{/if}
