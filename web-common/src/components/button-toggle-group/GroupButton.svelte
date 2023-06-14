<script lang="ts">
  import { getContext } from "svelte";
  import { buttonGroup } from "./ButtonToggleGroup.svelte";

  export let key: number | string;
  const {
    registerSubButton,
    selectSubButton,
    selectedSubButton,
    firstSubButtonKey,
    lastSubButtonKey,
  } = getContext(buttonGroup);

  registerSubButton(key);

  $: isFirst = key === $firstSubButtonKey;
  $: isLast = key === $lastSubButtonKey;

  $: console.log("isFirst", isFirst, key, $firstSubButtonKey);

  $: console.log("isLast", isLast, key, $lastSubButtonKey);

  // const roundings = {
  //   first: "rounded-l",
  //   last: "rounded-r",
  //   middle: "",
  // };

  const baseStyles = ` flex flex-row items-center px-2 py-1
  transition-transform duration-100
  focus:outline-none focus:ring-2 
  border border-gray-400
  hover:bg-gray-50 hover:text-gray-600 hover:border-gray-500 focus:ring-blue-300 `;

  $: selectionStyles =
    $selectedSubButton === key
      ? "bg-gray-200 text-gray-900"
      : "bg-gray-300 text-gray-900";

  $: roundings = `${isFirst ? "rounded-l" : ""} ${isLast ? "rounded-r" : ""} `;

  $: finalStyles = `${baseStyles} ${roundings}`;

  // function buttonClasses({ compact = true, disabled = false }) {
  //   const padding = compact ? "px-1 py-1" : "px-3 py-1";
  //   if (disabled) {
  //     return `
  // ${padding} rounded flex flex-row items-center transition-transform duration-100
  // focus:outline-none focus:ring-2 text-gray-500 border border-gray-400 hover:bg-gray-200 hover:text-gray-600 hover:border-gray-500 focus:ring-blue-300`;
  //   }
  //   return `
  // ${padding} rounded flex flex-row items-center transition-transform duration-100
  // focus:outline-none focus:ring-2 text-gray-500 border border-gray-400 hover:bg-gray-200 hover:text-gray-600 hover:border-gray-500 focus:ring-blue-300
  // `;
  // }
</script>

<button
  class:selected={$selectedSubButton === key}
  class={finalStyles}
  on:click={() => selectSubButton(key)}
>
  <slot />
</button>

<style>
  .selected {
    /* border-bottom: 2px solid teal; */
    color: #333;
  }
</style>
