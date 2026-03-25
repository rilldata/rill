<script lang="ts">
  import type { Snippet } from "svelte";
  import LoadingSpinner from "../icons/LoadingSpinner.svelte";
  import type { ButtonType } from "./types";

  // svelte-ignore custom_element_props_identifier
  let {
    type = "tertiary" as ButtonType,
    onClick,
    disabled = false,
    compact = false,
    submitForm = false,
    form = "",
    label = null as string | null | undefined,
    square = false,
    circle = false,
    selected = false,
    large = false,
    small = false,
    wide = false,
    noStroke = false,
    rounded = false,
    href = null as string | null,
    rel = undefined as string | undefined,
    loading = false,
    target = undefined as string | undefined,
    fit = false,
    noWrap = false,
    preload = true,
    active = false,
    loadingCopy = "Loading",
    theme = false,
    forcedStyle = "",
    dataAttributes = {} as Record<`data-${string}`, string>,
    class: className,
    children,
    ...restProps
  }: {
    type?: ButtonType;
    onClick?: (event: MouseEvent) => void;
    disabled?: boolean;
    compact?: boolean;
    submitForm?: boolean;
    form?: string;
    label?: string | null;
    square?: boolean;
    circle?: boolean;
    selected?: boolean;
    large?: boolean;
    small?: boolean;
    wide?: boolean;
    noStroke?: boolean;
    rounded?: boolean;
    href?: string | null;
    rel?: string;
    loading?: boolean;
    target?: string;
    fit?: boolean;
    noWrap?: boolean;
    preload?: boolean;
    active?: boolean;
    loadingCopy?: string;
    theme?: boolean;
    forcedStyle?: string;
    dataAttributes?: Record<`data-${string}`, string>;
    class?: string;
    children?: Snippet;
    [key: string]: unknown;
  } = $props();

  function handleClick(event: MouseEvent) {
    // Forward to rest onclick (e.g., bits-ui trigger handler)
    if (typeof restProps.onclick === "function") {
      (restProps.onclick as (e: MouseEvent) => void)(event);
    }
    // Call component's own handler
    if (!disabled && onClick) {
      onClick(event);
    }
  }
</script>

<svelte:element
  this={disabled || !href ? "button" : "a"}
  role="button"
  tabindex={disabled ? -1 : 0}
  {href}
  {...restProps}
  class="{className} {type}"
  {disabled}
  class:square
  class:circle
  class:selected
  class:loading
  class:large
  class:small
  class:theme
  class:wide
  class:compact
  class:rounded
  class:active
  class:!w-fit={fit}
  class:whitespace-nowrap={noWrap}
  class:no-stroke={noStroke}
  type={submitForm ? "submit" : "button"}
  form={submitForm ? form : undefined}
  aria-label={label}
  {target}
  aria-disabled={disabled}
  rel={target === "_blank" ? "noopener noreferrer" : rel}
  onclick={handleClick}
  style={forcedStyle}
  {...href ? { "data-sveltekit-preload-data": preload ? "hover" : "off" } : {}}
  {...dataAttributes}
