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
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let label: string | number;
  export let selected: boolean;
  export let atLeastOneActive;
</script>

<TooltipContent>
  <TooltipTitle>
    <svelte:fragment slot="name">
      {label}
    </svelte:fragment>
  </TooltipTitle>

  {#if !selected && atLeastOneActive}
    <TooltipShortcutContainer>
      <div>Exclusively select this dimension value</div>
      <Shortcut>
        <MetaKey />
      </Shortcut>
    </TooltipShortcutContainer>
  {/if}
</TooltipContent>
