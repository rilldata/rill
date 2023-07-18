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

  // export let values;
  // export let comparisonValues;
  export let showContext: "time" | "percent" | false = false;

  // export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let formatPreset;

  export let renderValues = [];
</script>

{#each renderValues as itemData (itemData.label)}
  <LeaderboardListItem
    {itemData}
    {showContext}
    {atLeastOneActive}
    {filterExcludeMode}
    {isSummableMeasure}
    {referenceValue}
    {formatPreset}
    on:click
    on:keydown
    on:select-item
  />
{/each}
