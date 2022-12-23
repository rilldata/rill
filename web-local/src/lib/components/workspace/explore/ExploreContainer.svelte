<script lang="ts">
  import type { MetricsExplorerEntity } from "@rilldata/web-local/lib/application-state-stores/explorer-stores";

  import { metricsExplorerStore } from "../../../application-state-stores/explorer-stores";
  import { hasDefinedTimeSeries } from "./utils";

  export let metricViewName: string;
  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: hasTimeSeries = hasDefinedTimeSeries(metricsExplorer);
</script>

<section
  class="grid items-stretch leaderboard-layout surface"
  style:grid-template-columns="{hasTimeSeries ? "560px" : "240px"} minmax(355px,
  auto)"
>
  <div class="explore-header">
    <slot name="header" />
  </div>
  <hr class="pb-3 pt-1 ui-divider" />
  <div class="explore-metrics pl-8 pb-8">
    <slot name="metrics" />
  </div>
  <div class="explore-leaderboards pr-4 pb-8">
    <slot name="leaderboards" />
  </div>
</section>

<style>
  section {
    grid-template-rows: auto 1fr;
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
