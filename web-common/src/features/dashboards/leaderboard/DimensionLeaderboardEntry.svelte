<script lang="ts">
  import LeaderboardListItem from "./LeaderboardListItem.svelte";

  // FIXME: there is really no need for both this component and
  // LeaderboardListItem to exist, they do the same job, and this
  // component really just forwards slots and props to LeaderboardListItem.

  /** grays out the value if this is true */
  export let loading = false;

  export let active;
  /** the measure value to be displayed on the right side */
  export let measureValue;

  /** show the context number next to the actual value */
  export let showContext = false;
  /** we'll use special styling when at least one value elsewhere is active */
  export let atLeastOneActive = false;
  /** if this value is a summable measure, we'll show the bar. Otherwise, don't. */
  export let isSummableMeasure;
  /** for summable measures, this is the value we use to calculate the bar % to fill */
  export let referenceValue;

  export let excluded = false;

  export let label: string | number;
  export let previousValueString: string;
  export let formattedValue: string;
  export let filterExcludeMode;
  export let percentChangeFormatted;

  /** if this is a summable measure and there's a reference value, show measureValue / referenceValue.
   * This value is between 0-1 (in theroy!). If it is > 1, the BarAndLabel component shows teeth expressing
   * the value is > 100% of the reference.
   */
</script>

<LeaderboardListItem
  {measureValue}
  {showContext}
  isActive={active}
  {excluded}
  on:click
  {atLeastOneActive}
  {loading}
  {label}
  {previousValueString}
  {formattedValue}
  {percentChangeFormatted}
  {filterExcludeMode}
  {isSummableMeasure}
  {referenceValue}
/>
