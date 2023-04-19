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
  import { FormattedDataType, Number } from "../../../components/data-types";

  export let values;
  export let comparisonValues;
  export let showComparison = false;

  export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;
  export let formatPreset;

  const { shiftClickAction } = createShiftClickAction();

  const dispatch = createEventDispatcher();
  let renderValues = [];

  $: comparsionMap = new Map(comparisonValues?.map((v) => [v.label, v.value]));
  $: renderValues = values.map((v) => {
    const active = activeValues.findIndex((value) => value === v.label) >= 0;
    const comparisonValue = comparsionMap.get(v.label);

    // Super important special case: if there is not at least one "active" (selected) value,
    // we need to set *all* items to be included, because by default if a user has not
    // selected any values, we assume they want all values included in all calculations.
    const excluded = atLeastOneActive
      ? (filterExcludeMode && active) || (!filterExcludeMode && !active)
      : false;

    return { ...v, active, excluded, comparisonValue };
  });

  let comparisonLabelToReveal = undefined;
  function revealComparisonNumber(value) {
    return () => {
      if (showComparison) comparisonLabelToReveal = value;
    };
  }

  function getFormatterValueForPercDiff(comparisonValue, value) {
    if (comparisonValue === 0) return PERC_DIFF.PREV_VALUE_ZERO;
    if (!comparisonValue) return PERC_DIFF.PREV_VALUE_NO_DATA;
    if (value === null || value === undefined)
      return PERC_DIFF.CURRENT_VALUE_NO_DATA;

    const percDiff = (value - comparisonValue) / comparisonValue;
    return formatMeasurePercentageDifference(percDiff);
  }
</script>

{#each renderValues as { label, value, active, excluded, comparisonValue } (label)}
  {@const formattedValue = humanizeDataType(value, formatPreset)}
  {@const showComparisonForThisValue = comparisonLabelToReveal === label}

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
    on:shift-click={async () => {
      await navigator.clipboard.writeText(label);
      let truncatedLabel = label?.toString();
      if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
        truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
      }
      notifications.send({
        message: `copied dimension value "${truncatedLabel}" to clipboard`,
      });
    }}
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
        <FormattedDataType isNull={label === null} value={label} />
      </svelte:fragment>
      <div slot="right" class="flex items-baseline gap-x-1">
        {#if showComparisonForThisValue && comparisonValue !== undefined}
          <span
            class="inline-block opacity-50"
            transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
          >
            {humanizeDataType(comparisonValue, formatPreset)}
            →
          </span>
        {/if}
        <Number
          type="INTEGER"
          isNull={!(formattedValue || value)}
          value={formattedValue || value}
        />
      </div>
      <span slot="context">
        <PercentageChange
          value={getFormatterValueForPercDiff(comparisonValue, value)}
        />
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
