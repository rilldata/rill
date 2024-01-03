<script>
  import { previousValueStore } from "@rilldata/web-common/lib/store-utils";
  import { onMount, onDestroy } from "svelte";
  import { writable } from "svelte/store";

  export let value;
  export let duration = 300;
  export let isDimension = false;

  let valueStore = writable(value);
  let previousValue = previousValueStore(valueStore);

  let showLabel = false;
  let timeoutId;

  onMount(() => {
    startTimeout();
  });

  function startTimeout() {
    clearTimeout(timeoutId);
    // Reset label visibility when value changes or timeout is reset
    showLabel = false;

    // Set timeout to show label after duration
    timeoutId = setTimeout(() => {
      showLabel = true;
    }, duration);
  }

  // reset timeout when value changes
  $: if (isDimension && value !== $previousValue) startTimeout();

  onDestroy(() => {
    clearTimeout(timeoutId);
  });
</script>

<slot
  visibility={isDimension ? (showLabel ? "visible" : "hidden") : "visible"}
/>
