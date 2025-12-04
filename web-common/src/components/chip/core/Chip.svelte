<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { slideRight } from "../../../lib/transitions";
  import CancelCircle from "../../icons/CancelCircle.svelte";
  import CaretDownIcon from "../../icons/CaretDownIcon.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";

  export let removable = false;
  export let active = false;
  export let readOnly = false;
  export let type: "measure" | "dimension" | "time" | "special" | "amber" =
    "dimension";
  export let exclude = false;
  export let grab = false;
  export let compact = false;
  export let fullWidth = false;
  export let builders: Builder[] = [];
  export let caret = builders.length > 0;
  export let slideDuration = 150;
  export let supressTooltip = false;
  export let label: string | undefined = undefined;
  export let removeTooltipText: string | undefined = undefined;
  export let allowPointerEvents = false;
  export let theme = false;
  export let onRemove: () => void = () => {};

  const tooltipSuppression = getContext<Writable<boolean>>(
    "rill:app:childRequestedTooltipSuppression",
  );

  function focusOnRemove() {
    if ($tooltipSuppression) tooltipSuppression.set(true);
  }
  function blurOnRemove() {
    if ($tooltipSuppression) tooltipSuppression.set(false);
  }
</script>

<div in:slideRight={{ duration: slideDuration }}>
  <div
    class="chip {type}"
    class:theme
    class:active
    class:grab
    class:exclude
    class:compact
    class:fullWidth
    class:pointer-events-none={readOnly && !allowPointerEvents}
    {...getAttrs(builders)}
    use:builderActions={{ builders }}
    aria-label={label}
  >
    {#if removable && !readOnly}
      <Tooltip
        alignment="start"
        distance={12}
        suppress={supressTooltip || !removeTooltipText}
      >
        <button
          class="text-inherit mr-0.5"
          aria-label="Remove"
          on:mouseover={focusOnRemove}
          on:focus={focusOnRemove}
          on:mouseleave={blurOnRemove}
          on:blur={blurOnRemove}
          on:click|stopPropagation={onRemove}
          type="button"
        >
          <CancelCircle size="16px" />
        </button>

        <TooltipContent maxWidth="300px" slot="tooltip-content">
          {removeTooltipText}
        </TooltipContent>
      </Tooltip>
    {/if}

    {#if $$slots.body}
      <button
        on:click
        on:mousedown
        aria-label={`Open ${label}`}
        class="text-inherit w-full select-none flex items-center justify-between gap-x-1 px-0.5"
        type="button"
      >
        <slot name="body" />

        {#if caret}
          <span class="transition-transform -mr-0.5" class:-rotate-180={active}>
            <CaretDownIcon size="10px" />
          </span>
        {/if}
      </button>
    {/if}
  </div>
</div>

<style lang="postcss">
  .grab {
    @apply cursor-grab;
  }

  .chip {
    @apply flex flex-none gap-x-1;
    @apply items-center justify-center;
    @apply px-2 py-[3px] border w-fit;
  }

  .dimension {
    @apply rounded-2xl;
    @apply bg-primary-50 border-primary-200 text-primary-800;
  }

  .dimension:hover,
  .dimension:active,
  .dimension.active {
    @apply bg-primary-100;
  }

  .dimension:hover,
  .dimension:active,
  .dimension.active {
    @apply bg-primary-100;
  }

  .dimension.theme {
    @apply bg-theme-50 border-theme-200 text-theme-800;
  }

  .dimension.theme:active,
  .dimension.theme.active {
    @apply border-theme-400;
  }

  .dimension.theme:active,
  .dimension.theme.active {
    @apply border-theme-400;
  }

  .measure {
    @apply rounded-sm;
    @apply bg-secondary-50 border-secondary-200 text-secondary-800;
  }

  .measure:hover,
  .measure:active,
  .measure.active {
    @apply bg-secondary-100;
  }

  .measure:active,
  .measure.active {
    @apply border-secondary-400;
  }

  .measure.theme {
    @apply rounded-sm;
    @apply bg-theme-secondary-50 border-theme-secondary-200 text-theme-secondary-800;
  }

  .measure.theme:hover,
  .measure.theme:active,
  .measure.theme.active {
    @apply bg-theme-secondary-100;
  }

  .measure.theme:active,
  .measure.theme.active {
    @apply border-theme-secondary-400;
  }

  .exclude {
    @apply bg-gray-50 text-gray-600;
  }

  .exclude:hover,
  .exclude:active,
  .exclude.active {
    @apply bg-gray-100;
  }

  .exclude:active,
  .exclude.active {
    @apply border-gray-400;
  }

  .time {
    @apply rounded-2xl;
    @apply bg-surface border-slate-200 text-slate-800;
  }

  .time:hover,
  .time.active,
  .time:active {
    @apply bg-slate-50;
  }

  .time.active,
  .time:active {
    @apply border-slate-400;
  }

  .amber {
    @apply rounded-2xl h-[18px] text-xs;
    @apply bg-amber-50 border-amber-300 text-amber-600;
    @apply font-normal;
  }

  .amber:hover,
  .amber:active,
  .amber.active {
    @apply bg-amber-100;
  }

  .compact {
    @apply py-0;
  }

  .fullWidth {
    @apply w-full;
  }
</style>
