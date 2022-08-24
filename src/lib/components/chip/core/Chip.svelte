<!-- Chips have two areas:
  = left (icon) – used primarily for icons, action buttons, and small images
  - center (text) – used primarily for label information
-->
<script>
  import { slideRight } from "$lib/transitions";
  import { createEventDispatcher } from "svelte";
  import RemoveChipButton from "./RemoveChipButton.svelte";

  export let removable = false;
  export let active = false;

  /** color elements elements */
  export let bgBaseColor = "bg-blue-50";
  export let bgHoverColor = "bg-blue-100";
  export let textColor = "text-blue-900";
  export let bgActiveColor = bgHoverColor;
  export let ringOffsetColor = "ring-offset-blue-500";

  /** if removable is true, these props control the tooltip positioning */
  export let removeButtonTooltipLocation = "bottom";
  export let removeButtonTooltipAlignment = "start";
  export let removeButtonTooltipDistance = 12;

  /** the maximum width for the tooltip of the main chip */

  const dispatch = createEventDispatcher();
</script>

<div transition:slideRight|local={{ duration: 150 }}>
  <button
    on:click
    class="
    grid gap-x-2 items-center pl-2 pr-4 py-1 rounded-2xl cursor-pointer
    {textColor}
    {bgBaseColor}
    {ringOffsetColor}
    hover:{bgHoverColor}
    {active ? bgActiveColor : ''}

  "
    class:ring-2={active}
    style:grid-template-columns="{$$slots.icon || removable
      ? "max-content"
      : ""}
    {$$slots.body ? "max-content" : ""}"
  >
    <!-- a cancelable element, e.g. filter buttons -->
    {#if removable}
      <RemoveChipButton
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
