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
  import TableCellTooltip from "@rilldata/web-common/components/table-cell-tooltip/TableCellTooltip.svelte";

  export let label: string | number;
  export let selected: boolean;
  export let excluded: boolean;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let atLeastOneActive;
</script>

<TableCellTooltip
  {label}
  {selected}
  {excluded}
  {filterExcludeMode}
  {atLeastOneActive}
  cellLabel="dimension value"
/>
