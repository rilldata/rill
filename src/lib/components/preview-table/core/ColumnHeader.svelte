<script lang="ts">
  import { DataTypeIcon } from "$lib/components/data-types";
  import Pin from "$lib/components/icons/Pin.svelte";
  import notificationStore from "$lib/components/notifications";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";
  import type { HeaderPosition } from "../types";
  import StickyHeader from "./StickyHeader.svelte";

  export let pinned = false;
  export let name: string;
  export let type: string;
  export let header;
  export let position: HeaderPosition = "top";

  const dispatch = createEventDispatcher();

  const { shiftClickAction } = createShiftClickAction();

  let showMore = false;
</script>

<StickyHeader
  {position}
  {header}
  on:focus={() => {
    showMore = true;
  }}
  on:blur={() => {
    showMore = false;
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
        w-full gap-x-2 items-center cursor-pointer"
        style:grid-template-columns="max-content auto max-content"
      >
        <DataTypeIcon suppressTooltip color={"text-gray-500"} {type} />
        <span class="text-ellipsis overflow-hidden whitespace-nowrap font-bold">
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
    {#if showMore}
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
