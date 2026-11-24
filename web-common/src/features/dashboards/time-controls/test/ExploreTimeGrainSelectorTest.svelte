<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import AggregationGrainSelector from "@rilldata/web-common/features/dashboards/time-controls/AggregationGrainSelector.svelte";

  /**
   * Renders the aggregation grain selector the way the explore chart toolbar does,
   * so that tests can drive the grain without rendering the time series charts.
   */
  const { timeFilterManager } = getStateManagers();
</script>

<AggregationGrainSelector {timeFilterManager} />
