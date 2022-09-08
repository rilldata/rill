<section class="grid items-stretch leaderboard-layout bg-white">
  <div class="explore-header">
    <slot name="header" />
  </div>
  <hr class="pb-3 pt-1" />
  <div class="explore-metrics pl-8 pb-8">
    <slot name="metrics" />
  </div>
  <div class="explore-leaderboards pr-8 pb-8">
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    grid-template-rows: auto 1fr;
    grid-template-columns: 560px minmax(355px, auto);
    height: 100vh;
    overflow-x: auto;
    overflow-y: hidden;
    grid-template-areas:
      "header header"
      "hr hr"
      "metrics leaderboards";
  }

  hr {
    grid-area: hr;
  }
  .explore-header {
    grid-area: header;
  }
  .explore-metrics {
    grid-area: metrics;
    overflow-y: auto;
  }
  .explore-leaderboards {
    grid-area: leaderboards;
  }
</style>
