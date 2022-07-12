<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { TimeRange, timeRanges } from "$lib/util/time-ranges";
  import { tick } from "svelte";
  import { defaultTimeRange, prettyFormatTimeRange } from "./utils";

  let selectedTimeRange: TimeRange = defaultTimeRange;

  let timeSelectorMenu;
  let timeSelectorMenuOpen = false;
  let clickOutsideListener;
  $: if (!timeSelectorMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
  }

  const buttonClickHandler = async () => {
    timeSelectorMenuOpen = !timeSelectorMenuOpen;
    if (!clickOutsideListener) {
      await tick();
      clickOutsideListener = onClickOutside(() => {
        timeSelectorMenuOpen = false;
      }, timeSelectorMenu);
    }
  };

  let target: HTMLElement;
</script>

<Button
  bind:element={target}
  override="border-none gap-x-4"
  on:click={buttonClickHandler}
>
  <span class="font-bold">
    {selectedTimeRange.name}
  </span>
  <span>
    {prettyFormatTimeRange(selectedTimeRange)}
  </span>
  <CaretDownIcon size="16px" />
</Button>

{#if timeSelectorMenuOpen}
  <div bind:this={timeSelectorMenu}>
    <FloatingElement
      relationship="direct"
      location="bottom"
      alignment="start"
      {target}
      distance={8}
    >
      <Menu on:escape={() => (timeSelectorMenuOpen = false)}>
        {#each timeRanges as timeRange}
          <MenuItem on:select={() => (selectedTimeRange = timeRange)}>
            <div class="font-bold">
              {timeRange.name}
            </div>
            <div slot="right" let:hovered class:opacity-0={!hovered}>
              {prettyFormatTimeRange(timeRange)}
            </div>
          </MenuItem>
        {/each}
      </Menu>
    </FloatingElement>
  </div>
{/if}
