<script lang="ts">
  import { getContext } from "svelte";
  import { buttonGroup } from "./ButtonToggleGroup.svelte";

  export let key: number | string;
  const {
    registerSubButton,
    toggleSubButton,
    selectedSubButton,
    firstSubButtonKey,
    lastSubButtonKey,
    disabledKeys,
  } = getContext(buttonGroup);

  registerSubButton(key);

  $: isDisabled = disabledKeys.includes(key);
  $: isSelected = $selectedSubButton === key;

  $: isFirst = key === $firstSubButtonKey;
  $: isLast = key === $lastSubButtonKey;

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
  $: roundings = `${isFirst ? "rounded-l" : ""} ${isLast ? "rounded-r" : ""} `;

  $: finalStyles = `${baseStyles} ${roundings} ${textStyle} ${bgStyle}`;
</script>

<button class={finalStyles} on:click={() => toggleSubButton(key)}>
  <slot />
</button>
