<script lang="ts">
  import {
    createTimeRangeSummary,
    useMetricsViewValidSpec,
  } from "@rilldata/web-common/features/dashboards/selectors/index";
  import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let ctx: StateManagers;
  const metricsView = useMetricsViewValidSpec(ctx);
  const timeRangeSummaryQuery = createTimeRangeSummary(ctx);
</script>

{!!$metricsView.data}
{!!$timeRangeSummaryQuery.data}
