<script>
  import { previousValueStore } from "@rilldata/web-common/lib/store-utils";
  import { onMount, onDestroy } from "svelte";
  import { writable } from "svelte/store";

  export let label;
  export let value;

  let valueStore = writable(value);
  let previousValue = previousValueStore(valueStore);

  let showLabel = false;
  let timeoutId;

  onMount(() => {
    startTimeout();
  });

  function startTimeout() {
    clearTimeout(timeoutId);

    // Set timeout to show label after 2 seconds
    timeoutId = setTimeout(() => {
      showLabel = true;
    }, 600);
  }

  // reset timeout when value changes
  $: if (value !== $previousValue) startTimeout();

  onDestroy(() => {
    clearTimeout(timeoutId);
    showLabel = false;
  });
</script>

{showLabel ? label : ""}
