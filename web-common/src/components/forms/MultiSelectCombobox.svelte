<script lang="ts">
  import { slide } from "svelte/transition";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let id = "";
  export let label = "";
  export let error: string;
  export let placeholder = "";
  export let options: string[] = [];
  export let selectedValues: string[] = [];
  export let hint = "";

  let selectedValue = "";
  let showPopover = false;

  $: filteredOptions = options.filter(
    (option) =>
      !selectedValues.includes(option) &&
      option.toLowerCase().includes(selectedValue.toLowerCase())
  );

  function toggleOption(option: string) {
    const index = selectedValues.indexOf(option);
    if (index === -1) {
      selectedValues = [...selectedValues, option];
    } else {
      selectedValues = [
        ...selectedValues.slice(0, index),
        ...selectedValues.slice(index + 1),
      ];
    }
    showPopover = false;
  }
</script>

<div class="flex flex-col gap-y-2">
  <div class="flex gap-x-1 items-center">
    <label for={id} class="text-gray-800 text-sm font-medium w-fit">
      {label}
    </label>
    {#if hint}
      <Tooltip location="right" alignment="middle" distance={8}>
        <div class="text-gray-500" style="transform:translateY(-.5px)">
          <InfoCircle size="13px" />
        </div>
        <TooltipContent maxWidth="400px" slot="tooltip-content">
          {hint}
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
  <Menu>
    <MenuButton className="w-full">
      <input
        {id}
        name={id}
        {placeholder}
        type="text"
        class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-blue-500 w-full text-xs {error &&
          'border-red-500'}"
        value={selectedValue}
        on:input={(e) => {
          selectedValue = e.target.value;
          showPopover = true;
        }}
      />
    </MenuButton>
    <MenuItems>
      {#each filteredOptions as option}
        <MenuItem on:click={() => toggleOption(option)}>
          {option}
        </MenuItem>
      {/each}
    </MenuItems>
  </Menu>
  {#if error}
    <div in:slide|local={{ duration: 200 }} class="text-red-500 text-sm py-px">
      {error}
    </div>
  {/if}
</div>
