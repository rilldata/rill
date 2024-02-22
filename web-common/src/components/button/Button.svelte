<script context="module" lang="ts">
  export type ButtonType =
    | "brand"
    | "primary"
    | "secondary"
    | "noStroke"
    | "dashed"
    | "link"
    | "text"
    | "add"
    // NOTE: this is deprecated in the new design system
    | "highlighted";

  export type ButtonShape = "normal" | "square" | "circle";
  export type ButtonSize = "medium" | "large" | "small";
</script>

<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { createEventDispatcher } from "svelte";
  export let type: ButtonType = "primary";
  export let status: "info" | "error" = "info";
  export let disabled = false;
  export let compact = false;
  export let submitForm = false;
  export let form = "";
  export let label: string | undefined = undefined;
  export let shape: ButtonShape = "normal";
  export let selected = false;
  export let size: ButtonSize = "medium";
  export let rounded = false;
  export let builders: Builder[] = [];

  $: circle = shape === "circle";
  $: square = shape === "square";

  $: small = size === "small";
  $: large = size === "large";

  $: noStroke = type === "noStroke";
  $: dashed = type === "dashed";

  $: danger = status === "error";

  if (noStroke && danger) {
    console.warn(
      `Button cannot be both "No Stroke" and "dangerous", falling back to "Text" and "dangerous"`,
    );
  }

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };
</script>

<button
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
</button>

<style lang="postcss">
  button {
    @apply flex text-center items-center justify-center;
    @apply text-ellipsis overflow-hidden whitespace-nowrap;
    @apply text-xs leading-snug font-normal;
    @apply gap-x-2 min-w-fit;
    @apply rounded-[2px];
    @apply px-3 h-7 min-h-[28px] cursor-pointer;
  }

  button:disabled {
    @apply cursor-not-allowed;
  }

  /* BRAND STYLES */

  .brand {
    @apply bg-primary-600 text-white;

    &:hover,
    &.selected {
      @apply bg-primary-700;
    }
    &:active {
      @apply bg-primary-800;
    }
    &:disabled {
      @apply bg-primary-600;
      @apply opacity-50;
    }
  }

  /* PRIMARY STYLES */

  .primary {
    @apply bg-primary-600 text-white;

    &:hover,
    &.selected {
      @apply bg-slate-800;
    }
    &:active {
      @apply bg-slate-900;
    }
    &:disabled {
      @apply bg-slate-700;
      @apply opacity-50;
    }
  }

  /* SECONDARY STYLES */

  .secondary,
  .add {
    @apply bg-white text-slate-600;
    @apply px-3 h-7 border border-slate-300;

    &:hover,
    &.selected {
      @apply bg-slate-100;
    }
    &:active {
      @apply bg-slate-200;
    }
    &:disabled {
      @apply text-slate-400;
      @apply bg-slate-50;
    }
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
    @apply text-primary-600;

    &:hover,
    &.selected {
      @apply text-primary-800;
    }
    &:active {
      @apply text-primary-700;
    }
    &:disabled {
      @apply text-primary-300;
    }
  }

  /* SHAPE STYLES */

  .square,
  .circle {
    @apply p-0 aspect-square;
    @apply flex-grow-0 flex-shrink-0;
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
