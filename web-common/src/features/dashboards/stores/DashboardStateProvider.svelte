<script lang="ts">
  import {
    createTimeRangeSummary,
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";

  export let metricViewName: string;

  $: initLocalUserPreferenceStore(metricViewName);
  const stateManagers = getStateManagers();
  const metaQuery = useMetaQuery(stateManagers);
  const hasTimeSeries = useModelHasTimeSeries(stateManagers);
  const timeRangeQuery = createTimeRangeSummary(stateManagers);

  function syncDashboardState() {
    if (!$metaQuery.data) return;
    if (metricViewName in $metricsExplorerStore.entities) {
      metricsExplorerStore.sync(metricViewName, $metaQuery.data);
    } else {
      metricsExplorerStore.init(
        metricViewName,
        $metaQuery.data,
        $timeRangeQuery.data,
      );
    }
  }

  $: if ($metaQuery.data && ($timeRangeQuery.data || !$hasTimeSeries.data)) {
    syncDashboardState();
  }

  $: ready = metricViewName in $metricsExplorerStore.entities;
</script>

{#if ready}
  <slot />
{:else}
  <div class="grid place-items-center mt-40">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{/if}
