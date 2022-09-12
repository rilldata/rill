<script lang="ts">
  import { createEventDispatcher, getContext } from "svelte";
  import { fly } from "svelte/transition";
  import { DataTypeIcon } from "$lib/components/data-types";
  import ArrowDown from "$lib/components/icons/ArrowDown.svelte";
  import Pin from "$lib/components/icons/Pin.svelte";
  import notificationStore from "$lib/components/notifications";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import type { HeaderPosition, VirtualizedTableConfig } from "../types";
  import StickyHeader from "./StickyHeader.svelte";

  export let pinned = false;
  export let noPin = false;
  export let showDataIcon = false;
  export let name: string;
  export let type: string;
  export let header;
  export let position: HeaderPosition = "top";
  export let enableResize = true;
  export let isSelected = false;

  const config: VirtualizedTableConfig = getContext("config");
  const dispatch = createEventDispatcher();

  const { shiftClickAction } = createShiftClickAction();

  let showMore = false;
  $: isSortingDesc = true;

  $: columnFontWeight = isSelected
    ? "font-bold"
    : config.columnHeaderFontWeightClass;
</script>

<StickyHeader
  {enableResize}
  on:reset-column-width={() => {
    dispatch("reset-column-size", { name });
  }}
  on:resize={(event) => {
    dispatch("resize-column", {
      size: event.detail.size,
      name,
    });
  }}
  {position}
  {header}
  on:focus={() => {
    showMore = true;
  }}
  on:blur={() => {
    showMore = false;
  }}
  on:click={() => {
    if (isSelected) isSortingDesc = !isSortingDesc;
    else isSortingDesc = true;
    dispatch("click-column");
  }}
>
  <div
    use:shiftClickAction
    on:shift-click={async () => {
      await navigator.clipboard.writeText(name);
      notificationStore.send({
        message: `copied column name "${name}" to clipboard`,
      });
    }}
    class="
           flex
           items-center
           justify-stretch
           select-none
           gap-x-2
           "
  >
    <Tooltip location="top" alignment="middle" distance={16}>
      <div
        class="
        grid
        items-center cursor-pointer
        {isSelected ? '' : 'w-full gap-x-2'}
        "
        style:grid-template-columns="max-content auto {!noPin && showMore
          ? "max-content"
          : ""}"
      >
        {#if showDataIcon}
          <DataTypeIcon suppressTooltip color={"text-gray-500"} {type} />
        {/if}
        <span
          class="text-ellipsis overflow-hidden whitespace-nowrap {columnFontWeight}"
        >
          {name}
        </span>
      </div>
      <TooltipContent slot="tooltip-content">
        <TooltipTitle>
          <svelte:fragment slot="name">
            {name}
          </svelte:fragment>
          <svelte:fragment slot="description">
            {type}
          </svelte:fragment>
        </TooltipTitle>
        <TooltipShortcutContainer>
          <div>
            <StackingWord key="shift">copy</StackingWord>
            column name to clipboard
          </div>
          <Shortcut>
            <span style="font-family: var(--system);">⇧</span> + Click
          </Shortcut>
        </TooltipShortcutContainer>
      </TooltipContent>
    </Tooltip>
    {#if isSelected}
      {#if isSortingDesc}
        <div in:fly|local={{ duration: 200, y: -8 }}>
          <ArrowDown size="16px" />
        </div>
      {:else}
        <div in:fly|local={{ duration: 200, y: 8 }}>
          <ArrowDown transform="scale(1 -1)" size="16px" />
        </div>
      {/if}
    {/if}
    {#if !noPin && showMore}
      <Tooltip location="top" alignment="middle" distance={16}>
        <button
          transition:fly|local={{ duration: 200, y: 4 }}
          class:text-gray-900={pinned}
          class:text-gray-400={!pinned}
          class="transition-colors duration-100 justify-self-end"
          on:click={() => {
            dispatch("pin");
          }}
        >
          <Pin size="16px" />
        </button>
        <TooltipContent slot="tooltip-content">
          {pinned
            ? "unpin this column from the right side of the table"
            : "pin this column to the right side of the table"}
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</StickyHeader>
