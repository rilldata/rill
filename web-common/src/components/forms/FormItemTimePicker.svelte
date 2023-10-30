<script lang="ts">
  import { onMount } from "svelte";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import Menu from "../menu-v2/Menu.svelte";
  import MenuButton from "../menu-v2/MenuButton.svelte";
  import MenuItem from "../menu-v2/MenuItem.svelte";
  import MenuItems from "../menu-v2/MenuItems.svelte";

  export let value: string;
  export let id: string;
  export let label: string;

  interface TimeOption {
    time12Hour: string;
    time24Hour: string;
  }
  let timeOptions: TimeOption[] = [];

  onMount(() => {
    let currentDate = new Date();
    let currentMinutes = currentDate.getMinutes();
    let nextRoundedMinutes = Math.ceil(currentMinutes / 15) * 15;

    currentDate.setMinutes(nextRoundedMinutes);
    currentDate.setSeconds(0);

    for (let i = 0; i < 96; i++) {
      let hours = currentDate.getHours();
      let minutes = currentDate.getMinutes();

      let time12Hour = `${hours % 12 || 12}:${minutes === 0 ? "00" : minutes}${
        hours >= 12 ? "pm" : "am"
      }`;
      let time24Hour = `${hours}:${minutes === 0 ? "00" : minutes}`;

      timeOptions.push({ time12Hour: time12Hour, time24Hour: time24Hour });
      currentDate.setMinutes(currentDate.getMinutes() + 15);
    }
  });

  function get12HourTimeFrom24HourTime(time24Hour: string): string {
    let hours = parseInt(time24Hour.split(":")[0]);
    let minutes = parseInt(time24Hour.split(":")[1]);

    let time12Hour = `${hours % 12 || 12}:${minutes === 0 ? "00" : minutes}${
      hours >= 12 ? "pm" : "am"
    }`;

    return time12Hour;
  }
</script>

<div>
  <label for={id} class="text-gray-600">{label ?? ""}</label>
  <Menu>
    <MenuButton
      className="w-full border px-3 py-1 h-8 flex gap-x-2 justify-between items-center"
    >
      {get12HourTimeFrom24HourTime(value)}
      <CaretDownIcon />
    </MenuButton>
    <MenuItems>
      {#each timeOptions as option}
        <MenuItem on:click={() => (value = option.time24Hour)}
          >{option.time12Hour}</MenuItem
        >
      {/each}
    </MenuItems>
  </Menu>
</div>