>
  {#if loading}
    <LoadingSpinner size="15px" />
    {#if !square && !circle && !compact}
      <span>{loadingCopy}</span>
    {/if}
  {:else}
    {@render children?.()}
  {/if}
</svelte:element>

<style lang="postcss">
  button,
  a {
    @apply flex flex-none text-center items-center justify-center;
    @apply text-xs leading-snug font-normal;
    @apply select-none  cursor-pointer;
    @apply rounded-[2px];
    @apply px-3 gap-x-2;
    @apply h-7 min-h-[28px] min-w-fit;
    @apply font-medium pointer-events-auto;
    --focus-color: var(--fg-inverse);
  }

  button:disabled {
    @apply cursor-not-allowed;
  }

  /* button:focus {
    @apply outline-none;
    box-shadow: 0px 0px 4px 1px
      color-mix(in oklab, var(--focus-color) 50%, transparent);
  } */

  /* PRIMARY STYLES */

  .primary {
    @apply bg-accent-primary text-fg-inverse;
  }

  .primary:hover:not(:disabled) {
    @apply opacity-80;
  }

  /* .primary:active,
  .primary.selected {
    @apply bg-primary-800;
  } */

  .primary:disabled {
    @apply opacity-50;
  }

  .primary.theme {
    @apply bg-theme-500 text-fg-inverse;
  }

  /* SECONDARY STYLES */

  .secondary {
    --focus-color: var(--color-primary-600);
    @apply bg-transparent border border-accent-primary-action text-accent-primary-action;
  }

  :global(.dark) .secondary {
    @apply bg-transparent;
  }

  .secondary.theme {
    --focus-color: var(--color-theme-600);
    @apply border-theme-500;
  }

  .secondary:hover:not(:disabled) {
    @apply bg-surface-hover text-fg-accent;
  }

  .secondary:disabled {
    @apply opacity-50;
  }

  /* SECONDARY DESTRUCTIVE STYLES */

  .secondary-destructive {
    --focus-color: var(--color-red-600);
    @apply bg-transparent border text-red-600;
    border-color: var(--color-red-400);
  }

  .secondary-destructive:hover:not(:disabled) {
    @apply bg-red-50 text-red-700;
  }

  :global(.dark) .secondary-destructive:hover:not(:disabled) {
    @apply bg-red-950;
  }

  .secondary-destructive:disabled {
    @apply opacity-50;
  }

  /* GHOST STYLES */

  .ghost {
    @apply bg-transparent text-fg-primary;
  }

  .ghost:hover {
    @apply bg-surface-hover;
  }

  .ghost.selected {
    @apply bg-primary-100;
  }

  .ghost:disabled {
    @apply opacity-50;
  }

  /* TERTIARY STYLES */

  .tertiary {
    --focus-color: var(--color-primary-600);
    @apply bg-input text-fg-primary border;
  }

  .tertiary.theme {
    --focus-color: var(--color-theme-600);
  }

  .tertiary:hover:not(:disabled) {
    @apply bg-surface-hover;
  }

  .tertiary.disabled {
    @apply opacity-50;
  }

  /* NEUTRAL STYLES */

  .neutral {
    @apply bg-surface-muted text-fg-secondary;
  }

  .neutral:hover:not(:disabled) {
    @apply opacity-80;
  }

  .neutral.disabled {
    @apply opacity-50;
  }

  /* LINK STYLES */

  .link {
    @apply text-accent-primary-action p-0;
  }

  .link:hover:not(:disabled) {
    @apply underline;
  }

  .link.loading {
    @apply text-fg-secondary;
  }

  .link:disabled {
    @apply opacity-50;
  }

  .link.theme {
    @apply text-theme-600 p-0;
  }

  .link.theme:hover {
    @apply text-theme-700;
  }

  .link.theme:active,
  .link.theme.selected {
    @apply text-theme-800;
  }

  /* TEXT STYLES */

  .text {
    @apply text-fg-muted p-0;
  }

  .text:hover {
    @apply text-primary-700;
  }

  .text:active,
  .text.selected {
    @apply text-primary-800;
  }

  .text.loading {
    @apply text-fg-secondary;
  }

  .text:disabled {
    @apply text-fg-secondary/30;
  }

  .text.theme:hover {
    @apply text-green-700;
  }

  .text.theme:active,
  .text.theme.selected {
    @apply text-theme-800;
  }

  /* TOOLBAR STYLES */

  .toolbar {
    @apply font-normal text-fg-secondary;
    @apply h-6 px-1.5 rounded-sm;
    @apply gap-x-1.5;
  }

  .toolbar:hover:not(:disabled) {
    @apply bg-gray-600/15;
  }

  .toolbar:active,
  .toolbar.selected {
    @apply bg-gray-600/15;
  }

  .toolbar:disabled {
    @apply opacity-50;
  }

  .text.theme:hover {
    @apply text-theme-700;
  }

  .text.theme:active,
  .text.theme.selected {
    @apply text-theme-800;
  }

  /* DESTRUCTIVE STYLES */

  .destructive {
    @apply bg-destructive text-destructive-foreground;
    --focus-color: var(--color-destructive-600);
  }

  :global(.dark) .destructive {
    @apply bg-destructive/65;
  }

  .destructive:hover {
    @apply bg-destructive;
  }

  .destructive:disabled {
    @apply opacity-50;
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

  /* TWEAKS */

  .small {
    @apply text-[11px] h-6 min-h-6;
  }

  .large {
    @apply h-9 text-sm;
  }

  .large.square,
  .large.circle {
    @apply h-10;
  }

  .wide {
    @apply w-full max-w-[400px];
    @apply h-10 text-sm;
  }

  .compact {
    @apply px-2;
  }

  .no-stroke {
    @apply border-none;
  }

  .gray:not(:hover) {
    @apply text-fg-secondary border-gray-300;
  }

  .gray:not(.ghost):not(:hover) {
    @apply bg-surface-background;
  }
</style>
