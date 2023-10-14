<script lang="ts">
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";

  export let value: string;
  export let id: string;
  export let label: string;
  export let options: string[];
</script>

<!-- horizontal: -->
<!-- <div class="flex gap-x-2"> -->
<!-- <label for={id} class="text-gray-800 text-sm">{label}:</label> -->
<div>
  <label for={id} class="text-gray-600">{label}</label>
  <Menu>
    <MenuButton
      className="w-full border px-3 py-1 h-8 flex justify-between items-center"
    >
      {value}
      <CaretDownIcon />
    </MenuButton>
    <MenuItems>
      {#each options as option}
        <MenuItem on:click={() => (value = option)}>{option}</MenuItem>
      {/each}
    </MenuItems>
  </Menu>
</div>
