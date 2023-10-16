<script lang="ts">
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";
  import { formatTime, getNextQuarterHour } from "./time-utils";

  export let value: string;

  const start = getNextQuarterHour();
  const options = Array.from({ length: 24 * 4 }, (_, i) => {
    const nextTime = new Date(start.getTime() + i * 15 * 60000);
    return formatTime(nextTime);
  });
</script>

<Menu>
  <MenuButton
    className="w-full border px-3 py-1 h-8 flex gap-x-2 justify-between items-center"
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
