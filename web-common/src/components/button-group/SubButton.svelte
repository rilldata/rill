<script lang="ts">
  import { getContext } from "svelte";
  import { ButtonGroupContext, buttonGroupContext } from "./ButtonGroup.svelte";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let value: number | string;
  export let tooltips: {
    selected?: string;
    unselected?: string;
    disabled?: string;
  } = undefined;

  const {
    registerSubButton,
    subButtons,
    selectedKeys,
    disabledKeys,
    dispatch,
  }: ButtonGroupContext = getContext(buttonGroupContext);

  registerSubButton(value);

  $: isDisabled = $disabledKeys.includes(value);
  $: isSelected = $selectedKeys.includes(value);

  const baseStyles = `shrink flex flex-row items-center px-1 py-1
  transition-transform duration-100`;

  $: textStyle = isDisabled
    ? "text-gray-400"
    : "text-gray-700 hover:text-gray-900 ";

  $: bgStyle = isDisabled
    ? "bg-white"
    : isSelected
    ? "bg-gray-100 hover:bg-gray-200 "
    : "bg-white hover:bg-gray-50 ";

  // This is needed to make sure that the left and right most child
  // elements don't break out of the border drawn by the parent element
  $: isFirst = $subButtons.at(0) === value;
  $: isLast = $subButtons.at(-1) === value;
  $: roundings = `${isFirst ? "rounded-l" : ""} ${isLast ? "rounded-r" : ""} `;

  $: finalStyles = `${baseStyles} ${roundings} ${textStyle} ${bgStyle}`;

  $: tooltipText = isDisabled
    ? tooltips?.disabled
    : isSelected
    ? tooltips?.selected
    : tooltips?.unselected;
</script>

<!-- Note: this wrapper div is needed for the styles `divide-x divide-gray-400` in the parent container to work correctly-->
<div>
  <Tooltip distance={8} location={"bottom"} alignment={"center"}>
    <button
      class={finalStyles}
      on:click={() => {
        if (!isDisabled) {
          dispatch("subbutton-click", value);
        }
      }}
    >
      <slot />
    </button>
    <div slot="tooltip-content">
      {#if tooltipText}
        <TooltipContent>{tooltipText}</TooltipContent>
      {/if}
    </div>
  </Tooltip>
</div>
