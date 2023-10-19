<script lang="ts">
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";

  export let id = "";
  export let label = "";
  export let placeholder = "";
  export let options: string[] = [];
  export let selectedValues: string[] = [];

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
  <label for={id} class="text-gray-800 text-sm font-medium w-fit">
    {label}
  </label>
  <Menu>
    <MenuButton className="w-full">
      <input
        {id}
        name={id}
        {placeholder}
        type="text"
        class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-blue-500 w-full text-xs"
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
</div>
