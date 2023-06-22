<!-- Chips have two areas:
  = left (icon) – used primarily for icons, action buttons, and small images
  - center (text) – used primarily for label information
-->
<script>
  import { createEventDispatcher } from "svelte";
  import { slideRight } from "../../../lib/transitions";
  import { defaultChipColors } from "../chip-types";
  import RemoveChipButton from "./RemoveChipButton.svelte";

  export let removable = false;
  export let active = false;

  /** color elements elements */
  export let bgBaseClass = defaultChipColors.bgBaseClass;
  export let bgHoverClass = defaultChipColors.bgHoverClass;
  export let textClass = defaultChipColors.textClass;
  export let bgActiveClass = defaultChipColors.bgActiveClass;
  export let outlineClass = defaultChipColors.outlineClass;

  /** if removable is true, these props control the tooltip positioning */
  export let removeButtonTooltipLocation = "bottom";
  export let removeButtonTooltipAlignment = "start";
  export let removeButtonTooltipDistance = 12;

  export let label = undefined;

  /** the maximum width for the tooltip of the main chip */

  const dispatch = createEventDispatcher();
</script>

<div transition:slideRight|local={{ duration: 150 }}>
  <button
    on:click
    class="
    grid gap-x-2 items-center pl-2 pr-4 py-1 rounded-2xl cursor-pointer
    {textClass}
    {bgBaseClass} 
    {outlineClass} 
    {bgHoverClass} 
    {active ? bgActiveClass : ''}
  "
    class:outline-2={active}
    class:outline={active}
    style:grid-template-columns="{$$slots.icon || removable
      ? "max-content"
      : ""}
    {$$slots.body ? "max-content" : ""}"
    aria-label={label}
  >
    <!-- a cancelable element, e.g. filter buttons -->
    {#if removable}
      <RemoveChipButton
        {textClass}
        tooltipLocation={removeButtonTooltipLocation}
        tooltipAlignment={removeButtonTooltipAlignment}
        tooltipDistance={removeButtonTooltipDistance}
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
      <div>
        <slot name="body" />
      </div>
    {/if}
  </button>
</div>
