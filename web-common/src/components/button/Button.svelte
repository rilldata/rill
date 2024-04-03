<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { createEventDispatcher } from "svelte";

  type ButtonType =
    | "primary"
    | "secondary"
    | "highlighted"
    | "text"
    | "link"
    | "brand"
    | "add";

  export let type: ButtonType = "primary";
  export let status: "info" | "error" = "info";
  export let disabled = false;
  export let compact = false;
  export let submitForm = false;
  export let form = "";
  export let label: string | undefined | null = null;
  export let square = false;
  export let circle = false;
  export let selected = false;
  export let large = false;
  export let small = false;
  export let noStroke = false;
  export let dashed = false;
  export let rounded = false;
  export let href: string | null = null;
  export let builders: Builder[] = [];
  export let div = false;

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };
</script>

<svelte:element
  this={div ? "div" : href ? "a" : "button"}
  role={div ? "presentation" : "button"}
  tabindex={0}
  {href}
  class="{$$props.class} {type}"
  {disabled}
  class:square
  class:circle
  class:selected
  class:large
  class:small
  class:dashed
  class:compact
  class:rounded
  class:danger={status === "error"}
  class:no-stroke={noStroke}
  type={submitForm ? "submit" : "button"}
  form={submitForm ? form : undefined}
  aria-label={label}
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  on:click={handleClick}
>
  <slot />
</svelte:element>

<style lang="postcss">
  button,
  a,
  div {
    @apply flex text-center items-center justify-center;
    @apply text-xs leading-snug font-normal;
    @apply gap-x-2 min-w-fit;
    @apply rounded-[2px];
    @apply px-3 h-7 min-h-[28px] cursor-pointer;
  }

  button:disabled {
    @apply opacity-50 cursor-not-allowed;
  }

  /* PRIMARY STYLES */

  .primary {
    @apply bg-slate-800 text-white;
  }

  .primary:hover,
  .primary.selected {
    @apply bg-slate-700;
  }

  .primary:active {
    @apply bg-slate-900;
  }

  /* SECONDARY STYLES */

  .secondary {
    @apply bg-white text-slate-600;
    @apply px-3 h-7 border border-slate-300;
  }

  .secondary:hover,
  .secondary:disabled,
  .secondary.selected,
  .add:hover,
  .add:disabled,
  .add.selected {
    @apply bg-slate-100;
  }

  .secondary:active,
  .add:active {
    @apply bg-slate-200;
  }

  /* HIGHLGHTED STYLES (REMOVE) */

  .highlighted {
    @apply bg-white text-slate-700;
    @apply border border-slate-100;
    @apply shadow-md;
  }

  .highlighted:hover,
  .highlighted.selected {
    @apply bg-slate-50;
  }

  .highlighted:active {
    @apply bg-slate-200;
  }

  /* LINK STYLES */

  .link {
    @apply text-blue-500;
  }

  .link:hover,
  .link.selected {
    @apply text-blue-600;
  }

  .link:active {
    @apply text-blue-700;
  }

  .link:disabled {
    @apply text-slate-400;
  }

  /* SHAPE STYLES */

  .square,
  .circle {
    @apply p-0 aspect-square;
    @apply text-ellipsis overflow-hidden whitespace-nowrap flex-grow-0 flex-shrink-0;
  }

  .rounded,
  .circle {
    @apply rounded-full;
  }

  /* DANGER STYLES */

  .danger {
    @apply bg-red-500 text-white;
  }

  .danger:hover,
  .danger.selected {
    @apply bg-red-600;
  }

  .danger:active {
    @apply bg-red-700;
  }

  .danger.secondary {
    @apply bg-white;
    @apply text-red-500;
    @apply border-red-500;
  }

  .danger:disabled {
    @apply text-slate-400;
    @apply bg-slate-50;
    @apply border-slate-300;
  }

  /* BRAND STYLES */

  .brand {
    @apply bg-blue-600 text-white;
  }

  .brand:hover {
    @apply bg-blue-500;
  }

  .brand:active {
    @apply bg-blue-700;
  }

  /* TEXT STYLES */

  .text {
    @apply px-0 font-medium text-slate-600;
  }

  .text:hover {
    @apply text-primary-700;
  }

  .text:active {
    @apply text-primary-800;
  }

  /* TWEAKS */

  .small {
    @apply h-6 text-[11px];
  }

  .large {
    @apply h-9 text-sm;
  }

  .large.square,
  .large.circle {
    @apply h-10;
  }

  .compact {
    @apply px-2;
  }

  .no-stroke {
    @apply border-none;
  }

  .dashed {
    @apply border border-dashed;
  }

  /* ADD BUTTON STYLES */

  .add {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply border border-dashed border-slate-300;
    @apply bg-white px-0;
  }
</style>
