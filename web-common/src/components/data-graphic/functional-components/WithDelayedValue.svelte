<!-- @component
Provides a slot prop output, whose value matches the value prop, but with a delay in ms.
-->
<script lang="ts">
  import { writable } from "svelte/store";

  export let value: any;
  const internalState = writable(value);
  export let delay = 0;
  let timeoutID;

  /** trigger delayed update */
  $: {
    if (timeoutID) clearTimeout(timeoutID);
    timeoutID = setTimeout(() => {
      internalState.set(value);
    }, delay);
  }
</script>

<slot output={$internalState} />
