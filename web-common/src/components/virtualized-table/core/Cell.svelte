<script lang="ts">
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import {
    copyToClipboard,
    isClipboardApiSupported,
  } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { STRING_LIKES } from "@rilldata/web-common/lib/duckdb-data-types";
  import { formatDataTypeAsDuckDbQueryString } from "@rilldata/web-common/lib/formatters";
  import { createEventDispatcher, getContext } from "svelte";
  import { cellInspectorStore } from "@rilldata/web-common/features/dashboards/stores/cell-inspector-store";
  import BarAndLabel from "../../BarAndLabel.svelte";
  import type { VirtualizedTableConfig } from "../types";

  export let row;
  export let column;
  export let value;
  export let formattedValue: string | null = null;
  export let type;
  export let barValue = 0;
  export let rowActive = false;
  export let suppressTooltip = false;
  export let rowSelected = false;
  export let colSelected = false;
  export let atLeastOneSelected = false;
  export let excludeMode = false;
  export let positionStatic = false;
  export let label: string | undefined = undefined;

  const config: VirtualizedTableConfig = getContext("config");
  const isDimensionTable = config.table === "DimensionTable";

  let cellActive = false;
  $: isTextColumn = type === "VARCHAR" || type === "CODE_STRING";

  const dispatch = createEventDispatcher();

  function onFocus() {
    dispatch("inspect", row.index);
    cellActive = true;
    // Update the cell inspector store with the cell value
    if (value !== undefined && value !== null) {
      cellInspectorStore.updateValue(value.toString());
    }
  }

  function onSelectItem(e: MouseEvent) {
    if (e.shiftKey) return;

    // Check if user has selected text
    const selection = window.getSelection();
    if (selection && selection.toString().length > 0) {
      // User has selected text, don't trigger row selection
      return;
    }

    dispatch("select-item", { index: row.index, meta: e.ctrlKey || e.metaKey });
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
      // Specific cell active color, used to be bg-gray-200
      // bg-gray-100 to match the hover color, and not too hard on the eyes
      activityStatus = "bg-gray-100 ";
    } else if (rowActive && !cellActive) {
      activityStatus = "bg-gray-100 ";
    } else if (colSelected) {
      activityStatus = "bg-surface";
    } else {
      activityStatus = "bg-surface";
    }
  }

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneSelected
    ? excludeMode
      ? rowSelected
      : !rowSelected
    : false;

  $: barColor = excluded
    ? "ui-measure-bar-excluded"
    : rowSelected
      ? "ui-measure-bar-included-selected"
      : "ui-measure-bar-included";

  $: tooltipValue =
    value && STRING_LIKES.has(type) && value.length >= TOOLTIP_STRING_LIMIT
      ? value?.slice(0, TOOLTIP_STRING_LIMIT) + "..."
      : value;

  $: formattedDataTypeStyle = excluded
    ? "font-normal ui-copy-disabled-faint"
    : rowSelected
      ? "font-normal ui-copy-strong"
      : "font-normal ui-copy";

  const shiftClick = async () => {
    let exportedValue = formatDataTypeAsDuckDbQueryString(value, type);
    copyToClipboard(exportedValue);
    // update this to set the active animation in the tooltip text
  };
</script>

<Tooltip
  distance={16}
  location="top"
  suppress={suppressTooltip || !isClipboardApiSupported()}
>
  <div
    class="
      {positionStatic ? 'static' : 'absolute'}
      z-9
      text-ellipsis
      whitespace-nowrap
      {isDimensionTable ? '' : 'border-r border-b'}
      {activityStatus}
      "
    on:blur={onBlur}
    on:click={onSelectItem}
    on:focus={onFocus}
    on:keydown
    on:mouseout={onBlur}
    on:mouseover={onFocus}
    role="gridcell"
    style:height="{row.size}px"
    style:left="{column.start}px"
    style:top="{row.start}px"
    style:width="{column.size}px"
    style:padding-right="10px"
    tabindex="0"
  >
    <BarAndLabel
      color={barColor}
      customBackgroundColor="rgba(0,0,0,0)"
      justify="left"
      showBackground={false}
      value={barValue}
      compact
    >
      <button
        aria-label={label}
        class="{isTextColumn ? 'text-left' : 'text-right'} w-full truncate"
        class:px-4={!isDimensionTable}
        on:click={modified({
          shift: shiftClick,
        })}
        style:height="{row.size}px"
      >
        <FormattedDataType
          customStyle={formattedDataTypeStyle}
          inTable
          isNull={value === null || value === undefined}
          {type}
          value={formattedValue || value}
          color="text-gray-500"
        />
      </button>
    </BarAndLabel>
  </div>
  <TooltipContent maxWidth="360px" slot="tooltip-content">
    <TooltipTitle>
      <FormattedDataType slot="name" value={tooltipValue} />
    </TooltipTitle>
    <TooltipShortcutContainer>
      <div>
        <StackingWord key="shift">Copy</StackingWord> this value to clipboard
      </div>
      <Shortcut>
        <span style="font-family: var(--system);">⇧</span> + Click
      </Shortcut>
    </TooltipShortcutContainer>
  </TooltipContent>
</Tooltip>
