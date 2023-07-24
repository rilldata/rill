<!-- @component
Creates a set of DimensionLeaderboardEntry components. This component makes it easy
to stitch together  chunks of a list. For instance, we can have:
leaderboard values above the fold
divider
leaderboard values not visible but selected
divider
see more button
-->
<script lang="ts">
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import {
    LIST_SLIDE_DURATION,
    TOOLTIP_STRING_LIMIT,
  } from "@rilldata/web-common/layout/config";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { createEventDispatcher } from "svelte";
  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import { PERC_DIFF } from "../../../components/data-types/type-utils";
  import {
    formatMeasurePercentageDifference,
    humanizeDataType,
  } from "../humanize-numbers";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";
  import { FormattedDataType } from "../../../components/data-types";

  export let values;
  export let comparisonValues;
  export let showTimeComparison = false;
  export let showPercentOfTotal = false;

  export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let unfilteredTotal: number;

  export let atLeastOneActive;
  export let loading = false;
  export let formatPreset;

  const { shiftClickAction } = createShiftClickAction();

  const dispatch = createEventDispatcher();
  let renderValues = [];

  $: showComparison = showTimeComparison || showPercentOfTotal;

  $: comparisonMap = new Map(comparisonValues?.map((v) => [v.label, v.value]));

  // FIXME: in no world should it be the responsibility of this component to
  // merge `values` and `comparisonValues` and `activeValues`. This should be
  // done somewhere upstream -- ideally, not in a component at all, but given
  // the current architecture, it should at least happen in the parent component.
  $: renderValues = values.map((v) => {
    const active = activeValues.findIndex((value) => value === v.label) >= 0;
    const comparisonValue = comparisonMap.get(v.label);

    // Super important special case: if there is not at least one "active" (selected) value,
    // we need to set *all* items to be included, because by default if a user has not
    // selected any values, we assume they want all values included in all calculations.
    const excluded = atLeastOneActive
      ? (filterExcludeMode && active) || (!filterExcludeMode && !active)
      : false;

    // FIXME: `showComparisonForThisValue` should not be the responsibility
    // of this component; the handling of the on:mouseenter/on:mouseleave
    // events should be done in each individual DimensionLeaderboardEntry
    const showComparisonForThisValue = comparisonLabelToReveal === v.label;

    const previousValueString: string | undefined =
      showComparisonForThisValue &&
      comparisonValue !== undefined &&
      comparisonValue !== null
        ? humanizeDataType(comparisonValue, formatPreset)
        : undefined;

    const percentChangeFormatted = showTimeComparison
      ? getFormatterValueForPercDiff(
          v.value && comparisonValue ? v.value - comparisonValue : null,
          comparisonValue
        )
      : showPercentOfTotal
      ? getFormatterValueForPercDiff(v.value, unfilteredTotal)
      : undefined;

    return {
      ...v,
      active,
      excluded,
      comparisonValue,
      formattedValue: humanizeDataType(v.value, formatPreset),
      previousValueString,
      percentChangeFormatted,
    };
  });

  let comparisonLabelToReveal = undefined;
  function revealComparisonNumber(value) {
    return () => {
      if (showTimeComparison) comparisonLabelToReveal = value;
    };
  }

  function getFormatterValueForPercDiff(numerator, denominator) {
    if (denominator === 0) return PERC_DIFF.PREV_VALUE_ZERO;
    if (!denominator) return PERC_DIFF.PREV_VALUE_NO_DATA;
    if (numerator === null || numerator === undefined)
      return PERC_DIFF.CURRENT_VALUE_NO_DATA;

    const percDiff = numerator / denominator;
    return formatMeasurePercentageDifference(percDiff);
  }

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
</script>

{#each renderValues as { label, value, active, excluded, percentChangeFormatted, formattedValue, previousValueString } (label)}
  <!-- FIXME: this wrapper div is almost certainly not required. All of this functionality should be able to be handled in DimensionLeaderboardEntry -->
  <div
    use:shiftClickAction
    on:click={() => {
      dispatch("select-item", {
        label,
      });
    }}
    on:mouseenter={revealComparisonNumber(label)}
    on:mouseleave={revealComparisonNumber(undefined)}
    on:keydown
    on:shift-click={() => shiftClickHandler(label)}
  >
    <DimensionLeaderboardEntry
      measureValue={value}
      showContext={showComparison}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
      {excluded}
    >
      <svelte:fragment slot="label">
        <FormattedDataType value={label} />
      </svelte:fragment>
      <div slot="right" class="flex items-baseline gap-x-1">
        {#if previousValueString}
          <span
            class="inline-block opacity-50"
            transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
          >
            {previousValueString}
            →
          </span>
        {/if}
        <FormattedDataType type="INTEGER" value={formattedValue || value} />
      </div>
      <span slot="context">
        {#if showTimeComparison || showPercentOfTotal}
          <PercentageChange value={percentChangeFormatted} />
        {/if}
      </span>
      <svelte:fragment slot="tooltip">
        <TooltipTitle>
          <svelte:fragment slot="name">
            {label}
          </svelte:fragment>
        </TooltipTitle>

        <TooltipShortcutContainer>
          {#if atLeastOneActive}
            <div>
              {excluded ? "Include" : "Exclude"}
              this dimension value
            </div>
          {:else}
            <div class="text-ellipsis overflow-hidden whitespace-nowrap">
              Filter {filterExcludeMode ? "out" : "on"}
              this dimension value
            </div>
          {/if}
          <Shortcut>Click</Shortcut>
        </TooltipShortcutContainer>
        <TooltipShortcutContainer>
          <div>
            <StackingWord key="shift">Copy</StackingWord>
            this dimension value to clipboard
          </div>
          <Shortcut>
            <span style="font-family: var(--system);">⇧</span> + Click
          </Shortcut>
        </TooltipShortcutContainer>
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
