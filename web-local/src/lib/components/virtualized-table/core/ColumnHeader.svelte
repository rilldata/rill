<script lang="ts">
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import Pin from "@rilldata/web-common/components/icons/Pin.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { createEventDispatcher, getContext } from "svelte";
  import { fly } from "svelte/transition";
  import { createShiftClickAction } from "../../../util/shift-click-action";
  import { DataTypeIcon } from "../../data-types";
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

  $: isDimensionTable = config.table === "DimensionTable";
  $: isDimensionColumn = isDimensionTable && type === "VARCHAR";

  $: textAlignment = isDimensionColumn ? "text-left pl-1" : "text-right pr-1";

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
      notifications.send({
        message: `copied column name "${name}" to clipboard`,
      });
    }}
    class="
           flex
           justify-stretch
           select-none
           over
           {isDimensionTable ? '' : 'items-center gap-x-2'}
           "
  >
    <Tooltip location="top" alignment="middle" distance={16}>
      <div
        class="
        grid
        items-center cursor-pointer w-full
        {isSelected ? '' : 'gap-x-2'}
        "
        style:grid-template-columns={isDimensionTable
          ? ""
          : `max-content auto ${!noPin && showMore ? "max-content" : ""}`}
      >
        {#if showDataIcon}
          <DataTypeIcon suppressTooltip color={"text-gray-500"} {type} />
        {/if}
        <span
          class="text-ellipsis
          {columnFontWeight}
          {isDimensionTable
            ? `${textAlignment} break-words line-clamp-2`
            : 'overflow-hidden whitespace-nowrap'}
          "
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
            {isDimensionTable || showDataIcon ? "" : type}
          </svelte:fragment>
        </TooltipTitle>
        <TooltipShortcutContainer>
          <div>
            <StackingWord key="shift">Copy</StackingWord>
            column name to clipboard
          </div>
          <Shortcut>
            <span style="font-family: var(--system);">⇧</span> + Click
          </Shortcut>
        </TooltipShortcutContainer>
      </TooltipContent>
    </Tooltip>

    {#if isDimensionTable}
      <div class="mt-0.5 ui-copy-icon">
        {#if isSortingDesc}
          <div
            in:fly={{ duration: 200, y: -8 }}
            style:opacity={isSelected ? 1 : 0}
          >
            <ArrowDown size="16px" />
          </div>
        {:else}
          <div
            in:fly={{ duration: 200, y: 8 }}
            style:opacity={isSelected ? 1 : 0}
          >
            <ArrowDown transform="scale(1 -1)" size="16px" />
          </div>
        {/if}
      </div>
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
