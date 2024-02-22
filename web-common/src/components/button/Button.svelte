<script context="module" lang="ts">
  /**
   * Button types from the design system
   * https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0
   *
   * Note that "brand", "highlighted" and "add" are not
   * included in the new design system
   */
  export type ButtonKind =
    | "primary"
    | "secondary"
    | "subtle"
    | "ghost"
    | "dashed"
    | "link"
    | "text"
    // NOTE: these are not included the new design system, see:
    // https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0
    | "brand"
    | "add"
    | "highlighted";

  /**
   * Button shapes from the design system
   * https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0
   */
  export type ButtonShape = "normal" | "square" | "circle";

  /**
   * Button sizes from the design system
   * https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0
   */
  export type ButtonSize = "small" | "medium" | "large" | "xl";
</script>

<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { createEventDispatcher } from "svelte";
  export let type: ButtonKind = "primary";
  export let status: "info" | "error" = "info";
  export let disabled = false;
  export let compact = false;
  export let submitForm = false;
  export let active = false;

  /**
   * Note: "loading" is equivalent to "Status" in the design system
   * because the "status" prop is already used for "info" and "error" states.
   * https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0
   */
  export let loading = true;
  export let form = "";
  export let label: string | undefined = undefined;
  export let shape: ButtonShape = "normal";
  export let selected = false;
  export let size: ButtonSize = "medium";
  export let rounded = false;
  export let builders: Builder[] = [];

  $: circle = shape === "circle";
  $: square = shape === "square";

  // $: small = size === "small";
  // $: large = size === "large";

  $: finalType = type;

  let dashed = false;

  // Dashed buttons are just secondary buttons with a dashed border
  $: if (type === "dashed") {
    finalType = "secondary";
    dashed = true;
  }

  $: danger = status === "error";

  $: if ((type === "ghost" || type === "subtle") && danger) {
    console.warn(
      `Button cannot be both type==="${type}" and "dangerous", falling back to type="secondary". See https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0`,
    );
    finalType = "secondary";
  }

  $: if (
    (type === "link" || type === "text") &&
    (shape === "circle" || shape === "square")
  ) {
    console.warn(
      `Button cannot be both type==="${type}" and shape==="${shape}", falling back to type="ghost". See https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0`,
    );
    finalType = "ghost";
  }

  $: if (size === "xl" && danger) {
    console.warn(
      `"Dangerous" buttons should not be of size "XL". See https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=74-246&mode=design&t=XqvJDtq7QqQfNKE5-0`,
    );
  }

  $: console.log({ type, finalType, dashed });

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };
</script>

<button
  class="{$$props.class} {finalType} {size}"
  {disabled}
  class:square
  class:circle
  class:selected
  class:dashed
  class:compact
  class:rounded
  class:danger
  class:loading
  class:active
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
    @apply cursor-pointer;
  }

  button:disabled {
    @apply cursor-not-allowed;
  }

  /* BRAND STYLES - REMOVED IN DESIGN SYSTEM */

  .brand {
    @apply bg-primary-600 text-white;

    &:hover,
    &.selected {
      @apply bg-primary-700;
    }
    &:active,
    &.active {
      @apply bg-primary-800;
    }
    &.loading {
      @apply bg-primary-600;
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
      @apply bg-primary-700;
    }
    &:active,
    &.active {
      @apply bg-primary-800;
    }
    &.loading {
      @apply bg-slate-600;
    }
    &:disabled {
      @apply bg-slate-400;
    }
  }

  /* SECONDARY STYLES */

  .secondary {
    @apply bg-white text-primary-600;
    @apply border border-primary-300;

    &:hover,
    &.selected {
      @apply bg-primary-50;
    }
    &:active,
    &.active {
      @apply bg-primary-100;
    }
    &.loading {
      @apply bg-slate-50;
      @apply text-slate-600;
      @apply border-slate-300;
    }
    &:disabled {
      @apply bg-slate-50;
      @apply text-slate-400;
      @apply border-slate-300;
    }
  }

  /* SUBTLE STYLES */
  .subtle {
    @apply bg-primary-50 text-primary-700;

    &:hover,
    &.selected {
      @apply bg-primary-100;
    }
    &:active,
    &.active {
      @apply bg-primary-200;
    }
    &.loading {
      @apply bg-slate-50;
      @apply text-slate-600;
    }
    &:disabled {
      @apply bg-slate-50;
      @apply text-slate-400;
    }
  }

  /* GHOST STYLES */
  .ghost {
    @apply text-primary-600;

    &:hover,
    &.selected {
      @apply bg-primary-50;
    }
    &:active,
    &.active {
      @apply bg-primary-100;
    }
    &.loading {
      @apply bg-slate-50;
      @apply text-slate-600;
    }
    &:disabled {
      @apply text-slate-400;
    }
  }

  /* DASHED STYLES -- note: dashed is just "secondary" with a dashed border */
  .dashed {
    @apply border-dashed;
  }

  /* LINK STYLES */
  .link {
    @apply text-primary-600;

    &:hover,
    &.selected {
      @apply text-primary-700;
    }
    &:active,
    &.active {
      @apply text-primary-800;
    }
    &.loading {
      @apply text-slate-600;
    }
    &:disabled {
      @apply text-slate-400;
    }
  }

  /* TEXT STYLES */
  .text {
    @apply text-slate-600;

    &:hover,
    &.selected {
      @apply text-primary-700;
    }
    &:active,
    &.active {
      @apply text-primary-800;
    }
    &.loading {
      @apply text-slate-600;
    }
    &:disabled {
      @apply text-slate-400;
    }
  }

  /* DANGER STYLES */

  .danger {
    @apply bg-red-500 text-white;
    &:hover,
    &.selected {
      @apply bg-red-600;
    }
    &:active {
      @apply bg-red-700;
    }

    &:disabled {
      @apply text-slate-400;
      @apply bg-slate-50;
      @apply border-slate-300;
    }
  }

  .danger.secondary {
    @apply bg-white;
    @apply text-red-500;
    @apply border-red-500;
  }

  /* ADD BUTTON STYLES */

  .add {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply border border-dashed border-slate-300;
    @apply bg-white px-0;

    &:hover,
    &.selected {
      @apply bg-slate-100;
    }
    &:active,
    &.active {
      @apply bg-slate-200;
    }
    &.loading {
      @apply bg-primary-600;
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
    @apply h-6 text-[11px] min-h-[24px];
    @apply px-2;
  }

  .medium {
    @apply px-3 h-7 min-h-[28px];
  }

  .large {
    @apply h-9 text-sm;
    @apply px-3;
  }

  .xl {
    @apply h-10 text-sm;
    @apply px-4;
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
</style>
