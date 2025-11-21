<script lang="ts">
  import { onMount } from "svelte";
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
  export let theme = false;
  export let rounded: "sm" | "md" | "lg" = "sm";
  export let onSubmit: () => void = () => {};

  /* Reference of input DOM element */
  let ref: HTMLInputElement | HTMLTextAreaElement;

  function handleKeyDown(event) {
    if (event.code == "Enter") {
      event.preventDefault();
      event.stopPropagation();
      onSubmit();
      return false;
    }
  }

  /**
   * We cant do a bind on svelte:element. We get `'value' is not a valid binding on <svelte:element> elements` error.
   * So we need this to still keep `<svelte:element this={multiline ? "textarea" : "input"} ...`
   */
  function handleInput(event) {
    value = event.target?.value;
    if (multiline) updateTextAreaHeight();
  }

  function updateTextAreaHeight() {
    ref.style.height = ref.scrollHeight + "px"; // Set to scroll height
  }

  onMount(() => {
    if (!retainValueOnMount) value = "";
    // Keep ref optional here. If component is unmounted before this animation frame runs, ref will be null and throw a TypeError
    if (autofocus) window.requestAnimationFrame(() => ref?.focus());
    if (multiline) updateTextAreaHeight();
  });
</script>

<form
  class:theme
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
  <svelte:element
    this={multiline ? "textarea" : "input"}
    bind:this={ref}
    type="text"
    autocomplete="off"
    class:focus={showBorderOnFocus}
    class:bg-gray-50={background}
    class:border
    class:rounded-sm={rounded === "sm"}
    class:rounded-md={rounded === "md"}
    class:rounded-lg={rounded === "lg"}
    class="outline-none block w-full pl-8 p-1 {forcedInputStyle} resize-none"
    class:h-full={large}
    {disabled}
    {placeholder}
    on:input={handleInput}
    on:keydown={handleKeyDown}
    aria-label={label}
    role="textbox"
    tabindex="-1"
    {value}
  />
</form>

<style lang="postcss">
  .focus:focus {
    @apply border-primary-400;
  }

  .theme .focus:focus {
    @apply border-theme-400;
  }

  textarea {
    height: 28px;
    /* min height for 1 row */
    min-height: 28px;
    /* Max of 5 rows. 28 + 16 * 5 = 92 */
    max-height: 92px;
  }
</style>
