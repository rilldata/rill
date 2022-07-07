<section
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
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    --header: 160px;
    grid-template-rows: var(--header) 1fr;
    grid-template-columns: 532px auto;
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
