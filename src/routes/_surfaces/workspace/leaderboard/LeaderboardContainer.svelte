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
  <slot {columns} />
</section>

<style>
  section {
    --header: 160px;
  }
</style>
