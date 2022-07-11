<section class="grid items-stretch leaderboard-layout bg-white p-8">
  <div class="explore-header">
    <slot name="header" />
  </div>
  <hr />
  <div class="explore-metrics">
    <slot name="metrics" />
  </div>
  <div class="explore-leaderboards">
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    --header: 120px;
    grid-template-rows: var(--header) 1fr;
    grid-template-columns: 540px auto;
    grid-row-gap: 1rem;
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
  }
  .explore-leaderboards {
    grid-area: leaderboards;
  }
</style>
