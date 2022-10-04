<script lang="ts">
  import { createEventDispatcher, getContext, onMount } from "svelte";
  import { get, Writable } from "svelte/store";

  export let selected = false;
  export let value;
  let element;

  const dispatch = createEventDispatcher();
  const callback = getContext("rill:app:tabgroup-callback") as Function;
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
  style:height="2rem"
  class:font-bold={element === $selectedValue}
  class="border-b-4 border-b-transparent px-4"
  on:click={() => {
    callback(element, value);
  }}
>
  <slot />
</button>
