<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import Search from "../icons/Search.svelte";

  /* Autofocus search bar on mount */
  export let autofocus = true;
  export let showBorderOnFocus = true;
  /* Input value being searched */
  export let value: string | number;
  /* Aria label for input */
  export let label = "Search";
  export let placeholder = "Search";
  export let border = true;
  export let background = true;
  export let large = false;
  export let disabled = false;
  export let retailValueOnMount = false;

  /* Reference of input DOM element */
  let ref: HTMLInputElement;

  const dispatch = createEventDispatcher();

  function handleKeyDown(event) {
    if (event.code == "Enter") {
      event.preventDefault();
      dispatch("submit");
      return false;
    }
  }

  onMount(() => {
    if (!retailValueOnMount) value = "";
    // Keep ref optional here. If component is unmounted before this animation frame runs, ref will be null and throw a TypeError
    if (autofocus) window.requestAnimationFrame(() => ref?.focus());
  });
</script>

<form
  class="relative w-full {disabled
    ? 'pointer-events-none opacity-50 cursor-not-allowed'
    : ''}"
>
  <button
    type="button"
    class="flex absolute inset-y-0 items-center pl-2 ui-copy-icon"
    on:click={() => {
      ref?.focus();
    }}
  >
    <Search size={large ? "18px" : "16px"} />
  </button>
  <input
    bind:this={ref}
    type="text"
    autocomplete="off"
    class:focus={showBorderOnFocus}
    class:bg-slate-50={background}
    class:border
    class:border-gray-200={border}
    class="outline-none rounded-sm block w-full pl-8 p-1"
    class:h-full={large}
    {disabled}
    {placeholder}
    bind:value
    on:input
    on:keydown={handleKeyDown}
    aria-label={label}
  />
</form>

<style lang="postcss">
  .focus:focus {
    @apply border-primary-400;
  }
</style>
