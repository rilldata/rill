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
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let label: string | number;
</script>

<TooltipContent>
  <div
    class="grid gap-x-2 pointer-events-none pt-1 pb-1 items-baseline"
    style="grid-template-columns: auto max-content"
    style:min-width="200px"
  >
    <div class="truncate" aria-label="tooltip-name">
      {label}
    </div>
  </div>
</TooltipContent>
