<script>
  import { FormattedDataType } from "$lib/components/data-types";
  import notificationStore from "$lib/components/notifications";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import { INTERVALS, STRING_LIKES, TIMESTAMPS } from "$lib/duckdb-data-types";
  import { formatDataType } from "$lib/util/formatters";
  import { createShiftClickAction } from "$lib/util/shift-click-action";
  import { createEventDispatcher } from "svelte";

  export let row;
  export let column;
  export let value;
  export let type;
  export let rowActive = false;
  export let suppressTooltip = false;

  let cellActive = false;

  const dispatch = createEventDispatcher();

  const { shiftClickAction } = createShiftClickAction();

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

  let TOOLTIP_STRING_LIMIT = 200;
  $: tooltipValue =
    value && STRING_LIKES.has(type) && value.length >= TOOLTIP_STRING_LIMIT
      ? value?.slice(0, TOOLTIP_STRING_LIMIT) + "..."
      : value;
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
      <FormattedDataType {value} {type} inTable />
    </button>
  </div>
  <TooltipContent slot="tooltip-content" maxWidth="360px">
    <TooltipTitle>
      <svelte:fragment slot="name">
        <FormattedDataType value={tooltipValue} {type} dark />
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
