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
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { createEventDispatcher } from "svelte";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";
  import {
    LeaderboardRenderValue,
    valuesToRenderValues,
  } from "./leaderboard-render-values";

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

  let comparisonLabelToReveal = undefined;
  function revealComparisonNumber(value) {
    return () => {
      if (showComparison) comparisonLabelToReveal = value;
    };
  }
</script>

{#each renderValues as renderValue (renderValue.label)}
  <div
    use:shiftClickAction
    on:click={() => {
      dispatch("select-item", {
        label: renderValue.label,
      });
    }}
    on:mouseenter={revealComparisonNumber(renderValue.label)}
    on:mouseleave={revealComparisonNumber(undefined)}
    on:keydown
    on:shift-click={async () => {
      await navigator.clipboard.writeText(renderValue.label);
      let truncatedLabel = renderValue.label?.toString();
      if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
        truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
      }
      notifications.send({
        message: `copied dimension value "${truncatedLabel}" to clipboard`,
      });
    }}
  >
    <DimensionLeaderboardEntry
      {renderValue}
      showContext={showComparison}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {formatPreset}
      {filterExcludeMode}
      {totalFilteredRowCount}
    />
  </div>
{/each}
