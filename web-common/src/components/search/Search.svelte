<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { onMount } from "svelte";
  import Search from "../icons/Search.svelte";

  /* Autofocus search bar on mount */
  export let autofocus = true;
  export let showBorderOnFocus = true;
  /* Input value being searched */
  export let value: string | number;
  /* Aria label for input */
  export let label: string | undefined = undefined;
  export let placeholder: string | undefined = undefined;
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

  $: resolvedLabel = label ?? m.common_search();
  $: resolvedPlaceholder = placeholder ?? m.common_search();

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
    class="flex absolute inset-y-0 items-center pl-2 text-fg-secondary"
    onclick={() => {
      ref?.focus();
    }}
  >
    <Search size={large ? "18px" : "16px"} className="text-fg-secondary" />
  </button>
  <svelte:element
    this={multiline ? "textarea" : "input"}
    bind:this={ref}
    type="text"
    autocomplete="off"
    class:focus={showBorderOnFocus}
    class:border
    class:bg-input={background}
    class:rounded-sm={rounded === "sm"}
    class:rounded-md={rounded === "md"}
    class:rounded-lg={rounded === "lg"}
    class="outline-none block w-full pl-8 p-1 {forcedInputStyle} resize-none text-fg-secondary placeholder-fg-secondary {large
      ? 'min-h-9'
      : ''}"
    class:h-full={large}
    {disabled}
    placeholder={resolvedPlaceholder}
    oninput={handleInput}
    onkeydown={handleKeyDown}
    aria-label={resolvedLabel}
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
