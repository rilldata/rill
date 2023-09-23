<script lang="ts">
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { createEventDispatcher } from "svelte";
  import { fly, slide } from "svelte/transition";
  import BarAndLabel from "../../../components/BarAndLabel.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";

  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";

  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";

  import { notifications } from "@rilldata/web-common/components/notifications";

  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";

  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";

  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import { humanizeDataType } from "../humanize-numbers";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    LeaderboardItemData,
    formatContextColumnValue,
  } from "./leaderboard-utils";
  import ContextColumnValue from "./ContextColumnValue.svelte";

  export let itemData: LeaderboardItemData;
  $: label = itemData.dimensionValue;
  $: measureValue = itemData.value;
  $: selected = itemData.selectedIndex >= 0;
  $: comparisonValue = itemData.prevValue;

  export let contextColumn: LeaderboardContextColumn;

  export let atLeastOneActive = false;
  export let isBeingCompared = false;
  export let formattedValue: string;
  export let filterExcludeMode;

  export let formatPreset;

  /** if this value is a summable measure, we'll show the bar. Otherwise, don't. */
  export let isSummableMeasure;
  /** for summable measures, this is the value we use to calculate the bar % to fill */
  export let referenceValue;

  $: formattedValue = humanizeDataType(measureValue, formatPreset);

  $: contextColumnFormattedValue = formatContextColumnValue(
    itemData,
    contextColumn,
    formatPreset
  );

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? humanizeDataType(comparisonValue, formatPreset)
      : undefined;
  $: showPreviousTimeValue = hovered && previousValueString !== undefined;
  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  let renderedBarValue = 0; // should be between 0 and 1.
  $: {
    renderedBarValue = isSummableMeasure
      ? referenceValue
        ? measureValue / referenceValue
        : 0
      : 0;
    // if this somehow creates an NaN, let's set it to 0.
    renderedBarValue = !isNaN(renderedBarValue) ? renderedBarValue : 0;
  }
  $: color = excluded
    ? "ui-measure-bar-excluded"
    : selected
    ? "ui-measure-bar-included-selected"
    : "ui-measure-bar-included";

  const { shiftClickAction } = createShiftClickAction();
  async function shiftClickHandler(label) {
    await navigator.clipboard.writeText(label);
    let truncatedLabel = label?.toString();
    if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
      truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
    }
    notifications.send({
      message: `copied dimension value "${truncatedLabel}" to clipboard`,
    });
  }
  const dispatch = createEventDispatcher();

  let hovered = false;
  const onHover = () => {
    hovered = true;
    dispatch("focus");
  };
  const onLeave = () => {
    hovered = false;
    dispatch("blur");
  };
</script>

<Tooltip location="right">
  <button
    class="flex flex-row items-center w-full text-left transition-color"
    on:blur={onLeave}
    on:click={(e) => {
      if (e.shiftKey) return;
      dispatch("select-item", {
        label,
      });
    }}
    on:focus={onHover}
    on:keydown
    on:mouseleave={onLeave}
    on:mouseover={onHover}
    on:shift-click={() => shiftClickHandler(label)}
    transition:slide|local={{ duration: 200 }}
    use:shiftClickAction
  >
    <LeaderboardItemFilterIcon
      {isBeingCompared}
      {excluded}
      selectionIndex={itemData?.selectedIndex}
      defaultComparedIndex={itemData?.defaultComparedIndex}
    />
    <BarAndLabel
      {color}
      justify={false}
      showBackground={false}
      showHover
      tweenParameters={{ duration: 200 }}
      value={renderedBarValue}
    >
      <div
        class="grid leaderboard-entry items-center gap-x-3"
        style:height="22px"
      >
        <!-- NOTE: empty class leaderboard-label is used to locate this elt in e2e tests -->
        <div
          class="leaderboard-label justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
          class:ui-copy={!atLeastOneActive}
          class:ui-copy-disabled={excluded}
          class:ui-copy-strong={!excluded && selected}
        >
          <FormattedDataType value={label} />
        </div>

        <div
          class="justify-self-end overflow-hidden ui-copy-number flex gap-x-4 items-baseline"
        >
          <div
            class="flex items-baseline gap-x-1"
            in:fly={{ duration: 200, y: 4 }}
          >
            {#if showPreviousTimeValue}
              <span
                class="inline-block opacity-50"
                transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
              >
                {previousValueString}
                →
              </span>
            {/if}
            <FormattedDataType
              type="INTEGER"
              value={formattedValue || measureValue}
            />
          </div>
          <ContextColumnValue
            formattedValue={contextColumnFormattedValue}
            {contextColumn}
          />
        </div>
      </div>
    </BarAndLabel>
  </button>
  <!-- if the value is greater than 100%, we should add this little serration -->
  {#if renderedBarValue > 1.001}
    <LongBarZigZag />
  {/if}

  <LeaderboardTooltipContent
    {atLeastOneActive}
    {excluded}
    {filterExcludeMode}
    {label}
    slot="tooltip-content"
  />
</Tooltip>

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
