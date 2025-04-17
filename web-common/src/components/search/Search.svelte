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
  export let multiline = false;
  export let border = true;
  export let background = true;
  export let large = false;
  export let disabled = false;
  export let retainValueOnMount = false;
  export let forcedInputStyle = "";

  /* Reference of input DOM element */
  let ref: HTMLInputElement | HTMLTextAreaElement;

  const dispatch = createEventDispatcher();

  function handleKeyDown(event) {
    if (event.code == "Enter") {
      event.preventDefault();
      event.stopPropagation();
      dispatch("submit");
      return false;
    }
  }

  onMount(() => {
    if (!retainValueOnMount) value = "";
    // Keep ref optional here. If component is unmounted before this animation frame runs, ref will be null and throw a TypeError
    if (autofocus) window.requestAnimationFrame(() => ref?.focus());
  });

  const BaseHeight = 28;
  const HeightPerLine = 16;
  const MaxLines = 5;

  // For tailwind compiler: h-[28px] h-[44px] h-[60px] h-[76px] h-[92px]
  let height = "h-[28px]";
  function updateHeight(value: string) {
    const lines = value.split("\n").length;
    const correctedLines = Math.max(
      // Show minimum of 1 line
      1,
      // We expand the input only till MaxLines
      Math.min(lines, MaxLines),
    );

    height = `h-[${BaseHeight + (correctedLines - 1) * HeightPerLine}px]`;
    console.log(lines, correctedLines, height);
  }

  $: if (multiline && typeof value === "string") updateHeight(value);
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
  {#if multiline}
    <textarea
      bind:this={ref}
      autocomplete="off"
      class:focus={showBorderOnFocus}
      class:bg-slate-50={background}
      class:border
      class:border-gray-200={border}
      class="outline-none rounded-[2px] block w-full pl-8 p-1 {forcedInputStyle} {height} resize-none"
      class:h-full={large}
      {disabled}
      {placeholder}
      bind:value
      on:input
      on:keydown={handleKeyDown}
      aria-label={label}
    />
  {:else}
    <input
      bind:this={ref}
      type="text"
      autocomplete="off"
      class:focus={showBorderOnFocus}
      class:bg-slate-50={background}
      class:border
      class:border-gray-200={border}
      class="outline-none rounded-[2px] block w-full pl-8 p-1 {forcedInputStyle}"
      class:h-full={large}
      {disabled}
      {placeholder}
      bind:value
      on:input
      on:keydown={handleKeyDown}
      aria-label={label}
    />
  {/if}
</form>

<style lang="postcss">
  .focus:focus {
    @apply border-primary-400;
  }
</style>
