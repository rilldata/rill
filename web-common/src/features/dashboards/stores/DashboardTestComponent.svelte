<script lang="ts">
  import {
    createTimeRangeSummary,
    useMetricsView,
  } from "@rilldata/web-common/features/dashboards/selectors/index";
  import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let ctx: StateManagers;
  const metricsView = useMetricsView(ctx);
  const timeRangeSummaryQuery = createTimeRangeSummary(ctx);
</script>

{!!$metricsView.data}
{!!$timeRangeSummaryQuery.data}
