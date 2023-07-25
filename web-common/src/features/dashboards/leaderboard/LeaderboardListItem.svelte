<script lang="ts">
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

  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import { humanizeDataType } from "../humanize-numbers";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    LeaderboardItemData,
    getFormatterValueForPercDiff,
  } from "./leaderboard-utils";

  export let itemData: LeaderboardItemData;
  $: label = itemData.label;
  $: measureValue = itemData.value;
  $: selected = itemData.selected;
  $: comparisonValue = itemData.comparisonValue;

  export let showContext: "time" | "percent" | false = false;

  export let atLeastOneActive = false;

  export let formattedValue: string;
  export let filterExcludeMode;

  export let formatPreset;

  /** if this value is a summable measure, we'll show the bar. Otherwise, don't. */
  export let isSummableMeasure;
  /** for summable measures, this is the value we use to calculate the bar % to fill */
  export let referenceValue;

  /** the unfilteredTotal is used to compute the "percent of total" context value*/
  export let unfilteredTotal: number;

  $: formattedValue = humanizeDataType(measureValue, formatPreset);

  $: percentChangeFormatted =
    showContext === "time"
      ? getFormatterValueForPercDiff(
          measureValue && comparisonValue
            ? measureValue - comparisonValue
            : null,
          comparisonValue
        )
      : showContext === "percent"
      ? getFormatterValueForPercDiff(measureValue, unfilteredTotal)
      : undefined;

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
    class="flex flex-row w-full text-left transition-color"
    on:blur={onLeave}
    on:focus={onHover}
    on:mouseleave={onLeave}
    on:mouseover={onHover}
    transition:slide|local={{ duration: 200 }}
    use:shiftClickAction
    on:shift-click={() => shiftClickHandler(label)}
    on:click={() => {
      dispatch("select-item", {
        label,
      });
    }}
    on:keydown
  >
    <LeaderboardItemFilterIcon {selected} {excluded} />
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
          class:ui-copy={!atLeastOneActive}
          class:ui-copy-strong={!excluded && selected}
          class:ui-copy-disabled={excluded}
          class="leaderboard-label justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
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
          {#if showContext}
            <div
              class="text-xs text-gray-500 dark:text-gray-400"
              style:width="44px"
            >
              <PercentageChange value={percentChangeFormatted} />
            </div>
          {/if}
        </div>
      </div>
    </BarAndLabel>
  </button>
  <!-- if the value is greater than 100%, we should add this little serration -->
  {#if renderedBarValue > 1.001}
    <LongBarZigZag />
  {/if}

  <LeaderboardTooltipContent
    slot="tooltip-content"
    {label}
    {atLeastOneActive}
    {excluded}
    {filterExcludeMode}
  />
</Tooltip>

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
