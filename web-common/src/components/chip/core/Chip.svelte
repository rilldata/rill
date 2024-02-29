<!-- Chips have two areas:
  = left (icon) – used primarily for icons, action buttons, and small images
  - center (text) – used primarily for label information
-->
<script lang="ts">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import { createEventDispatcher } from "svelte";
  import { slideRight } from "../../../lib/transitions";
  import { defaultChipColors } from "../chip-types";
  import RemoveChipButton from "./RemoveChipButton.svelte";

  export let removable = false;
  export let active = false;
  export let outline = false;
  export let readOnly = false;
  export let grab = false;

  /** chip style props */
  export let extraRounded = true;
  export let extraPadding = true;

  /** color elements elements */
  export let bgBaseClass = defaultChipColors.bgBaseClass;
  export let bgHoverClass = defaultChipColors.bgHoverClass;
  export let textClass = defaultChipColors.textClass;
  export let bgActiveClass = defaultChipColors.bgActiveClass;
  export let outlineBaseClass = defaultChipColors.outlineBaseClass;
  export let outlineHoverClass = defaultChipColors.outlineHoverClass;
  export let outlineActiveClass = defaultChipColors.outlineActiveClass;

  /** if removable is true, these props control the tooltip positioning */
  export let supressTooltip = false;
  export let removeButtonTooltipLocation = "bottom";
  export let removeButtonTooltipAlignment = "start";
  export let removeButtonTooltipDistance = 12;

  export let label: string | undefined = undefined;

  export let builders: Builder[] = [];

  /** the maximum width for the tooltip of the main chip */

  const dispatch = createEventDispatcher();
</script>

<div in:slideRight={{ duration: 150 }}>
  {#if readOnly}
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
      <!-- body -->
      {#if $$slots.body}
        <div>
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
      <!-- a cancelable element, e.g. filter buttons -->
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
        <!-- if there is a left icon, render it here -->
        <button
          on:click|stopPropagation={() => {
            dispatch("click-icon");
          }}
        >
          <slot name="icon" />
        </button>
      {/if}
      <!-- body -->
      {#if $$slots.body}
        <button
          on:click
          on:mousedown
          class="px-2 pr-{extraPadding ? '4' : '2'} text-inherit w-full"
          class:grab
        >
          <slot name="body" />
        </button>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .grab {
    @apply cursor-grab;
  }
</style>
