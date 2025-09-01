<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import LoadingSpinner from "../icons/LoadingSpinner.svelte";

  type ButtonType =
    | "primary"
    | "secondary"
    | "plain"
    | "subtle"
    | "ghost"
    | "dashed"
    | "link"
    | "text"
    | "add"
    | "toolbar";

  export let type: ButtonType = "plain";
  export let onClick: ((event: MouseEvent) => void) | undefined = undefined;
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
  export let gray = false;
  export let danger = false;
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
  class:gray
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
  class:danger={status === "error" || danger}
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
    @apply h-7  min-h-[28px] min-w-fit;
    @apply font-medium pointer-events-auto;
  }

  button:disabled {
    @apply cursor-not-allowed;
  }

  /* PRIMARY STYLES */

  .primary {
    @apply bg-primary-600 text-white;
  }

  .primary:hover {
    @apply bg-primary-700;
  }

  .primary:active,
  .primary.selected {
    @apply bg-primary-800;
  }

  .primary:disabled {
    @apply bg-slate-400;
  }

  .primary.theme {
    @apply bg-theme-600 text-white;
  }

  .primary.theme:hover {
    @apply bg-theme-700;
  }

  .primary.theme:active,
  .primary.theme.selected {
    @apply bg-theme-800;
  }

  /* SECONDARY, GHOST, DASHED STYLES */

  .secondary,
  .ghost,
  .dashed {
    @apply bg-transparent text-primary-600;
  }

  .secondary,
  .dashed {
    @apply border border-primary-300;
  }

  .secondary:hover,
  .ghost:hover,
  .dashed:hover {
    @apply bg-primary-50;
  }

  .secondary:active,
  .secondary.selected,
  .ghost:active,
  .ghost.selected,
  .dashed:active,
  .dashed.selected {
    @apply bg-primary-100;
  }

  .secondary.theme,
  .ghost.theme,
  .dashed.theme {
    @apply bg-transparent text-theme-600;
  }

  .secondary.theme,
  .dashed.theme {
    @apply border border-theme-300;
  }

  .secondary.theme:hover,
  .ghost.theme:hover,
  .dashed.theme:hover {
    @apply bg-theme-50;
  }

  .secondary.theme:active,
  .secondary.theme.selected,
  .ghost.theme:active,
  .ghost.theme.elected,
  .dashed.theme:active,
  .dashed.theme.selected {
    @apply bg-theme-100;
  }

  .secondary.loading,
  .ghost.loading,
  .dashed.loading {
    @apply bg-slate-50;
    @apply border-slate-300;
    @apply text-slate-600;
  }

  .secondary:disabled,
  .dashed:disabled {
    @apply text-slate-400 bg-slate-50 border-slate-300;
  }

  .ghost:disabled {
    @apply bg-transparent text-slate-400;
  }

  .secondary:active:hover,
  .secondary.selected:hover,
  .ghost:active:hover,
  .ghost.selected:hover,
  .dashed:active:hover,
  .dashed.selected:hover {
    @apply bg-primary-200;
  }

  .secondary.theme:active:hover,
  .secondary.theme.selected:hover,
  .ghost.theme:active:hover,
  .ghost.theme.selected:hover,
  .dashed.theme:active:hover,
  .dashed.theme.selected:hover {
    @apply bg-theme-200;
  }

  /* PLAIN STYLES */

  .plain {
    @apply bg-transparent text-slate-600;
    @apply border border-slate-300;
  }

  .plain:hover {
    @apply bg-slate-100;
  }

  .plain:active,
  .plain.selected {
    @apply bg-slate-200;
  }

  .plain.disabled {
    @apply text-slate-400;
  }

  /* SUBTLE STYLES */

  .subtle {
    @apply bg-primary-50 text-primary-700;
  }

  .subtle:hover {
    @apply bg-primary-100;
  }

  .subtle:active,
  .subtle.selected {
    @apply bg-primary-200 text-primary-900;
  }

  .subtle.loading {
    @apply bg-slate-50 text-slate-600;
  }

  .subtle:disabled {
    @apply text-slate-400 bg-slate-50;
  }

  /* LINK STYLES */

  .link {
    @apply text-primary-600 p-0;
  }

  .link:hover {
    @apply text-primary-700;
  }

  .link:active,
  .link.selected {
    @apply text-primary-800;
  }

  .link.loading {
    @apply text-slate-600;
  }

  .link:disabled {
    @apply text-slate-400;
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
    @apply text-gray-700 p-0;
  }

  .text:hover {
    @apply text-primary-700;
  }

  .text:active,
  .text.selected {
    @apply text-primary-800;
  }

  .text.loading {
    @apply text-slate-600;
  }

  .text:disabled {
    @apply text-slate-400;
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
    @apply font-normal text-gray-800;
    @apply h-6 px-1.5 rounded-sm;
    @apply gap-x-1.5;
  }

  .toolbar:hover {
    @apply bg-slate-600/15;
  }

  .toolbar:active,
  .toolbar.selected {
    @apply bg-slate-600/15;
  }

  .toolbar:disabled {
    @apply text-slate-400;
  }

  .text.theme:hover {
    @apply text-theme-700;
  }

  .text.theme:active,
  .text.theme.selected {
    @apply text-theme-800;
  }

  /* DANGER STYLES */

  .danger.primary {
    @apply bg-red-500 text-white;
  }

  .danger.primary:hover {
    @apply bg-red-600;
  }

  .danger.primary:active,
  .danger.selected {
    @apply bg-red-700;
  }

  .danger.primary:disabled {
    @apply bg-slate-400;
  }

  .danger.secondary {
    @apply bg-surface;
    @apply text-red-500;
    @apply border-red-500;
  }

  .danger.secondary:hover {
    @apply text-red-600;
    @apply border-red-600;
  }

  .danger.secondary:disabled {
    @apply text-slate-400;
    @apply bg-slate-50;
    @apply border-slate-300;
  }

  .danger.text {
    @apply text-slate-600 p-0;
  }

  .danger.text:hover {
    @apply text-red-600;
  }

  .danger.subtle {
    @apply bg-red-50 text-red-600;
  }

  .danger.subtle:hover {
    @apply bg-red-100;
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

  .dashed {
    @apply border border-dashed;
  }

  /* TODO: variants for types like danger */
  .active {
    @apply bg-primary-100;
  }

  /* ADD BUTTON STYLES */

  .add {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply border border-dashed border-slate-300;
    @apply bg-surface px-0;
  }

  .gray:not(:hover) {
    @apply text-slate-600 border-slate-300;
  }

  .gray:not(.ghost):not(:hover) {
    @apply bg-slate-50;
  }
</style>
