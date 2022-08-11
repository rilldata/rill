<script lang="ts">
  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import { createEventDispatcher } from "svelte";
  import { DataTypeIcon } from "../data-types";
  import Pin from "../icons/Pin.svelte";
  import notificationStore from "../notifications";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";
  import StickyHeader from "./StickyHeader.svelte";

  export let pinned = false;
  export let name: string;
  export let type: string;
  export let header;
  export let position = "top";

  const dispatch = createEventDispatcher();

  const { shiftClickAction } = createShiftClickAction();
</script>

<StickyHeader {position} {header}>
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
    <Tooltip location="top" alignment="middle" distance={16}>
      <button
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
  </div>
</StickyHeader>
