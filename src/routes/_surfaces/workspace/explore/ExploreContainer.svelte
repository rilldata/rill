<script lang="ts">
  import { onMount } from "svelte";
  let columns = 3;
  let leaderboardContainer: HTMLElement;
  let availableWidth = 0;
  function onResize() {
    availableWidth = leaderboardContainer.offsetWidth;
    columns = Math.floor(availableWidth / (315 + 20));
  }
  onMount(() => {
    onResize();
  });
</script>

<svelte:window on:resize={onResize} />

<section
  bind:this={leaderboardContainer}
  class="grid items-stretch leaderboard-layout bg-white p-8"
  style:grid-template-rows="var(--header) 1fr"
>
  <div class="explore-header">
    <slot name="header" />
  </div>
  <div class="explore-metrics">
    <slot name="metrics" />
  </div>
  <div class="explore-leaderboards">
    <slot name="leaderboards" {columns} />
  </div>
</section>

<style>
  section {
    --header: 160px;
    grid-template-rows: var(--header) 1fr;
    grid-template-columns: 600px auto;
    grid-template-areas:
      "header header"
      "metrics leaderboards";
  }
  .explore-header {
    grid-area: header;
  }
  .explore-metrics {
    grid-area: metrics;
  }
  .explore-leaderboards {
    grid-area: leaderboards;
  }
</style>
