<script lang="ts">
  import { onMount } from "svelte";
  import Search from "../icons/Search.svelte";

  /* Autofocus search bar on mount */
  export let autofocus = true;
  /* Get reference of input DOM element */
  export let ref = null;
  /* Input value being searched */
  export let value;
  /* Aria label for input */
  export let label = "Search";
  export let placeholder = "Search";

  function handleKeyDown(event) {
    if (event.code == "Enter") {
      event.preventDefault();
      return false;
    }
  }

  onMount(() => {
    // Keep ref optional here. If component is unmounted before this animation frame runs, ref will be null and throw a TypeError
    if (autofocus) window.requestAnimationFrame(() => ref?.focus());
  });
</script>

<form class="relative w-full">
  <div class="flex absolute inset-y-0 items-center pl-2 ui-copy-icon">
    <Search />
  </div>
  <input
    bind:this={ref}
    type="text"
    autocomplete="off"
    class="bg-white border border-gray-200 focus:border-blue-400 outline-none rounded-sm block w-full pl-8 p-1"
    {placeholder}
    bind:value
    on:input
    on:keydown={handleKeyDown}
    aria-label={label}
  />
</form>
