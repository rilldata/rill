<!-- Chips have two areas:
  = left (icon) – used primarily for icons, action buttons, and small images
  - center (text) – used primarily for label information
-->
<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { createEventDispatcher, getContext } from "svelte";
  import { slideRight } from "../../../lib/transitions";

  import { get, Writable } from "svelte/store";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import CancelCircle from "../../icons/CancelCircle.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";

  export let removable = false;
  export let active = false;
  export let outline = false;
  export let readOnly = false;
  export let type: "measure" | "dimension" | "time";
  export let exclude = false;
  export let slideDuration = 150;

  /** chip style props */
  // export let extraRounded = true;
  // export let extraPadding = true;

  /** color elements elements */
  // export let bgBaseClass = defaultChipColors.bgBaseClass;
  // export let bgHoverClass = defaultChipColors.bgHoverClass;
  export let textClass = "text-black";
  // export let bgActiveClass = defaultChipColors.bgActiveClass;
  // export let outlineBaseClass = defaultChipColors.outlineBaseClass;
  // export let outlineHoverClass = defaultChipColors.outlineHoverClass;
  // export let outlineActiveClass = defaultChipColors.outlineActiveClass;

  /** if removable is true, these props control the tooltip positioning */
  export let supressTooltip = false;
  export let removeButtonTooltipLocation = "bottom";
  export let removeButtonTooltipAlignment = "start";
  export let removeButtonTooltipDistance = 12;

  export let label: string | undefined = undefined;

  export let builders: Builder[] = [];

  /** the maximum width for the tooltip of the main chip */

  const dispatch = createEventDispatcher();

  const tooltipSuppression = getContext(
    "rill:app:childRequestedTooltipSuppression",
  ) as Writable<boolean>;

  function focusOnRemove() {
    if (tooltipSuppression) tooltipSuppression.set(true);
  }
  function blurOnRemove() {
    if (tooltipSuppression) tooltipSuppression.set(false);
  }
</script>

<div in:slideRight={{ duration: slideDuration }}>
  <div class="chip {type}" aria-label={label}>
    {#if removable}
      <Tooltip
        location={removeButtonTooltipLocation}
        alignment={removeButtonTooltipAlignment}
        distance={removeButtonTooltipDistance}
        suppress={supressTooltip}
      >
        <button
          class="{textClass} pl-2"
          on:mouseover={focusOnRemove}
          on:focus={focusOnRemove}
          on:mouseleave={blurOnRemove}
          on:blur={blurOnRemove}
          on:click|stopPropagation={() => dispatch("remove")}
          aria-label="Remove"
        >
          <CancelCircle size="16px" />
        </button>
        <div slot="tooltip-content">
          {#if $$slots["remove-tooltip"]}
            <TooltipContent maxWidth="300px">
              <slot name="remove-tooltip" />
            </TooltipContent>
          {/if}
        </div>
      </Tooltip>
    {/if}

    {#if $$slots.body}
      <button
        class="px-2 text-inherit w-full select-none"
        {...getAttrs(builders)}
        use:builderActions={{ builders }}
      >
        <slot name="body" />
      </button>
    {/if}
  </div>

  <!-- {#if readOnly}
    <div
      class="
      grid gap-x-2 items-center
      py-1 {extraRounded ? 'rounded-2xl' : 'rounded-sm'}
      {textClass}
      {active ? bgActiveClass : bgBaseClass}
      {outline ? outlineBaseClass : ''}
      {active && outline ? outlineActiveClass : ''} 
      "
      style:grid-template-columns="{$$slots.icon || removable
        ? "max-content"
        : ""}
      {$$slots.body ? "max-content" : ""}"
      aria-label={label}
    >
   
      {#if $$slots.body}
        <div class="px-2 text-inherit w-full select-none">
          <slot name="body" />
        </div>
      {/if}
    </div>
  {:else}
    <div
      class="w-fit
    grid items-center
      py-1 {extraRounded ? 'rounded-2xl' : 'rounded-sm'} cursor-pointer
    {textClass}
    {bgHoverClass} 
    {active ? bgActiveClass : bgBaseClass}
    {outline ? outlineBaseClass + ' ' + outlineHoverClass : ''}
    {active && outline ? outlineActiveClass : ''} 
  "
      style:grid-template-columns="{$$slots.icon || removable
        ? "max-content"
        : ""}
      {$$slots.body ? "max-content" : ""}"
      aria-label={label}
      {...getAttrs(builders)}
      use:builderActions={{ builders }}
    >
     
      {#if removable}
        <RemoveChipButton
          {textClass}
          tooltipLocation={removeButtonTooltipLocation}
          tooltipAlignment={removeButtonTooltipAlignment}
          tooltipDistance={removeButtonTooltipDistance}
          {supressTooltip}
          on:remove
        >
          <svelte:fragment slot="remove-tooltip">
            {#if $$slots["remove-tooltip"]}
              <slot name="remove-tooltip" />
            {/if}
          </svelte:fragment>
        </RemoveChipButton>
      {:else if $$slots.icon}
      
        <button
          on:click|stopPropagation={() => {
            dispatch("click-icon");
          }}
        >
          <slot name="icon" />
        </button>
      {/if}
    
      {#if $$slots.body}
        <button
          on:click
          on:mousedown
          class="px-2 pr-{extraPadding
            ? '4'
            : '2'} text-inherit w-full select-none"
          class:grab
          aria-label={label}
        >
          <slot name="body" />
        </button>
      {/if}
    </div>
  {/if} -->
</div>

<!-- bgBaseClass: "bg-secondary-50 dark:bg-secondary-600",
bgHoverClass: "hover:bg-secondary-100 hover:dark:bg-secondary-800",
bgActiveClass: "bg-secondary-100 dark:bg-secondary-600",
outlineBaseClass:
  "outline outline-1 outline-secondary-200 dark:outline-secondary-500",
outlineHoverClass: "hover:outline-secondary-300",
outlineActiveClass: "!outline-secondary-500 dark:outline-secondary-500",
textClass: "text-secondary-800", -->

<!-- bgBaseClass: "bg-primary-50 dark:bg-primary-600",
bgHoverClass: "hover:bg-primary-100 hover:dark:bg-primary-800",
bgActiveClass: "bg-primary-100 dark:bg-primary-700",
outlineBaseClass:
  "outline outline-1 outline-primary-100 dark:outline-primary-500",
outlineHoverClass: "hover:outline-primary-200",
outlineActiveClass: "!outline-primary-500 dark:outline-primary-500",
textClass: "text-primary-800 dark:text-primary-50", -->

<style lang="postcss">
  .grab {
    @apply cursor-grab;
  }

  .chip {
    @apply flex flex-none gap-x-2;
    @apply items-center justify-center;
    @apply px-2 py-1 border;
  }

  .dimension {
    @apply rounded-2xl;
    @apply bg-primary-50;
    @apply border-primary-100;
    @apply text-primary-800;
  }

  .dimension:hover {
    @apply bg-primary-100;
  }

  .measure {
    @apply rounded-sm;
    @apply bg-secondary-50;
    @apply border-secondary-200;
  }
</style>
