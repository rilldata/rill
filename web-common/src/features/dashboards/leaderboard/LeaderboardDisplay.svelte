<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import Leaderboard from "./Leaderboard.svelte";
  import LeaderboardControls from "./LeaderboardControls.svelte";

  const {
    selectors: {
      dimensions: { visibleDimensions },
    },
    metricsViewName,
  } = getStateManagers();

  let parentElement: HTMLDivElement;
</script>

<div class="flex flex-col overflow-hidden size-full">
  <div class="pl-1 pb-3">
    <LeaderboardControls metricViewName={$metricsViewName} />
  </div>
  <div bind:this={parentElement} class="overflow-y-auto leaderboard-display">
    {#if parentElement}
      <div class="leaderboard-grid overflow-hidden pb-4">
        {#each $visibleDimensions as item (item.name)}
          {#if item.name}
            <Leaderboard dimensionName={item.name} {parentElement} />
          {/if}
        {/each}
      </div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .leaderboard-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, 315px);
    gap: 1.5rem;
    @apply h-fit overflow-hidden;
  }
</style>
