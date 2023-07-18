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
  import LeaderboardListItem from "./LeaderboardListItem.svelte";

  export let values;
  $: console.log("values", values);
  export let comparisonValues;
  $: console.log("comparisonValues", comparisonValues);
  export let showTimeComparison = false;
  export let showPercentOfTotal = false;

  export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;
  export let formatPreset;

  let renderValues = [];

  let showContext: "time" | "percent" | false = false;
  $: showContext = showTimeComparison
    ? "time"
    : showPercentOfTotal
    ? "percent"
    : false;

  $: comparisonMap = new Map(comparisonValues?.map((v) => [v.label, v.value]));

  // FIXME: in no world should it be the responsibility of this component to
  // merge `values` and `comparisonValues` and `activeValues`. This should be
  // done somewhere upstream -- ideally, not in a component at all, but given
  // the current architecture, it should at least happen in the parent component.
  $: renderValues = values.map((v) => {
    const active = activeValues.findIndex((value) => value === v.label) >= 0;
    const comparisonValue = comparisonMap.get(v.label);

    return {
      ...v,
      active,
      comparisonValue,
    };
  });
</script>

{#each renderValues as { label, value, active, comparisonValue } (label)}
  <LeaderboardListItem
    measureValue={value}
    {showContext}
    isActive={active}
    {atLeastOneActive}
    {loading}
    {label}
    {comparisonValue}
    {filterExcludeMode}
    {isSummableMeasure}
    {referenceValue}
    {formatPreset}
    on:click
    on:keydown
    on:select-item
  />
{/each}
