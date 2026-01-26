<script lang="ts" context="module">
  export type ButtonType =
    | "primary"
    | "secondary"
    | "tertiary"
    | "neutral"
    | "destructive"
    | "ghost"
    | "link"
    | "text"
    | "toolbar";
</script>

<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import LoadingSpinner from "../icons/LoadingSpinner.svelte";

  export let type: ButtonType = "tertiary";
  export let onClick: ((event: MouseEvent) => void) | undefined = undefined;
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
  export let wide = false;
  export let noStroke = false;
  export let rounded = false;
  export let href: string | null = null;
  export let rel: string | undefined = undefined;
  export let builders: Builder[] = [];
  export let loading = false;
  export let target: string | undefined = undefined;
  export let fit = false;
  export let noWrap = false;
  // export let gray = false;
  export let preload = true;
  export let active = false;
  export let loadingCopy = "Loading";
  export let theme = false;
  // needed to set certain style that could be overridden by the style block in this component
  export let forcedStyle = "";
  export let dataAttributes: Record<`data-${string}`, string> = {};

  let className: string | undefined = undefined;
  export { className as class };

  const handleClick = (event: MouseEvent) => {
    if (!disabled && onClick) {
      onClick(event);
    }
  };
</script>

<svelte:element
  this={disabled || !href ? "button" : "a"}
  role="button"
  tabindex={disabled ? -1 : 0}
  {href}
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
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  on:click={handleClick}
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
    <slot />
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
  }

  button:disabled {
    @apply cursor-not-allowed;
  }

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
    @apply bg-transparent border border-accent-primary-action text-accent-primary-action;
  }

  .secondary.theme {
    @apply border-theme-500;
  }

  .secondary:hover:not(:disabled) {
    @apply bg-surface-hover;
  }

  .secondary:disabled {
    @apply opacity-50;
  }

  /* GHOST STYLES */

  .ghost {
    @apply bg-transparent text-fg-primary;
  }

  .ghost:hover {
    @apply bg-surface-hover;
  }

  .ghost:disabled {
    @apply opacity-50;
  }

  /* TERTIARY STYLES */

  .tertiary {
    @apply bg-input text-fg-primary border;
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
    @apply text-primary-600 p-0;
  }

  .link:hover:not(:disabled) {
    @apply text-primary-700;
  }

  .link:active,
  .link.selected {
    @apply text-primary-800;
  }

  .link.loading {
    @apply text-fg-secondary;
  }

  .link:disabled {
    @apply text-fg-secondary;
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
    @apply font-normal text-fg-muted;
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
