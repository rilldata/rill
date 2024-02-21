<script lang="ts">
  import { page } from "$app/stores";
  import {
    createTimeRangeSummary,
    useMetricsView,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";

  export let metricViewName: string;

  $: initLocalUserPreferenceStore(metricViewName);
  const stateManagers = getStateManagers();
  const metricsView = useMetricsView(stateManagers);
  const hasTimeSeries = useModelHasTimeSeries(stateManagers);
  const timeRangeQuery = createTimeRangeSummary(stateManagers);
  const metricsViewSchema = createQueryServiceMetricsViewSchema(
    $runtime.instanceId,
    metricViewName,
  );

  function syncDashboardState() {
    if (!$metricsView.data || !$metricsViewSchema.data?.schema) return;
    if (metricViewName in $metricsExplorerStore.entities) {
      metricsExplorerStore.sync(metricViewName, $metricsView.data);
    } else {
      metricsExplorerStore.init(
        metricViewName,
        $metricsView.data,
        $timeRangeQuery.data,
      );
      const urlState = $page.url.searchParams.get("state");
      if (urlState) {
        metricsExplorerStore.syncFromUrl(
          metricViewName,
          urlState,
          $metricsView.data,
          $metricsViewSchema.data.schema,
        );
        // Call sync to make sure changes in dashboard are honoured
        metricsExplorerStore.sync(metricViewName, $metricsView.data);
      }
    }
  }

  $: if (
    $metricsView.data &&
    $metricsViewSchema.data &&
    ($timeRangeQuery.data || !$hasTimeSeries.data)
  ) {
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
