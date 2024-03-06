<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";

  export let value: string;
  export let id: string;
  export let label: string;
  export let options: { value: string; label?: string }[];
  export let placeholder: string = "";
  export let optional: boolean = false;

  // temporary till we figure out the menus
  export let detach = false;
  export let itemsClass = "";

  const dispatch = createEventDispatcher();

  let displayValue: string;
  let hasNoValue = false;
  $: {
    const foundOption = options.find((option) => option.value === value);
    displayValue = foundOption?.label ?? value;
    hasNoValue = !foundOption;
  }
</script>

<div class="flex flex-col gap-y-2">
  {#if label?.length}
    <label for={id} class="text-sm flex gap-x-1">
      <span class="text-gray-800 font-medium">
        {label}
      </span>
      {#if optional}
        <span class="text-gray-500">(optional)</span>
      {/if}
    </label>
  {/if}
  <Menu {detach}>
    <MenuButton
      className="w-full border px-3 py-1 h-8 flex gap-x-2 justify-between items-center {hasNoValue
        ? 'text-gray-400'
        : ''}"
    >
      {#if hasNoValue}
        {placeholder}
      {:else}
        {displayValue}
      {/if}
      <CaretDownIcon />
    </MenuButton>
    <MenuItems positioningOverride={itemsClass}>
      {#each options as option}
        <MenuItem
          on:click={() => {
            value = option.value;
            dispatch("change", value);
          }}
        >
          {option?.label ?? option.value}
        </MenuItem>
      {/each}
    </MenuItems>
  </Menu>
</div>
