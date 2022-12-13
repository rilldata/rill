<script lang="ts">
  import { getContext, onMount } from "svelte";
  import { get, Writable } from "svelte/store";

  export let selected = false;
  export let value;

  let element;

  const callback = getContext("rill:app:tabgroup-callback") as (
    element,
    value
  ) => void;
  const selectedValue = getContext(
    "rill:app:tabgroup-selected"
  ) as Writable<unknown>;
  onMount(() => {
    if (get(selectedValue) === undefined) selectedValue.set(element);
  });
</script>

<button
  bind:this={element}
  role="tab"
  aria-selected={selected}
  style:min-width="40px"
  class:ui-copy-strong={element === $selectedValue}
  class="px-4 pb-0 mb-0"
  on:click={() => {
    if (!selected) callback(element, value);
  }}
>
  <slot />
</button>
