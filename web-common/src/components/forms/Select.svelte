<script lang="ts">
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";

  export let value: string;
  export let id: string;
  export let label: string;
  export let options: { value: string; label?: string }[];

  // temporary till we figure out the menus
  export let detach = false;
  export let itemsClass = "";

  let displayValue: string;
  $: {
    const foundOption = options.find((option) => option.value === value);
    displayValue = foundOption?.label ?? value;
  }
</script>

<div class="flex flex-col gap-y-2">
  <label class="text-gray-800 text-sm font-medium" for={id}>{label}</label>
  <Menu {detach}>
    <MenuButton
      className="w-full border px-3 py-1 h-8 flex gap-x-2 justify-between items-center"
    >
      {displayValue}
      <CaretDownIcon />
    </MenuButton>
    <MenuItems positioningOverride={itemsClass}>
      {#each options as option}
        <MenuItem on:click={() => (value = option.value)}>
          {option?.label ?? option.value}
        </MenuItem>
      {/each}
    </MenuItems>
  </Menu>
</div>
