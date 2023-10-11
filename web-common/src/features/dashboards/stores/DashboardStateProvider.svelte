<script lang="ts">
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    createTimeRangeSummary,
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";

  export let metricViewName: string;

  $: initLocalUserPreferenceStore(metricViewName);
  const stateManagers = getStateManagers();
  const metaQuery = useMetaQuery(stateManagers);
  const hasTimeSeries = useModelHasTimeSeries(stateManagers);
  const timeRangeQuery = createTimeRangeSummary(stateManagers);

  function syncDashboardState() {
    if (metricViewName in $metricsExplorerStore.entities) {
      metricsExplorerStore.sync(metricViewName, $metaQuery.data);
    } else {
      metricsExplorerStore.init(
        metricViewName,
        $metaQuery.data,
        $timeRangeQuery.data
      );
    }
  }

  $: if ($metaQuery.data && ($timeRangeQuery.data || !$hasTimeSeries.data)) {
    syncDashboardState();
  }
</script>

{#if metricViewName in $metricsExplorerStore.entities}
  <slot />
{/if}
