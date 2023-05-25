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
  // import LeaderboardItemTooltip from "./LeaderboardItemTooltip.svelte";
  import {
    // LIST_SLIDE_DURATION,
    TOOLTIP_STRING_LIMIT,
  } from "@rilldata/web-common/layout/config";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  // import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { createEventDispatcher } from "svelte";
  // import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  // import { PERC_DIFF } from "../../../components/data-types/type-utils";
  // import {
  //   // formatMeasurePercentageDifference,
  //   // humanizeDataType,
  // } from "../humanize-numbers";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";
  import {
    LeaderboardRenderValue,
    valuesToRenderValues,
  } from "./leaderboard-render-values";
  // import { FormattedDataType } from "../../../components/data-types";
  // import LeaderboardEntryRightValue from "./LeaderboardEntryRightValue.svelte";

  export let values;
  export let comparisonValues;
  export let showComparison = false;

  export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let totalFilteredRowCount: number;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;
  export let formatPreset;

  const { shiftClickAction } = createShiftClickAction();

  const dispatch = createEventDispatcher();
  let renderValues: LeaderboardRenderValue[] = [];

  $: comparisonMap = new Map(comparisonValues?.map((v) => [v.label, v.value]));

  $: renderValues = valuesToRenderValues(
    values,
    activeValues,
    comparisonMap,
    comparisonLabelToReveal,
    filterExcludeMode,
    atLeastOneActive,
    formatPreset
  );
  // values.map((v) => {
  //   const active = activeValues.findIndex((value) => value === v.label) >= 0;
  //   const comparisonValue = comparsionMap.get(v.label);

  //   // Super important special case: if there is not at least one "active" (selected) value,
  //   // we need to set *all* items to be included, because by default if a user has not
  //   // selected any values, we assume they want all values included in all calculations.
  //   const excluded = atLeastOneActive
  //     ? (filterExcludeMode && active) || (!filterExcludeMode && !active)
  //     : false;

  //   return {
  //     ...v,
  //     active,
  //     excluded,
  //     comparisonValue,
  //     formattedValue: humanizeDataType(v.value, formatPreset),
  //     showComparisonForThisValue: comparisonLabelToReveal === v.label,
  //   };
  // });
  // $: console.log("renderValues", renderValues);

  let comparisonLabelToReveal = undefined;
  function revealComparisonNumber(value) {
    return () => {
      if (showComparison) comparisonLabelToReveal = value;
    };
  }
</script>

<!-- {#each renderValues as { label, value, rowCount, active, excluded, comparisonValue, formattedValue, showComparisonForThisValue } (label)} -->
{#each renderValues as renderValue (renderValue.label)}
  {@const {
    label,
    value,
    // rowCount,
    active,
    excluded,
    comparisonValue,
    formattedValue,
    showComparisonForThisValue,
  } = renderValue}
  <!-- {@const showComparisonForThisValue = comparisonLabelToReveal === label} -->
  <!-- {@const foo = (() => {
    //console.log("foo");
    if (label == "Facebook") {
      console.log(
        "Facebook",
        "value",
        value,
        "comparisonValue",
        comparisonValue,
        "formattedValue",
        formattedValue,
        "rowCount",
        rowCount
      );
    }
  })()} -->

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
    <!-- <DimensionLeaderboardEntry
      {renderValue}
      showContext={showComparison}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {formatPreset}
    > -->
    <DimensionLeaderboardEntry
      {renderValue}
      {label}
      measureValue={value}
      showContext={showComparison}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
      {excluded}
      {comparisonValue}
      {showComparisonForThisValue}
      {formatPreset}
      {formattedValue}
      {filterExcludeMode}
      {totalFilteredRowCount}
    >
      <!-- <LeaderboardEntryRightValue
        slot="right"
        {value}
        {comparisonValue}
        {showComparisonForThisValue}
        {formatPreset}
        {formattedValue}
      /> -->
      <!-- <div slot="right" class="flex items-baseline gap-x-1">
        {#if showComparisonForThisValue && comparisonValue !== undefined && comparisonValue !== null}
          <span
            class="inline-block opacity-50"
            transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
          >
            {humanizeDataType(comparisonValue, formatPreset)}
            →
          </span>
        {/if}
        <FormattedDataType type="INTEGER" value={formattedValue || value} />
      </div> -->
      <!-- <span slot="context">
        
      </span> -->
      <!-- <LeaderboardItemTooltip
        slot="tooltip"
        {rowCount}
        {totalFilteredRowCount}
        {excluded}
        filtered={atLeastOneActive}
        {filterExcludeMode}
      /> -->
    </DimensionLeaderboardEntry>
  </div>
{/each}
