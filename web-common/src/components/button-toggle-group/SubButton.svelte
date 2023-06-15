<script lang="ts">
  import { getContext } from "svelte";
  import { buttonGroup } from "./ButtonToggleGroup.svelte";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let key: number | string;
  export let tootips: {
    selected?: string;
    unselected?: string;
    disabled?: string;
  };

  const {
    registerSubButton,
    toggleSubButton,
    selectedKey,
    firstKey,
    lastKey,
    disabledKeys,
  } = getContext(buttonGroup);

  registerSubButton(key);

  $: isDisabled = disabledKeys.includes(key);
  $: isSelected = $selectedKey === key;

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
  $: isFirst = key === $firstKey;
  $: isLast = key === $lastKey;
  $: roundings = `${isFirst ? "rounded-l" : ""} ${isLast ? "rounded-r" : ""} `;

  $: finalStyles = `${baseStyles} ${roundings} ${textStyle} ${bgStyle}`;

  $: tooltipText = isDisabled
    ? tootips?.disabled
    : isSelected
    ? tootips?.selected
    : tootips?.unselected;
</script>

<!-- Note: this wrapper div is needed for the styles `divide-x divide-gray-400` in the parent container to work correctly-->
<div>
  <Tooltip distance={8} location={"bottom"} alignment={"center"}>
    <button class={finalStyles} on:click={() => toggleSubButton(key)}>
      <slot />
    </button>
    <div slot="tooltip-content">
      {#if tooltipText}
        <TooltipContent>{tooltipText}</TooltipContent>
      {/if}
    </div>
  </Tooltip>
</div>
