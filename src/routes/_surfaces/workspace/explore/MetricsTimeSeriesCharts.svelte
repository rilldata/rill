<script>
  import MetricsExploreTimeChart from "$lib/components/leaderboard/MetricsExploreTimeChart.svelte";
  export let metricsDefId;
</script>

<div>
  <div class="grid grid grid-flow-col">
    <div class="big-number">
      <h2>Place for name</h2>
      <div />
    </div>
    <MetricsExploreTimeChart {metricsDefId} />
  </div>
</div>
