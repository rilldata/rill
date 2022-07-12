<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import Menu from "$lib/components/menu/Menu.svelte";
  import MenuItem from "$lib/components/menu/MenuItem.svelte";
  import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
  import { onClickOutside } from "$lib/util/on-click-outside";
  import { TimeRange, timeRanges } from "$lib/util/time-ranges";
  import { tick } from "svelte";

  const defaultTimeRange = timeRanges[0];
  let selectedTimeRange: TimeRange = defaultTimeRange;

  const prettyFormatTimeRange = (timeRange: TimeRange): string => {
    // day is the same
    if (
      timeRange.start.getDate() === timeRange.end.getDate() &&
      timeRange.start.getMonth() === timeRange.end.getMonth() &&
      timeRange.start.getFullYear() === timeRange.end.getFullYear()
    ) {
      return `${timeRange.start.toLocaleDateString(undefined, {
        month: "long",
      })} ${timeRange.start.getDate()}, ${timeRange.start.getFullYear()} (${timeRange.start.toLocaleString(
        undefined,
        { hour12: true, hour: "numeric", minute: "numeric" }
      )} - ${timeRange.end.toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
      })})`;
    }
    // month is the same
    if (
      timeRange.start.getMonth() === timeRange.end.getMonth() &&
      timeRange.start.getFullYear() === timeRange.end.getFullYear()
    ) {
      return `${timeRange.start.toLocaleDateString(undefined, {
        month: "long",
      })} ${timeRange.start.getDate()}-${timeRange.end.getDate()}, ${timeRange.start.getFullYear()}`;
    }
    // year is the same
    if (timeRange.start.getFullYear() === timeRange.end.getFullYear()) {
      return `${timeRange.start.toLocaleDateString(undefined, {
        month: "long",
        day: "numeric",
      })} - ${timeRange.end.toLocaleDateString(undefined, {
        month: "long",
        day: "numeric",
      })}, ${timeRange.start.getFullYear()}`;
    }
    // year is different
    const dateFormatOptions: Intl.DateTimeFormatOptions = {
      year: "numeric",
      month: "long",
      day: "numeric",
    };
    return `${timeRange.start.toLocaleDateString(
      undefined,
      dateFormatOptions
    )} - ${timeRange.end.toLocaleDateString(undefined, dateFormatOptions)}`;
  };

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
