<script>
  import { INTERVALS, TIMESTAMPS } from "$lib/duckdb-data-types";
  import {
    formatDataType,
    standardTimestampFormat,
  } from "$lib/util/formatters";
  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import { createEventDispatcher } from "svelte";
  import { FormattedDataType } from "../data-types";
  import notificationStore from "../notifications";
  import Shortcut from "../tooltip/Shortcut.svelte";
  import StackingWord from "../tooltip/StackingWord.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";

  export let row;
  export let column;
  export let value;
  export let type;
  export let rowActive = false;
  export let suppressTooltip = false;

  let cellActive = false;

  const dispatch = createEventDispatcher();

  const { shiftClickAction } = createShiftClickAction();

  let formattedValue;
  $: {
    if (TIMESTAMPS.has(type)) {
      formattedValue = standardTimestampFormat(value, type);
    } else if (value === null) {
      formattedValue = `∅ null`;
    } else {
      if (typeof value === "string" && !value.length) {
        // replace with a whitespace chracter to preserve the cell height when we have an empty string
        formattedValue = "&nbsp;";
      } else {
        formattedValue = value;
      }
    }
  }

  function onFocus() {
    dispatch("inspect", row.index);
    cellActive = true;
  }

  function onBlur() {
    cellActive = false;
  }

  /** Because this table is virtualized,
   * it's a bit harder to get the proper
   * row-based hover highlighting. So let's
   * use javascript to solve this issue.
   */
  let activityStatus;
  $: {
    if (cellActive) {
      activityStatus = "bg-gray-200";
    } else if (rowActive && !cellActive) {
      activityStatus = "bg-gray-100";
    } else {
      activityStatus = "bg-white";
    }
  }
</script>

<Tooltip location="top" distance={16} suppress={suppressTooltip}>
  <div
    on:mouseover={onFocus}
    on:mouseout={onBlur}
    on:focus={onFocus}
    on:blur={onBlur}
    class="
      absolute 
      z-9 
      text-ellipsis 
      whitespace-nowrap 
      border-r border-b 
      {activityStatus}
      "
    style:left="{column.start}px"
    style:top="{row.start}px"
    style:width="{column.size}px"
    style:height="{row.size}px"
  >
    <button
      class="
      py-2 px-4 
      text-left w-full text-ellipsis overflow-x-hidden whitespace-nowrap"
      use:shiftClickAction
      on:shift-click={async () => {
        let exportedValue = value;
        if (INTERVALS.has(type)) {
          exportedValue = formatDataType(value, type);
        } else if (TIMESTAMPS.has(type)) {
          exportedValue = `TIMESTAMP '${value}'`;
        }
        await navigator.clipboard.writeText(exportedValue);
        notificationStore.send({ message: `copied value to clipboard` });
        // update this to set the active animation in the tooltip text
      }}
    >
      <FormattedDataType
        value={formattedValue}
        {type}
        isNull={value === null}
        inTable
      />
    </button>
  </div>
  <TooltipContent slot="tooltip-content">
    <TooltipTitle>
      <svelte:fragment slot="name">
        <FormattedDataType {value} {type} dark />
      </svelte:fragment>
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">copy</StackingWord> this value to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
